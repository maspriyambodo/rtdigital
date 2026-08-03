package reports

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestGenerateFormalPDF(t *testing.T) {
	report := PDFReport{
		Title:       "Cash Report",
		GeneratedAt: time.Date(2026, 3, 8, 10, 0, 0, 0, time.UTC),
		Headers:     []string{"Category", "Amount"},
		Rows:        [][]string{{"Kebersihan", "150000.00"}},
	}
	pdf, err := GenerateFormalPDF(report)
	if err != nil {
		t.Fatalf("GenerateFormalPDF() error = %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4\n")) {
		t.Fatalf("PDF header missing: %q", pdf[:min(len(pdf), 16)])
	}
	for _, want := range []string{
		"CASH REPORT",
		"Dicetak: 08-03-2026 10:00:00 UTC",
		"Total data: 1",
		"Category | Amount",
		"Kebersihan | 150000.00",
	} {
		if !bytes.Contains(pdf, []byte(want)) {
			t.Errorf("PDF missing %q", want)
		}
	}
}

func TestGenerateFormalPDFValidationAndMultipage(t *testing.T) {
	if _, err := GenerateFormalPDF(PDFReport{}); err == nil {
		t.Fatal("empty report accepted")
	}
	if _, err := GenerateFormalPDF(PDFReport{Title: "Report"}); err == nil {
		t.Fatal("report without headers accepted")
	}

	rows := make([][]string, 120)
	for index := range rows {
		rows[index] = []string{"Resident", "active"}
	}
	pdf, err := GenerateFormalPDF(PDFReport{
		Title:   "Long Report",
		Headers: []string{"Name", "Status"},
		Rows:    rows,
	})
	if err != nil {
		t.Fatalf("GenerateFormalPDF() error = %v", err)
	}
	if pages := strings.Count(string(pdf), "/Type /Page "); pages < 3 {
		t.Errorf("pages = %d, want at least 3", pages)
	}
}