package letters

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

type StorageClient interface {
	PutObject(context.Context, string, []byte, string) error
}

func (s *Service) IssueLetter(ctx context.Context, principal *auth.Principal, id string, storage StorageClient) (LetterRequestItem, error) {
	if principal == nil {
		return LetterRequestItem{}, ErrForbidden
	}
	if storage == nil {
		return LetterRequestItem{}, ErrStorage
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return LetterRequestItem{}, fmt.Errorf("begin issue letter: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		status         string
		template       string
		numberPattern  string
		formData       json.RawMessage
		residentName   string
		letterTypeName string
	)
	err = tx.QueryRow(ctx, `
		SELECT lr.status, lt.template, lt.number_pattern, lr.form_data, r.full_name, lt.name
		FROM letter_requests lr
		JOIN letter_types lt ON lt.organization_id = lr.organization_id AND lt.id = lr.letter_type_id
		JOIN residents r ON r.organization_id = lr.organization_id AND r.id = lr.resident_id
		WHERE lr.organization_id = $1 AND lr.id = $2
		FOR UPDATE`,
		principal.OrganizationID, id,
	).Scan(&status, &template, &numberPattern, &formData, &residentName, &letterTypeName)
	if errors.Is(err, pgx.ErrNoRows) {
		return LetterRequestItem{}, ErrLetterRequestNotFound
	}
	if err != nil {
		return LetterRequestItem{}, fmt.Errorf("lock letter request: %w", err)
	}
	if status != "approved" {
		return LetterRequestItem{}, ErrInvalidState
	}

	now := s.now()
	letterNumber, err := s.generateLetterNumber(ctx, tx, principal.OrganizationID, numberPattern, now)
	if err != nil {
		return LetterRequestItem{}, err
	}
	pdfData, err := generatePDF(template, letterNumber, residentName, letterTypeName, formData, now)
	if err != nil {
		return LetterRequestItem{}, err
	}

	fileID := newUUID()
	storageKey := fmt.Sprintf("private/%s/issued-letters/%s.pdf", principal.OrganizationID, fileID)
	if err := storage.PutObject(ctx, storageKey, pdfData, "application/pdf"); err != nil {
		return LetterRequestItem{}, fmt.Errorf("%w: upload issued letter", ErrStorage)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO file_objects (
			id, organization_id, storage_key, original_name, mime_type, size_bytes,
			visibility, uploaded_by, confirmed_at
		) VALUES ($1, $2, $3, $4, 'application/pdf', $5, 'private', $6, $7)`,
		fileID, principal.OrganizationID, storageKey,
		"surat-"+strings.ReplaceAll(letterNumber, "/", "-")+".pdf",
		len(pdfData), principal.UserID, now,
	); err != nil {
		return LetterRequestItem{}, fmt.Errorf("insert issued letter file: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE letter_requests
		SET letter_number = $1, status = 'issued', issued_file_id = $2, issued_at = $3
		WHERE organization_id = $4 AND id = $5`,
		letterNumber, fileID, now, principal.OrganizationID, id,
	); err != nil {
		return LetterRequestItem{}, fmt.Errorf("mark letter issued: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO file_attachments (id, organization_id, file_id, entity_type, entity_id, purpose)
		VALUES ($1, $2, $3, 'letter_request', $4, 'issued_letter')`,
		newUUID(), principal.OrganizationID, fileID, id,
	); err != nil {
		return LetterRequestItem{}, fmt.Errorf("attach issued letter: %w", err)
	}
	if err := s.audit(ctx, tx, principal, "letter_request.issue", "letter_request", id); err != nil {
		return LetterRequestItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return LetterRequestItem{}, fmt.Errorf("commit issued letter: %w", err)
	}
	return s.GetLetterRequest(ctx, principal, id)
}

func (s *Service) generateLetterNumber(ctx context.Context, tx pgx.Tx, organizationID, pattern string, at time.Time) (string, error) {
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, organizationID); err != nil {
		return "", fmt.Errorf("lock letter number sequence: %w", err)
	}

	var sequence int
	if err := tx.QueryRow(ctx, `
		SELECT count(*)::int + 1
		FROM letter_requests
		WHERE organization_id = $1
		  AND status = 'issued'
		  AND EXTRACT(YEAR FROM issued_at) = $2`,
		organizationID, at.Year(),
	).Scan(&sequence); err != nil {
		return "", fmt.Errorf("calculate letter number: %w", err)
	}

	number := strings.TrimSpace(pattern)
	if number == "" {
		return "", ErrValidation
	}
	number = strings.ReplaceAll(number, "{NUM}", fmt.Sprintf("%03d", sequence))
	number = strings.ReplaceAll(number, "{YEAR}", fmt.Sprintf("%04d", at.Year()))
	number = strings.ReplaceAll(number, "{MONTH}", fmt.Sprintf("%02d", int(at.Month())))
	return number, nil
}

func generatePDF(template, number, residentName, letterTypeName string, formData json.RawMessage, at time.Time) ([]byte, error) {
	var values map[string]any
	if err := json.Unmarshal(formData, &values); err != nil {
		return nil, ErrValidation
	}

	content := template
	replacements := map[string]string{
		"{LETTER_NUMBER}":    number,
		"{RESIDENT_NAME}":    residentName,
		"{LETTER_TYPE_NAME}": letterTypeName,
		"{DATE}":             at.Format("02-01-2006"),
	}
	for key, value := range values {
		if text, ok := value.(string); ok {
			replacements["{"+strings.ToUpper(key)+"}"] = text
		}
	}
	for token, value := range replacements {
		content = strings.ReplaceAll(content, token, value)
	}

	// ponytail: PDF teks satu halaman; gunakan renderer HTML/CSS bila template formal multi-halaman disetujui.
	lines := wrapPDFText(content, 90)
	var stream strings.Builder
	stream.WriteString("BT\n/F1 11 Tf\n50 790 Td\n14 TL\n")
	for _, line := range lines {
		stream.WriteString("(")
		stream.WriteString(escapePDFText(line))
		stream.WriteString(") Tj\nT*\n")
	}
	stream.WriteString("ET")

	streamData := stream.String()
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(streamData), streamData),
	}

	var pdf bytes.Buffer
	pdf.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects)+1)
	for index, object := range objects {
		offsets[index+1] = pdf.Len()
		fmt.Fprintf(&pdf, "%d 0 obj\n%s\nendobj\n", index+1, object)
	}
	xref := pdf.Len()
	fmt.Fprintf(&pdf, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&pdf, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&pdf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return pdf.Bytes(), nil
}

func wrapPDFText(text string, width int) []string {
	words := strings.Fields(strings.ReplaceAll(text, "\n", " "))
	lines := make([]string, 0, len(words))
	line := ""
	for _, word := range words {
		if len(line) > 0 && len(line)+len(word)+1 > width {
			lines = append(lines, line)
			line = word
			continue
		}
		if line != "" {
			line += " "
		}
		line += word
	}
	if line != "" {
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return []string{" "}
	}
	return lines
}

func escapePDFText(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "(", `\(`)
	return strings.ReplaceAll(value, ")", `\)`)
}