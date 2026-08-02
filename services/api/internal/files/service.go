package files

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
)

const uploadLifetime = 15 * time.Minute
const downloadLifetime = 5 * time.Minute

type Storage interface {
	PresignUpload(context.Context, string, string, time.Duration) (platform.PresignedURL, error)
	HeadObject(context.Context, string) (platform.ObjectMetadata, error)
	PresignDownload(context.Context, string, time.Duration) (string, error)
}

type Service struct {
	db      *pgxpool.Pool
	storage Storage
	now     func() time.Time
}

func NewService(db *pgxpool.Pool, storage Storage) *Service {
	return &Service{
		db:      db,
		storage: storage,
		now:     func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) PresignUpload(ctx context.Context, principal *auth.Principal, request PresignUploadRequest) (PresignUploadResponse, error) {
	if principal == nil || request.Validate() != nil {
		return PresignUploadResponse{}, ErrValidation
	}

	fileID := newUUID()
	storageKey := fmt.Sprintf("private/%s/payment-proof/%s", principal.OrganizationID, fileID)
	presigned, err := s.storage.PresignUpload(ctx, storageKey, request.MIMEType, uploadLifetime)
	if err != nil {
		return PresignUploadResponse{}, fmt.Errorf("%w: presign upload", ErrStorage)
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO file_objects (
			id, organization_id, storage_key, original_name, mime_type, size_bytes, visibility, uploaded_by
		) VALUES ($1, $2, $3, $4, $5, $6, 'private', $7)`,
		fileID,
		principal.OrganizationID,
		storageKey,
		request.OriginalName,
		request.MIMEType,
		request.SizeBytes,
		principal.UserID,
	)
	if err != nil {
		return PresignUploadResponse{}, fmt.Errorf("insert file object: %w", err)
	}

	return PresignUploadResponse{
		FileID:        fileID,
		UploadURL:     presigned.URL,
		UploadHeaders: presigned.Headers,
		ExpiresAt:     s.now().Add(uploadLifetime),
	}, nil
}

func (s *Service) ConfirmUpload(ctx context.Context, principal *auth.Principal, request ConfirmUploadRequest) (FileResponse, error) {
	if principal == nil || request.Validate() != nil {
		return FileResponse{}, ErrValidation
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return FileResponse{}, fmt.Errorf("begin file confirmation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		storageKey   string
		originalName string
		mimeType     string
		sizeBytes    int64
		confirmedAt  *time.Time
		createdAt    time.Time
	)
	err = tx.QueryRow(ctx, `
		SELECT storage_key, original_name, mime_type, size_bytes, confirmed_at, created_at
		FROM file_objects
		WHERE id = $1
		  AND organization_id = $2
		  AND uploaded_by = $3
		  AND deleted_at IS NULL
		FOR UPDATE`,
		request.FileID,
		principal.OrganizationID,
		principal.UserID,
	).Scan(&storageKey, &originalName, &mimeType, &sizeBytes, &confirmedAt, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return FileResponse{}, ErrFileNotFound
	}
	if err != nil {
		return FileResponse{}, fmt.Errorf("lock file object: %w", err)
	}

	if confirmedAt != nil {
		return FileResponse{
			ID:           request.FileID,
			OriginalName: originalName,
			MIMEType:     mimeType,
			SizeBytes:    sizeBytes,
			ConfirmedAt:  confirmedAt,
			CreatedAt:    createdAt,
		}, tx.Commit(ctx)
	}

	metadata, err := s.storage.HeadObject(ctx, storageKey)
	if err != nil {
		return FileResponse{}, fmt.Errorf("%w: uploaded object not found", ErrFileUnavailable)
	}
	if metadata.SizeBytes != sizeBytes || metadata.ContentType != mimeType {
		return FileResponse{}, ErrValidation
	}

	now := s.now()
	if _, err := tx.Exec(ctx, `
		UPDATE file_objects
		SET confirmed_at = $1
		WHERE id = $2 AND organization_id = $3`,
		now,
		request.FileID,
		principal.OrganizationID,
	); err != nil {
		return FileResponse{}, fmt.Errorf("confirm file object: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id)
		VALUES ($1, $2, 'file.upload_confirm', 'file_object', $3)`,
		principal.OrganizationID,
		principal.UserID,
		request.FileID,
	); err != nil {
		return FileResponse{}, fmt.Errorf("audit file confirmation: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return FileResponse{}, fmt.Errorf("commit file confirmation: %w", err)
	}

	return FileResponse{
		ID:           request.FileID,
		OriginalName: originalName,
		MIMEType:     mimeType,
		SizeBytes:    sizeBytes,
		ConfirmedAt:  &now,
		CreatedAt:    createdAt,
	}, nil
}

func (s *Service) PresignDownload(ctx context.Context, principal *auth.Principal, fileID string) (PresignDownloadResponse, error) {
	if principal == nil || fileID == "" {
		return PresignDownloadResponse{}, ErrValidation
	}

	var storageKey string
	err := s.db.QueryRow(ctx, `
		SELECT f.storage_key
		FROM file_objects f
		JOIN file_attachments a
		  ON a.organization_id = f.organization_id
		 AND a.file_id = f.id
		 AND a.entity_type = 'payment'
		 AND a.purpose = 'payment_proof'
		JOIN payments p
		  ON p.organization_id = a.organization_id
		 AND p.id = a.entity_id
		WHERE f.id = $1
		  AND f.organization_id = $2
		  AND f.deleted_at IS NULL
		  AND f.confirmed_at IS NOT NULL
		  AND (p.created_by = $3 OR $4)`,
		fileID,
		principal.OrganizationID,
		principal.UserID,
		principal.HasPermission("payment.read"),
	).Scan(&storageKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return PresignDownloadResponse{}, ErrFileNotFound
	}
	if err != nil {
		return PresignDownloadResponse{}, fmt.Errorf("authorize file download: %w", err)
	}

	url, err := s.storage.PresignDownload(ctx, storageKey, downloadLifetime)
	if err != nil {
		return PresignDownloadResponse{}, fmt.Errorf("%w: presign download", ErrStorage)
	}

	return PresignDownloadResponse{
		FileID:      fileID,
		DownloadURL: url,
		ExpiresAt:   s.now().Add(downloadLifetime),
	}, nil
}

func newUUID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic("secure random source unavailable")
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}
