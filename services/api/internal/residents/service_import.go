package residents

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

type ImportResult struct {
	TotalRows     int      `json:"total_rows"`
	ValidRows     int      `json:"valid_rows"`
	InvalidRows   int      `json:"invalid_rows"`
	DuplicateRows int      `json:"duplicate_rows"`
	ImportedRows  int      `json:"imported_rows"`
	Errors        []string `json:"errors"`
}

type importResidentRow struct {
	line int
	req  CreateResidentRequest
}

func (s *Service) ImportCSV(ctx context.Context, principal *auth.Principal, source io.Reader, dryRun bool) (ImportResult, error) {
	reader := csv.NewReader(source)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return ImportResult{}, fmt.Errorf("%w: invalid CSV", ErrValidation)
	}
	if len(records) < 2 {
		return ImportResult{}, fmt.Errorf("%w: CSV must contain a header and at least one row", ErrValidation)
	}
	if !validImportHeader(records[0]) {
		return ImportResult{}, fmt.Errorf("%w: required headers are full_name,resident_status", ErrValidation)
	}

	result := ImportResult{TotalRows: len(records) - 1}
	rows := make([]importResidentRow, 0, result.TotalRows)
	seenNIK := map[string]int{}

	for index, record := range records[1:] {
		line := index + 2
		if len(record) < 2 {
			result.InvalidRows++
			result.Errors = append(result.Errors, fmt.Sprintf("Baris %d: kolom tidak lengkap.", line))
			continue
		}
		req := CreateResidentRequest{
			FullName:       strings.TrimSpace(record[0]),
			ResidentStatus: strings.TrimSpace(record[1]),
		}
		if len(record) > 2 {
			req.NationalID = nullableTrim(&record[2])
		}
		if len(record) > 3 {
			req.Phone = nullableTrim(&record[3])
		}
		if req.FullName == "" || !validResidentStatus(req.ResidentStatus) {
			result.InvalidRows++
			result.Errors = append(result.Errors, fmt.Sprintf("Baris %d: nama atau status warga tidak valid.", line))
			continue
		}
		if req.NationalID != nil {
			index := auth.GenerateBlindIndex(*req.NationalID, s.blindKey)
			if previous, duplicate := seenNIK[index]; duplicate {
				result.DuplicateRows++
				result.Errors = append(result.Errors, fmt.Sprintf("Baris %d: NIK duplikat dengan baris %d.", line, previous))
				continue
			}
			seenNIK[index] = line
		}
		rows = append(rows, importResidentRow{line: line, req: req})
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return result, fmt.Errorf("begin resident import: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, row := range rows {
		if row.req.NationalID == nil {
			continue
		}
		blindIndex := auth.GenerateBlindIndex(*row.req.NationalID, s.blindKey)
		var exists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM residents
				WHERE organization_id = $1 AND national_id_blind_index = $2
			)`, principal.OrganizationID, blindIndex).Scan(&exists); err != nil {
			return result, fmt.Errorf("check resident duplicate: %w", err)
		}
		if exists {
			result.DuplicateRows++
			result.Errors = append(result.Errors, fmt.Sprintf("Baris %d: NIK sudah terdaftar.", row.line))
		}
	}

	result.ValidRows = result.TotalRows - result.InvalidRows - result.DuplicateRows
	if dryRun || result.InvalidRows > 0 || result.DuplicateRows > 0 {
		return result, nil
	}

	for _, row := range rows {
		encryptedID, blindIndex, err := s.encryptIndexedValue(row.req.NationalID)
		if err != nil {
			return result, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO residents (
				id, organization_id, national_id_encrypted, national_id_blind_index,
				full_name, phone, resident_status, verification_status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, 'verified')`,
			newUUID(), principal.OrganizationID, encryptedID, blindIndex,
			row.req.FullName, row.req.Phone, row.req.ResidentStatus,
		); err != nil {
			return result, mapDatabaseError(err, "import resident")
		}
		result.ImportedRows++
	}
	if err := s.audit(ctx, tx, principal, "resident.import", "organizations", principal.OrganizationID); err != nil {
		return result, err
	}
	if err := tx.Commit(ctx); err != nil {
		return result, fmt.Errorf("commit resident import: %w", err)
	}
	return result, nil
}

func validImportHeader(header []string) bool {
	return len(header) >= 2 &&
		strings.EqualFold(strings.TrimSpace(header[0]), "full_name") &&
		strings.EqualFold(strings.TrimSpace(header[1]), "resident_status")
}
