package residents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

var ErrCorrectionNotFound = errors.New("resident correction not found")

var correctionFields = map[string]struct{}{
	"full_name":       {},
	"birth_place":     {},
	"birth_date":      {},
	"gender":          {},
	"marital_status":  {},
	"occupation":      {},
	"education":       {},
	"phone":           {},
	"email":           {},
	"resident_status": {},
}

func (s *Service) SubmitCorrection(ctx context.Context, principal *auth.Principal, residentID string, req CreateResidentCorrectionRequest) (ResidentCorrection, error) {
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" || len(req.RequestedChanges) == 0 || !validCorrectionFields(req.RequestedChanges) {
		return ResidentCorrection{}, fmt.Errorf("%w: reason or requested_changes", ErrValidation)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ResidentCorrection{}, fmt.Errorf("begin correction submission: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var hasScope bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM users u
			WHERE u.id = $1
			  AND u.organization_id = $2
			  AND (
			      u.resident_id = $3
			      OR EXISTS (
			          SELECT 1
			          FROM household_members own
			          JOIN household_members target
			            ON target.household_id = own.household_id
			           AND target.organization_id = own.organization_id
			          WHERE own.organization_id = $2
			            AND own.resident_id = u.resident_id
			            AND target.resident_id = $3
			            AND own.is_active
			            AND target.is_active
			      )
			  )
		)`, principal.UserID, principal.OrganizationID, residentID).Scan(&hasScope); err != nil {
		return ResidentCorrection{}, fmt.Errorf("validate correction scope: %w", err)
	}
	if !hasScope {
		return ResidentCorrection{}, ErrResidentNotFound
	}

	changes, err := json.Marshal(req.RequestedChanges)
	if err != nil {
		return ResidentCorrection{}, fmt.Errorf("%w: requested_changes", ErrValidation)
	}

	var correction ResidentCorrection
	err = tx.QueryRow(ctx, `
		INSERT INTO resident_corrections (
			id, organization_id, resident_id, requester_user_id, requested_changes, reason, status
		)
		SELECT $1, $2, r.id, $3, $4::jsonb, $5, 'submitted'
		FROM residents r
		WHERE r.id = $6 AND r.organization_id = $2
		RETURNING id, resident_id, requester_user_id, requested_changes, reason, status, created_at, updated_at`,
		newUUID(), principal.OrganizationID, principal.UserID, changes, req.Reason, residentID,
	).Scan(
		&correction.ID, &correction.ResidentID, &correction.RequesterUserID,
		&correction.RequestedChanges, &correction.Reason, &correction.Status,
		&correction.CreatedAt, &correction.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ResidentCorrection{}, ErrResidentNotFound
	}
	if err != nil {
		return ResidentCorrection{}, fmt.Errorf("submit correction: %w", err)
	}
	if err := s.audit(ctx, tx, principal, "resident.correction.submit", "resident_corrections", correction.ID); err != nil {
		return ResidentCorrection{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ResidentCorrection{}, fmt.Errorf("commit correction submission: %w", err)
	}
	return correction, nil
}

func (s *Service) ListCorrections(ctx context.Context, principal *auth.Principal, status string) ([]ResidentCorrection, error) {
	if status != "" && !validCorrectionStatus(status) {
		return nil, fmt.Errorf("%w: status", ErrValidation)
	}
	rows, err := s.db.Query(ctx, `
		SELECT c.id, c.resident_id, r.full_name, c.requester_user_id, c.requested_changes,
		       c.reason, c.status, c.reviewer_user_id, c.reviewer_note, c.reviewed_at,
		       c.created_at, c.updated_at
		FROM resident_corrections c
		JOIN residents r ON r.id = c.resident_id AND r.organization_id = c.organization_id
		WHERE c.organization_id = $1 AND ($2 = '' OR c.status = $2)
		ORDER BY c.created_at DESC`, principal.OrganizationID, status)
	if err != nil {
		return nil, fmt.Errorf("list corrections: %w", err)
	}
	defer rows.Close()

	items := []ResidentCorrection{}
	for rows.Next() {
		var item ResidentCorrection
		if err := rows.Scan(
			&item.ID, &item.ResidentID, &item.ResidentName, &item.RequesterUserID,
			&item.RequestedChanges, &item.Reason, &item.Status, &item.ReviewerUserID,
			&item.ReviewerNote, &item.ReviewedAt, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan correction: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) ReviewCorrection(ctx context.Context, principal *auth.Principal, id, action string, req ReviewResidentCorrectionRequest) error {
	req.Note = strings.TrimSpace(req.Note)
	status, auditAction, err := correctionReviewAction(action)
	if err != nil {
		return err
	}
	if (status == "rejected" || status == "needs_revision") && req.Note == "" {
		return fmt.Errorf("%w: note required", ErrValidation)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin correction review: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var residentID string
	var changes []byte
	err = tx.QueryRow(ctx, `
		UPDATE resident_corrections
		SET status = $1, reviewer_user_id = $2, reviewer_note = NULLIF($3, ''), reviewed_at = now()
		WHERE id = $4 AND organization_id = $5 AND status IN ('submitted', 'needs_revision')
		RETURNING resident_id, requested_changes`,
		status, principal.UserID, req.Note, id, principal.OrganizationID,
	).Scan(&residentID, &changes)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrCorrectionNotFound
	}
	if err != nil {
		return fmt.Errorf("review correction: %w", err)
	}

	if status == "approved" {
		var requested map[string]any
		if err := json.Unmarshal(changes, &requested); err != nil || !validCorrectionFields(requested) {
			return fmt.Errorf("%w: stored requested_changes", ErrConstraint)
		}
		data, _ := json.Marshal(requested)
		if _, err := tx.Exec(ctx, `
			UPDATE residents
			SET full_name = COALESCE($2::jsonb->>'full_name', full_name),
			    birth_place = COALESCE($2::jsonb->>'birth_place', birth_place),
			    birth_date = COALESCE(($2::jsonb->>'birth_date')::date, birth_date),
			    gender = COALESCE($2::jsonb->>'gender', gender),
			    marital_status = COALESCE($2::jsonb->>'marital_status', marital_status),
			    occupation = COALESCE($2::jsonb->>'occupation', occupation),
			    education = COALESCE($2::jsonb->>'education', education),
			    phone = COALESCE($2::jsonb->>'phone', phone),
			    email = COALESCE($2::jsonb->>'email', email),
			    resident_status = COALESCE($2::jsonb->>'resident_status', resident_status)
			WHERE id = $1 AND organization_id = $3`, residentID, data, principal.OrganizationID); err != nil {
			return fmt.Errorf("apply correction: %w", err)
		}
	}

	if err := s.audit(ctx, tx, principal, auditAction, "resident_corrections", id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func validCorrectionFields(changes map[string]any) bool {
	for field, value := range changes {
		if _, ok := correctionFields[field]; !ok || value == nil {
			return false
		}
	}
	return true
}

func validCorrectionStatus(status string) bool {
	switch status {
	case "submitted", "approved", "rejected", "needs_revision":
		return true
	}
	return false
}

func correctionReviewAction(action string) (status, auditAction string, err error) {
	switch action {
	case "approve":
		return "approved", "resident.correction.approve", nil
	case "reject":
		return "rejected", "resident.correction.reject", nil
	case "request-revision":
		return "needs_revision", "resident.correction.request_revision", nil
	default:
		return "", "", fmt.Errorf("%w: action", ErrValidation)
	}
}