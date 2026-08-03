package reports

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

type PDFReport struct {
	Title       string
	GeneratedAt time.Time
	Headers     []string
	Rows        [][]string
}

// GenerateFormalPDF creates a printable, text-only report PDF.
// ponytail: fixed-width, ASCII-compatible layout; replace with an HTML/CSS renderer when branded, Unicode, or complex table layouts are approved.
func GenerateFormalPDF(report PDFReport) ([]byte, error) {
	if strings.TrimSpace(report.Title) == "" || len(report.Headers) == 0 {
		return nil, ErrValidation
	}
	if report.GeneratedAt.IsZero() {
		report.GeneratedAt = time.Now().UTC()
	}

	lines := []string{
		strings.ToUpper(pdfSafeText(report.Title)),
		fmt.Sprintf("Dicetak: %s UTC", report.GeneratedAt.UTC().Format("02-01-2006 15:04:05")),
		fmt.Sprintf("Total data: %d", len(report.Rows)),
		strings.Repeat("=", 88),
		strings.Join(pdfCells(report.Headers), " | "),
		strings.Repeat("-", 88),
	}
	if len(report.Rows) == 0 {
		lines = append(lines, "Tidak ada data untuk kriteria filter ini.")
	} else {
		for index, row := range report.Rows {
			lines = append(lines, wrapPDFText(fmt.Sprintf("%d. %s", index+1, strings.Join(pdfCells(row), " | ")), 88)...)
		}
	}
	return buildPDF(lines)
}

func pdfCells(cells []string) []string {
	result := make([]string, len(cells))
	for index, cell := range cells {
		result[index] = pdfSafeText(cell)
	}
	return result
}

func pdfSafeText(value string) string {
	value = strings.Map(func(r rune) rune {
		if r < 32 || r > 126 {
			return '?'
		}
		return r
	}, value)
	return strings.TrimSpace(value)
}

func wrapPDFText(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
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
	return lines
}

func buildPDF(lines []string) ([]byte, error) {
	const linesPerPage = 52
	pageObjects := make([]string, 0, (len(lines)+linesPerPage-1)/linesPerPage)
	for start := 0; start < len(lines); start += linesPerPage {
		end := min(start+linesPerPage, len(lines))
		var stream strings.Builder
		stream.WriteString("BT\n/F1 9 Tf\n40 800 Td\n14 TL\n")
		for _, line := range lines[start:end] {
			stream.WriteString("(")
			stream.WriteString(escapePDFText(line))
			stream.WriteString(") Tj\nT*\n")
		}
		stream.WriteString("ET")
		pageObjects = append(pageObjects, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream.String()), stream.String()))
	}
	if len(pageObjects) == 0 {
		pageObjects = append(pageObjects, "<< /Length 0 >>\nstream\n\nendstream")
	}

	pageIDs := make([]int, len(pageObjects))
	for index := range pageObjects {
		pageIDs[index] = 5 + index*2
	}
	kids := make([]string, len(pageIDs))
	for index, id := range pageIDs {
		kids[index] = fmt.Sprintf("%d 0 R", id)
	}
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), len(pageIDs)),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}
	for index, content := range pageObjects {
		contentID := 4 + index*2
		pageID := contentID + 1
		objects = append(objects,
			content,
			fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 3 0 R >> >> /Contents %d 0 R >>", contentID),
		)
		_ = pageID
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

func escapePDFText(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "(", `\(`)
	return strings.ReplaceAll(value, ")", `\)`)
}

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}