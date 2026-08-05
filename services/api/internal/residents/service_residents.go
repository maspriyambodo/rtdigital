package residents

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

func (s *Service) CreateResident(ctx context.Context, principal *auth.Principal, req CreateResidentRequest) (Resident, error) {
	req.FullName = strings.TrimSpace(req.FullName)
	if req.FullName == "" || !validResidentStatus(req.ResidentStatus) {
		return Resident{}, fmt.Errorf("%w: full_name or resident_status", ErrValidation)
	}

	encryptedID, blindIndex, err := s.encryptIndexedValue(req.NationalID)
	if err != nil {
		return Resident{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Resident{}, fmt.Errorf("begin resident create: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := validateResidentLookups(ctx, tx, req.EducationLevelID, req.MaritalStatusID); err != nil {
		return Resident{}, err
	}

	var resident Resident
	err = tx.QueryRow(ctx, `
		INSERT INTO residents (
			id, organization_id, national_id_encrypted, national_id_blind_index,
			full_name, birth_place, birth_date, gender, marital_status_id, occupation,
			education_level_id, phone, email, resident_status, verification_status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7::date, $8, $9, $10, $11, $12, $13, $14, 'unverified'
		)
		RETURNING id, full_name, birth_place, birth_date::text, gender,
			marital_status_id, occupation, education_level_id, phone, email,
			resident_status, verification_status, created_at, updated_at`,
		newUUID(), principal.OrganizationID, encryptedID, blindIndex, req.FullName,
		nullableTrim(req.BirthPlace), nullableTrim(req.BirthDate), nullableTrim(req.Gender),
		nullableTrim(req.MaritalStatusID), nullableTrim(req.Occupation), nullableTrim(req.EducationLevelID),
		nullableTrim(req.Phone), nullableTrim(req.Email), req.ResidentStatus,
	).Scan(
		&resident.ID, &resident.FullName, &resident.BirthPlace, &resident.BirthDate,
		&resident.Gender, &resident.MaritalStatusID, &resident.Occupation, &resident.EducationLevelID,
		&resident.Phone, &resident.Email, &resident.ResidentStatus, &resident.VerificationStatus,
		&resident.CreatedAt, &resident.UpdatedAt,
	)
	if err != nil {
		return Resident{}, mapDatabaseError(err, "create resident")
	}
	if err := s.audit(ctx, tx, principal, "resident.create", "residents", resident.ID); err != nil {
		return Resident{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Resident{}, fmt.Errorf("commit resident create: %w", err)
	}
	return resident, nil
}

func (s *Service) ListResidents(ctx context.Context, principal *auth.Principal, query, status string) ([]Resident, error) {
	query, status = strings.TrimSpace(query), strings.TrimSpace(status)
	if status != "" && !validResidentStatus(status) {
		return nil, fmt.Errorf("%w: resident_status", ErrValidation)
	}

	rows, err := s.db.Query(ctx, `
		SELECT r.id, r.full_name, r.birth_place, r.birth_date::text, r.gender,
			r.marital_status_id, ms.name, r.occupation, r.education_level_id, el.name,
			r.phone, r.email, r.resident_status, r.verification_status, r.created_at, r.updated_at
		FROM residents r
		LEFT JOIN marital_statuses ms ON ms.id = r.marital_status_id
		LEFT JOIN education_levels el ON el.id = r.education_level_id
		WHERE r.organization_id = $1
		  AND ($2 = '' OR r.resident_status = $2)
		  AND ($3 = '' OR r.full_name ILIKE '%' || $3 || '%')
		ORDER BY r.full_name, r.id`,
		principal.OrganizationID, status, query,
	)
	if err != nil {
		return nil, fmt.Errorf("list residents: %w", err)
	}
	defer rows.Close()

	residents := []Resident{}
	for rows.Next() {
		var resident Resident
		if err := rows.Scan(
			&resident.ID, &resident.FullName, &resident.BirthPlace, &resident.BirthDate,
			&resident.Gender, &resident.MaritalStatusID, &resident.MaritalStatusName, &resident.Occupation,
			&resident.EducationLevelID, &resident.EducationLevelName, &resident.Phone, &resident.Email,
			&resident.ResidentStatus, &resident.VerificationStatus, &resident.CreatedAt, &resident.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan resident: %w", err)
		}
		residents = append(residents, resident)
	}
	return residents, rows.Err()
}

func (s *Service) GetResident(ctx context.Context, principal *auth.Principal, id, sensitiveReason string) (Resident, error) {
	var resident Resident
	var encryptedID *string
	err := s.db.QueryRow(ctx, `
		SELECT r.id, r.national_id_encrypted, r.full_name, r.birth_place, r.birth_date::text, r.gender,
			r.marital_status_id, ms.name, r.occupation, r.education_level_id, el.name,
			r.phone, r.email, r.resident_status, r.verification_status, r.created_at, r.updated_at
		FROM residents r
		LEFT JOIN marital_statuses ms ON ms.id = r.marital_status_id
		LEFT JOIN education_levels el ON el.id = r.education_level_id
		WHERE r.id = $1 AND r.organization_id = $2`,
		id, principal.OrganizationID,
	).Scan(
		&resident.ID, &encryptedID, &resident.FullName, &resident.BirthPlace, &resident.BirthDate,
		&resident.Gender, &resident.MaritalStatusID, &resident.MaritalStatusName, &resident.Occupation,
		&resident.EducationLevelID, &resident.EducationLevelName, &resident.Phone, &resident.Email,
		&resident.ResidentStatus, &resident.VerificationStatus, &resident.CreatedAt, &resident.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Resident{}, ErrResidentNotFound
	}
	if err != nil {
		return Resident{}, fmt.Errorf("get resident: %w", err)
	}

	if encryptedID == nil {
		return resident, nil
	}
	if strings.TrimSpace(sensitiveReason) == "" || !principal.HasPermission("resident.read_sensitive") {
		masked := maskNationalID("")
		resident.NationalID = &masked
		return resident, nil
	}

	nationalID, err := s.crypter.Decrypt(*encryptedID)
	if err != nil {
		return Resident{}, fmt.Errorf("decrypt national id: %w", err)
	}
	resident.NationalID = &nationalID
	if _, err := s.db.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id, metadata)
		VALUES ($1, $2, 'resident.read_sensitive', 'residents', $3, jsonb_build_object('reason', $4::text))`,
		principal.OrganizationID, principal.UserID, resident.ID, strings.TrimSpace(sensitiveReason),
	); err != nil {
		return Resident{}, fmt.Errorf("audit sensitive resident read: %w", err)
	}
	return resident, nil
}

func (s *Service) VerifyResident(ctx context.Context, principal *auth.Principal, id string) (Resident, error) {
	var resident Resident
	err := s.db.QueryRow(ctx, `
		UPDATE residents SET verification_status = 'verified'
		WHERE id = $1 AND organization_id = $2
		RETURNING id, full_name, birth_place, birth_date::text, gender,
			marital_status_id, occupation, education_level_id, phone, email,
			resident_status, verification_status, created_at, updated_at`,
		id, principal.OrganizationID,
	).Scan(
		&resident.ID, &resident.FullName, &resident.BirthPlace, &resident.BirthDate,
		&resident.Gender, &resident.MaritalStatusID, &resident.Occupation, &resident.EducationLevelID,
		&resident.Phone, &resident.Email, &resident.ResidentStatus, &resident.VerificationStatus,
		&resident.CreatedAt, &resident.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Resident{}, ErrResidentNotFound
	}
	if err != nil {
		return Resident{}, fmt.Errorf("verify resident: %w", err)
	}
	if _, err := s.db.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id)
		VALUES ($1, $2, 'resident.verify', 'residents', $3)`,
		principal.OrganizationID, principal.UserID, id,
	); err != nil {
		return Resident{}, fmt.Errorf("audit resident verification: %w", err)
	}
	return resident, nil
}

func validateResidentLookups(ctx context.Context, tx pgx.Tx, educationLevelID, maritalStatusID *string) error {
	for _, lookup := range []struct {
		table string
		id    *string
	}{
		{"education_levels", nullableTrim(educationLevelID)},
		{"marital_statuses", nullableTrim(maritalStatusID)},
	} {
		if lookup.id == nil {
			continue
		}
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM `+lookup.table+` WHERE id = $1)`, *lookup.id).Scan(&exists); err != nil {
			return fmt.Errorf("validate resident lookup: %w", err)
		}
		if !exists {
			return fmt.Errorf("%w: lookup", ErrValidation)
		}
	}
	return nil
}

func (s *Service) encryptIndexedValue(value *string) (*string, *string, error) {
	value = nullableTrim(value)
	if value == nil {
		return nil, nil, nil
	}
	encrypted, err := s.crypter.Encrypt(*value)
	if err != nil {
		return nil, nil, fmt.Errorf("encrypt sensitive data: %w", err)
	}
	index := auth.GenerateBlindIndex(*value, s.blindKey)
	return &encrypted, &index, nil
}

func (s *Service) audit(ctx context.Context, tx pgx.Tx, principal *auth.Principal, action, entityType, entityID string) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id)
		VALUES ($1, $2, $3, $4, $5)`,
		principal.OrganizationID, principal.UserID, action, entityType, entityID,
	); err != nil {
		return fmt.Errorf("audit %s: %w", action, err)
	}
	return nil
}

func validResidentStatus(status string) bool {
	switch status {
	case "active", "moved", "deceased", "inactive":
		return true
	}
	return false
}

func maskNationalID(value string) string {
	if len(value) < 4 {
		return "••••"
	}
	return strings.Repeat("•", len(value)-4) + value[len(value)-4:]
}
