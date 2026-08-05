package users

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

func (s *Service) CreateOfficeHandover(ctx context.Context, principal *auth.Principal, req CreateOfficeHandoverRequest) (OfficeHandover, error) {
	if principal == nil || !principal.HasPermission("role.assign") {
		return OfficeHandover{}, ErrForbidden
	}
	if strings.TrimSpace(req.OutgoingUserID) == "" || !validHandoverChecklist(req.Checklist) {
		return OfficeHandover{}, ErrValidation
	}
	if req.OutgoingUserID == principal.UserID {
		return OfficeHandover{}, ErrCannotModifySelf
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return OfficeHandover{}, fmt.Errorf("begin create handover: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var eligible bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM users u
			JOIN user_roles ur ON ur.user_id = u.id
			JOIN roles r ON r.id = ur.role_id
			WHERE u.organization_id = $1
			  AND u.id = $2
			  AND r.code IN ('ketua_rt', 'sekretaris', 'bendahara')
		)`,
		principal.OrganizationID, req.OutgoingUserID,
	).Scan(&eligible); err != nil {
		return OfficeHandover{}, fmt.Errorf("validate outgoing handover user: %w", err)
	}
	if !eligible {
		return OfficeHandover{}, ErrUserNotFound
	}

	id, now := newUUID(), s.now()
	if _, err := tx.Exec(ctx, `
		INSERT INTO office_handovers (
			id, organization_id, outgoing_user_id, status, checklist, notes, created_at, updated_at
		) VALUES ($1, $2, $3, 'draft', $4, $5, $6, $6)`,
		id, principal.OrganizationID, req.OutgoingUserID, req.Checklist, nullableString(req.Notes), now,
	); err != nil {
		return OfficeHandover{}, fmt.Errorf("create office handover: %w", err)
	}
	if err := s.audit(ctx, tx, principal, "office_handover.create", "office_handover", id); err != nil {
		return OfficeHandover{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return OfficeHandover{}, fmt.Errorf("commit create handover: %w", err)
	}
	return s.GetOfficeHandover(ctx, principal, id)
}

func (s *Service) CompleteOfficeHandover(ctx context.Context, principal *auth.Principal, id string, req CompleteOfficeHandoverRequest) (OfficeHandover, error) {
	if principal == nil || !principal.HasPermission("role.assign") {
		return OfficeHandover{}, ErrForbidden
	}
	if strings.TrimSpace(id) == "" || strings.TrimSpace(req.IncomingUserID) == "" || !validHandoverChecklist(req.Checklist) {
		return OfficeHandover{}, ErrValidation
	}
	if req.IncomingUserID == principal.UserID {
		return OfficeHandover{}, ErrCannotModifySelf
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return OfficeHandover{}, fmt.Errorf("begin complete handover: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var outgoingUserID string
	if err := tx.QueryRow(ctx, `
		SELECT outgoing_user_id
		FROM office_handovers
		WHERE id = $1 AND organization_id = $2 AND status = 'draft'
		FOR UPDATE`,
		id, principal.OrganizationID,
	).Scan(&outgoingUserID); errors.Is(err, pgx.ErrNoRows) {
		return OfficeHandover{}, ErrValidation
	} else if err != nil {
		return OfficeHandover{}, fmt.Errorf("lock office handover: %w", err)
	}
	if outgoingUserID == req.IncomingUserID {
		return OfficeHandover{}, ErrValidation
	}

	var incomingActive bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM users
			WHERE id = $1 AND organization_id = $2 AND status = 'active'
		)`,
		req.IncomingUserID, principal.OrganizationID,
	).Scan(&incomingActive); err != nil {
		return OfficeHandover{}, fmt.Errorf("validate incoming handover user: %w", err)
	}
	if !incomingActive {
		return OfficeHandover{}, ErrUserNotFound
	}

	var incomingMFAEnabled bool
	if err := tx.QueryRow(ctx, `
		SELECT mfa_enabled_at IS NOT NULL
		FROM users
		WHERE id = $1 AND organization_id = $2`,
		req.IncomingUserID, principal.OrganizationID,
	).Scan(&incomingMFAEnabled); err != nil {
		return OfficeHandover{}, fmt.Errorf("check incoming user MFA: %w", err)
	}

	roles, err := tx.Query(ctx, `
		SELECT ur.role_id, r.mfa_required
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1
		  AND r.code IN ('ketua_rt', 'sekretaris', 'bendahara')
		  AND (r.organization_id IS NULL OR r.organization_id = $2)`,
		outgoingUserID, principal.OrganizationID,
	)
	if err != nil {
		return OfficeHandover{}, fmt.Errorf("list outgoing officer roles: %w", err)
	}
	roleIDs := []string{}
	for roles.Next() {
		var (
			roleID      string
			mfaRequired bool
		)
		if err := roles.Scan(&roleID, &mfaRequired); err != nil {
			roles.Close()
			return OfficeHandover{}, fmt.Errorf("scan outgoing officer role: %w", err)
		}
		if mfaRequired && !incomingMFAEnabled {
			roles.Close()
			return OfficeHandover{}, ErrMFAEnrollmentRequired
		}
		roleIDs = append(roleIDs, roleID)
	}
	if err := roles.Err(); err != nil {
		roles.Close()
		return OfficeHandover{}, fmt.Errorf("iterate outgoing officer roles: %w", err)
	}
	roles.Close()
	if len(roleIDs) == 0 {
		return OfficeHandover{}, ErrValidation
	}

	for _, roleID := range roleIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id, assigned_by)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id, role_id) DO NOTHING`,
			req.IncomingUserID, roleID, principal.UserID,
		); err != nil {
			return OfficeHandover{}, fmt.Errorf("assign incoming officer role: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM user_roles
		WHERE user_id = $1 AND role_id = ANY($2::uuid[])`,
		outgoingUserID, roleIDs,
	); err != nil {
		return OfficeHandover{}, fmt.Errorf("revoke outgoing officer roles: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE sessions
		SET revoked_at = $1, revoke_reason = 'office_handover_completed'
		WHERE user_id = $2 AND revoked_at IS NULL`,
		s.now(), outgoingUserID,
	); err != nil {
		return OfficeHandover{}, fmt.Errorf("revoke outgoing sessions: %w", err)
	}

	now := s.now()
	if _, err := tx.Exec(ctx, `
		UPDATE office_handovers
		SET incoming_user_id = $1,
		    status = 'completed',
		    checklist = $2,
		    notes = $3,
		    completed_by = $4,
		    completed_at = $5,
		    updated_at = $5
		WHERE id = $6 AND organization_id = $7`,
		req.IncomingUserID, req.Checklist, nullableString(req.Notes), principal.UserID,
		now, id, principal.OrganizationID,
	); err != nil {
		return OfficeHandover{}, fmt.Errorf("complete office handover: %w", err)
	}
	if err := s.audit(ctx, tx, principal, "office_handover.complete", "office_handover", id); err != nil {
		return OfficeHandover{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return OfficeHandover{}, fmt.Errorf("commit office handover: %w", err)
	}
	return s.GetOfficeHandover(ctx, principal, id)
}

func (s *Service) GetOfficeHandover(ctx context.Context, principal *auth.Principal, id string) (OfficeHandover, error) {
	if principal == nil || strings.TrimSpace(id) == "" {
		return OfficeHandover{}, ErrValidation
	}

	var item OfficeHandover
	err := s.db.QueryRow(ctx, `
		SELECT id, outgoing_user_id, incoming_user_id, status, checklist, notes,
		       completed_by, completed_at, created_at, updated_at
		FROM office_handovers
		WHERE id = $1 AND organization_id = $2`,
		id, principal.OrganizationID,
	).Scan(
		&item.ID, &item.OutgoingUserID, &item.IncomingUserID, &item.Status,
		&item.Checklist, &item.Notes, &item.CompletedBy, &item.CompletedAt,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return OfficeHandover{}, ErrUserNotFound
	}
	if err != nil {
		return OfficeHandover{}, fmt.Errorf("get office handover: %w", err)
	}
	return item, nil
}

func validHandoverChecklist(value json.RawMessage) bool {
	var items map[string]bool
	return len(value) > 0 && json.Unmarshal(value, &items) == nil && len(items) > 0
}

func nullableString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
