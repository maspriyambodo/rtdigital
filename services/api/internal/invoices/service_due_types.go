package invoices

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

func (s *Service) ListDueTypes(ctx context.Context, principal *auth.Principal, status string) ([]DueType, error) {
	status = strings.TrimSpace(status)
	if status != "" && status != "active" && status != "inactive" {
		return nil, fmt.Errorf("%w: status", ErrValidation)
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, name, description, amount, frequency, due_day, status, created_at, updated_at
		FROM due_types
		WHERE organization_id = $1 AND ($2 = '' OR status = $2)
		ORDER BY name, id`,
		principal.OrganizationID, status,
	)
	if err != nil {
		return nil, fmt.Errorf("list due types: %w", err)
	}
	defer rows.Close()

	items := []DueType{}
	for rows.Next() {
		var item DueType
		if err := rows.Scan(
			&item.ID, &item.Name, &item.Description, &item.Amount, &item.Frequency,
			&item.DueDay, &item.Status, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan due type: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) GetDueType(ctx context.Context, principal *auth.Principal, id string) (DueType, error) {
	var item DueType
	err := s.db.QueryRow(ctx, `
		SELECT id, name, description, amount, frequency, due_day, status, created_at, updated_at
		FROM due_types
		WHERE id = $1 AND organization_id = $2`,
		id, principal.OrganizationID,
	).Scan(
		&item.ID, &item.Name, &item.Description, &item.Amount, &item.Frequency,
		&item.DueDay, &item.Status, &item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return DueType{}, ErrDueTypeNotFound
	}
	if err != nil {
		return DueType{}, fmt.Errorf("get due type: %w", err)
	}
	return item, nil
}

func (s *Service) CreateDueType(ctx context.Context, principal *auth.Principal, req CreateDueTypeRequest) (DueType, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || !validFrequency(req.Frequency) {
		return DueType{}, fmt.Errorf("%w: name or frequency", ErrValidation)
	}
	if req.Amount != nil && *req.Amount <= 0 {
		return DueType{}, fmt.Errorf("%w: amount must be positive", ErrValidation)
	}
	if req.DueDay != nil && (*req.DueDay < 1 || *req.DueDay > 31) {
		return DueType{}, fmt.Errorf("%w: due_day", ErrValidation)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return DueType{}, fmt.Errorf("begin create due type: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var item DueType
	err = tx.QueryRow(ctx, `
		INSERT INTO due_types (id, organization_id, name, description, amount, frequency, due_day, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'active')
		RETURNING id, name, description, amount, frequency, due_day, status, created_at, updated_at`,
		newUUID(), principal.OrganizationID, req.Name, nullableTrim(req.Description),
		req.Amount, req.Frequency, req.DueDay,
	).Scan(
		&item.ID, &item.Name, &item.Description, &item.Amount, &item.Frequency,
		&item.DueDay, &item.Status, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return DueType{}, mapDatabaseError(err, "create due type")
	}
	if err := s.audit(ctx, tx, principal, "due_type.create", "due_types", item.ID); err != nil {
		return DueType{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return DueType{}, fmt.Errorf("commit create due type: %w", err)
	}
	return item, nil
}

func (s *Service) UpdateDueType(ctx context.Context, principal *auth.Principal, id string, req UpdateDueTypeRequest) (DueType, error) {
	if req.Name == nil && req.Description == nil && req.Amount == nil && req.Frequency == nil && req.DueDay == nil {
		return DueType{}, fmt.Errorf("%w: no changes", ErrValidation)
	}
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return DueType{}, fmt.Errorf("%w: name", ErrValidation)
	}
	if req.Frequency != nil && !validFrequency(*req.Frequency) {
		return DueType{}, fmt.Errorf("%w: frequency", ErrValidation)
	}
	if req.Amount != nil && *req.Amount <= 0 {
		return DueType{}, fmt.Errorf("%w: amount must be positive", ErrValidation)
	}
	if req.DueDay != nil && (*req.DueDay < 1 || *req.DueDay > 31) {
		return DueType{}, fmt.Errorf("%w: due_day", ErrValidation)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return DueType{}, fmt.Errorf("begin update due type: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var item DueType
	err = tx.QueryRow(ctx, `
		UPDATE due_types
		SET name = COALESCE(NULLIF($1, ''), name),
		    description = COALESCE($2, description),
		    amount = COALESCE($3, amount),
		    frequency = COALESCE($4, frequency),
		    due_day = COALESCE($5, due_day)
		WHERE id = $6 AND organization_id = $7
		RETURNING id, name, description, amount, frequency, due_day, status, created_at, updated_at`,
		nullableTrim(req.Name), nullableTrim(req.Description), req.Amount, req.Frequency,
		req.DueDay, id, principal.OrganizationID,
	).Scan(
		&item.ID, &item.Name, &item.Description, &item.Amount, &item.Frequency,
		&item.DueDay, &item.Status, &item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return DueType{}, ErrDueTypeNotFound
	}
	if err != nil {
		return DueType{}, mapDatabaseError(err, "update due type")
	}
	if err := s.audit(ctx, tx, principal, "due_type.update", "due_types", item.ID); err != nil {
		return DueType{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return DueType{}, fmt.Errorf("commit update due type: %w", err)
	}
	return item, nil
}

func (s *Service) DeactivateDueType(ctx context.Context, principal *auth.Principal, id string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin deactivate due type: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `
		UPDATE due_types SET status = 'inactive'
		WHERE id = $1 AND organization_id = $2 AND status = 'active'`,
		id, principal.OrganizationID,
	)
	if err != nil {
		return fmt.Errorf("deactivate due type: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrDueTypeNotFound
	}
	if err := s.audit(ctx, tx, principal, "due_type.deactivate", "due_types", id); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit deactivate due type: %w", err)
	}
	return nil
}
