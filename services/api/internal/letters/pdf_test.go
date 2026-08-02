package letters

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestGeneratePDF(t *testing.T) {
	pdf, err := generatePDF(
		"Surat {LETTER_NUMBER} untuk {RESIDENT_NAME}: {KEPERLUAN}.",
		"470/001/2026",
		"Budi",
		"Surat Keterangan",
		json.RawMessage(`{"keperluan":"Melamar kerja"}`),
		time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("generatePDF() error = %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4\n")) {
		t.Fatalf("PDF header missing: %q", pdf[:min(len(pdf), 16)])
	}
	for _, value := range []string{"470/001/2026", "Budi", "Melamar kerja"} {
		if !bytes.Contains(pdf, []byte(value)) {
			t.Errorf("PDF does not contain %q", value)
		}
	}
}

func TestValidateForm(t *testing.T) {
	if err := validateForm(
		json.RawMessage(`{"keperluan":"Domisili"}`),
		json.RawMessage(`[{"required":true}]`),
		0,
	); err != ErrValidation {
		t.Fatalf("validateForm() error = %v, want ErrValidation", err)
	}

	if err := validateForm(
		json.RawMessage(`{"keperluan":"Domisili"}`),
		json.RawMessage(`[{"required":true}]`),
		1,
	); err != nil {
		t.Fatalf("validateForm() error = %v", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}