package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

var ErrNotFound = errors.New("organization settings not found")

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) Get(ctx context.Context, organizationID string) (OrganizationSettings, error) {
	item, err := get(ctx, s.db, organizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return OrganizationSettings{}, ErrNotFound
	}
	if err != nil {
		return OrganizationSettings{}, fmt.Errorf("get organization settings: %w", err)
	}
	return item, nil
}

func (s *Service) Update(ctx context.Context, principal *auth.Principal, req UpdateOrganizationSettingsRequest, requestID string) (OrganizationSettings, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return OrganizationSettings{}, fmt.Errorf("begin settings update: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	current, err := getForUpdate(ctx, tx, principal.OrganizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return OrganizationSettings{}, ErrNotFound
	}
	if err != nil {
		return OrganizationSettings{}, fmt.Errorf("lock organization settings: %w", err)
	}
	before := auditSnapshot(current)

	if err := apply(&current, req); err != nil {
		return OrganizationSettings{}, err
	}

	updated, err := update(ctx, tx, current)
	if err != nil {
		return OrganizationSettings{}, fmt.Errorf("update organization settings: %w", err)
	}

	roles, err := json.Marshal(principal.RoleCodes)
	if err != nil {
		return OrganizationSettings{}, fmt.Errorf("encode active roles: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (
			organization_id, actor_user_id, actor_role_codes, action, entity_type, entity_id,
			before_data, after_data, request_id
		)
		VALUES ($1, $2, $3, 'organization.update', 'organizations', $4, $5, $6, $7)`,
		principal.OrganizationID, principal.UserID, roles, principal.OrganizationID,
		before, auditSnapshot(updated), nullableTrim(requestID),
	); err != nil {
		return OrganizationSettings{}, fmt.Errorf("write settings audit: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return OrganizationSettings{}, fmt.Errorf("commit settings update: %w", err)
	}
	return updated, nil
}

type queryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func get(ctx context.Context, db queryer, organizationID string) (OrganizationSettings, error) {
	var item OrganizationSettings
	err := db.QueryRow(ctx, `
		SELECT id, name, rt_number, rw_number, address, timezone, logo_file_id,
		       bank_name, bank_account_number, bank_account_holder, max_upload_size_bytes,
		       default_letter_number_pattern, status, settings, created_at, updated_at
		FROM organizations
		WHERE id = $1`, organizationID,
	).Scan(
		&item.ID, &item.Name, &item.RTNumber, &item.RWNumber, &item.Address, &item.Timezone, &item.LogoFileID,
		&item.BankName, &item.BankAccountNumber, &item.BankAccountHolder, &item.MaxUploadSizeBytes,
		&item.DefaultLetterNumberPattern, &item.Status, &item.Settings, &item.CreatedAt, &item.UpdatedAt,
	)
	return item, err
}

func getForUpdate(ctx context.Context, tx pgx.Tx, organizationID string) (OrganizationSettings, error) {
	var item OrganizationSettings
	err := tx.QueryRow(ctx, `
		SELECT id, name, rt_number, rw_number, address, timezone, logo_file_id,
		       bank_name, bank_account_number, bank_account_holder, max_upload_size_bytes,
		       default_letter_number_pattern, status, settings, created_at, updated_at
		FROM organizations
		WHERE id = $1 FOR UPDATE`, organizationID,
	).Scan(
		&item.ID, &item.Name, &item.RTNumber, &item.RWNumber, &item.Address, &item.Timezone, &item.LogoFileID,
		&item.BankName, &item.BankAccountNumber, &item.BankAccountHolder, &item.MaxUploadSizeBytes,
		&item.DefaultLetterNumberPattern, &item.Status, &item.Settings, &item.CreatedAt, &item.UpdatedAt,
	)
	return item, err
}

func update(ctx context.Context, tx pgx.Tx, item OrganizationSettings) (OrganizationSettings, error) {
	err := tx.QueryRow(ctx, `
		UPDATE organizations
		SET name = $1, rt_number = $2, rw_number = $3, address = $4, timezone = $5, logo_file_id = $6,
		    bank_name = $7, bank_account_number = $8, bank_account_holder = $9, max_upload_size_bytes = $10,
		    default_letter_number_pattern = $11, settings = $12
		WHERE id = $13
		RETURNING updated_at`,
		item.Name, item.RTNumber, item.RWNumber, item.Address, item.Timezone, item.LogoFileID,
		item.BankName, item.BankAccountNumber, item.BankAccountHolder, item.MaxUploadSizeBytes,
		item.DefaultLetterNumberPattern, item.Settings, item.ID,
	).Scan(&item.UpdatedAt)
	return item, err
}

func apply(item *OrganizationSettings, req UpdateOrganizationSettingsRequest) error {
	if req.Name != nil {
		item.Name = strings.TrimSpace(*req.Name)
	}
	if req.RTNumber != nil {
		item.RTNumber = strings.TrimSpace(*req.RTNumber)
	}
	if req.RWNumber != nil {
		item.RWNumber = strings.TrimSpace(*req.RWNumber)
	}
	if item.Name == "" || item.RTNumber == "" || item.RWNumber == "" {
		return errors.New("organization name, RT number, and RW number are required")
	}
	if req.Address != nil {
		item.Address = nullableTrim(*req.Address)
	}
	if req.Timezone != nil {
		location, err := time.LoadLocation(strings.TrimSpace(*req.Timezone))
		if err != nil {
			return errors.New("timezone is invalid")
		}
		item.Timezone = location.String()
	}
	if req.LogoFileID != nil {
		item.LogoFileID = nullableTrim(*req.LogoFileID)
	}
	if req.BankName != nil {
		item.BankName = nullableTrim(*req.BankName)
	}
	if req.BankAccountNumber != nil {
		item.BankAccountNumber = nullableTrim(*req.BankAccountNumber)
	}
	if req.BankAccountHolder != nil {
		item.BankAccountHolder = nullableTrim(*req.BankAccountHolder)
	}
	if req.MaxUploadSizeBytes != nil {
		if *req.MaxUploadSizeBytes < 1 || *req.MaxUploadSizeBytes > 52_428_800 {
			return errors.New("max_upload_size_bytes must be between 1 and 52428800")
		}
		item.MaxUploadSizeBytes = *req.MaxUploadSizeBytes
	}
	if req.DefaultLetterNumberPattern != nil {
		item.DefaultLetterNumberPattern = strings.TrimSpace(*req.DefaultLetterNumberPattern)
		if item.DefaultLetterNumberPattern == "" {
			return errors.New("default_letter_number_pattern is required")
		}
	}
	if req.Settings != nil {
		if !json.Valid(*req.Settings) {
			return errors.New("settings must be valid JSON")
		}
		item.Settings = *req.Settings
	}
	return nil
}

func auditSnapshot(item OrganizationSettings) []byte {
	snapshot, _ := json.Marshal(struct {
		Name                       string `json:"name"`
		RTNumber                   string `json:"rt_number"`
		RWNumber                   string `json:"rw_number"`
		Address                    string `json:"address,omitempty"`
		Timezone                   string `json:"timezone"`
		LogoFileID                 string `json:"logo_file_id,omitempty"`
		BankConfigured             bool   `json:"bank_configured"`
		MaxUploadSizeBytes         int64  `json:"max_upload_size_bytes"`
		DefaultLetterNumberPattern string `json:"default_letter_number_pattern"`
	}{
		Name: item.Name, RTNumber: item.RTNumber, RWNumber: item.RWNumber,
		Address: value(item.Address), Timezone: item.Timezone, LogoFileID: value(item.LogoFileID),
		BankConfigured: item.BankName != nil || item.BankAccountNumber != nil || item.BankAccountHolder != nil,
		MaxUploadSizeBytes: item.MaxUploadSizeBytes, DefaultLetterNumberPattern: item.DefaultLetterNumberPattern,
	})
	return snapshot
}

func nullableTrim(raw string) *string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}
	return &value
}

func value(raw *string) string {
	if raw == nil {
		return ""
	}
	return *raw
}