package files

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrValidation      = errors.New("validation failed")
	ErrFileNotFound    = errors.New("file not found")
	ErrFileUnavailable = errors.New("file unavailable")
	ErrStorage         = errors.New("storage operation failed")
)

const (
	maxUploadSize = 10 * 1024 * 1024
)

type PresignUploadRequest struct {
	EntityType   string `json:"entity_type"`
	EntityID     string `json:"entity_id"`
	Purpose      string `json:"purpose"`
	OriginalName string `json:"original_name"`
	MIMEType     string `json:"mime_type"`
	SizeBytes    int64  `json:"size_bytes"`
}

func (r PresignUploadRequest) Validate() error {
	switch r.EntityType {
	case "payment":
		if r.Purpose != "payment_proof" {
			return ErrValidation
		}
	case "cash_transaction":
		if r.Purpose != "proof" {
			return ErrValidation
		}
	default:
		return ErrValidation
	}
	if strings.TrimSpace(r.EntityID) == "" ||
		strings.TrimSpace(r.OriginalName) == "" ||
		!allowedMIMEType(r.MIMEType) ||
		r.SizeBytes <= 0 ||
		r.SizeBytes > maxUploadSize {
		return ErrValidation
	}
	return nil
}

type PresignUploadResponse struct {
	FileID        string            `json:"file_id"`
	UploadURL     string            `json:"upload_url"`
	UploadHeaders map[string]string `json:"upload_headers"`
	ExpiresAt     time.Time         `json:"expires_at"`
}

type ConfirmUploadRequest struct {
	FileID string `json:"file_id"`
}

func (r ConfirmUploadRequest) Validate() error {
	if strings.TrimSpace(r.FileID) == "" {
		return ErrValidation
	}
	return nil
}

type FileResponse struct {
	ID           string     `json:"id"`
	OriginalName string     `json:"original_name"`
	MIMEType     string     `json:"mime_type"`
	SizeBytes    int64      `json:"size_bytes"`
	ConfirmedAt  *time.Time `json:"confirmed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type PresignDownloadResponse struct {
	FileID      string    `json:"file_id"`
	DownloadURL string    `json:"download_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func allowedMIMEType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "image/jpeg", "image/png", "application/pdf":
		return true
	}
	return false
}
