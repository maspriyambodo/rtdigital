package residents

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

func (s *Service) CreateHousehold(ctx context.Context, principal *auth.Principal, req CreateHouseholdRequest) (Household, error) {
	req.InternalNumber = strings.TrimSpace(req.InternalNumber)
	if req.InternalNumber == "" || req.HouseUnitID == "" || !validDomicileStatus(req.DomicileStatus) {
		return Household{}, fmt.Errorf("%w: internal_number, house_unit_id, or domicile_status", ErrValidation)
	}

	encryptedKK, blindIndex, err := s.encryptIndexedValue(req.FamilyCardNumber)
	if err != nil {
		return Household{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Household{}, fmt.Errorf("begin household create: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var household Household
	err = tx.QueryRow(ctx, `
		INSERT INTO households (
			id, organization_id, house_unit_id, internal_number,
			family_card_number_encrypted, family_card_blind_index,
			domicile_status, move_in_date, verification_status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8::date, 'unverified')
		RETURNING id, house_unit_id, internal_number, domicile_status,
			move_in_date::text, move_out_date::text, verification_status, created_at, updated_at`,
		newUUID(), principal.OrganizationID, req.HouseUnitID, req.InternalNumber,
		encryptedKK, blindIndex, req.DomicileStatus, nullableTrim(req.MoveInDate),
	).Scan(
		&household.ID, &household.HouseUnitID, &household.InternalNumber, &household.DomicileStatus,
		&household.MoveInDate, &household.MoveOutDate, &household.VerificationStatus,
		&household.CreatedAt, &household.UpdatedAt,
	)
	if err != nil {
		return Household{}, mapDatabaseError(err, "create household")
	}
	if err := s.audit(ctx, tx, principal, "household.create", "households", household.ID); err != nil {
		return Household{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Household{}, fmt.Errorf("commit household create: %w", err)
	}
	return household, nil
}

func (s *Service) ListHouseholds(ctx context.Context, principal *auth.Principal) ([]Household, error) {
	rows, err := s.db.Query(ctx, `
		SELECT h.id, h.house_unit_id, h.internal_number, h.head_resident_id, r.full_name,
			h.domicile_status, h.move_in_date::text, h.move_out_date::text,
			h.verification_status, h.created_at, h.updated_at
		FROM households h
		LEFT JOIN residents r ON r.id = h.head_resident_id AND r.organization_id = h.organization_id
		WHERE h.organization_id = $1
		ORDER BY h.internal_number`, principal.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("list households: %w", err)
	}
	defer rows.Close()

	households := []Household{}
	for rows.Next() {
		var household Household
		if err := rows.Scan(
			&household.ID, &household.HouseUnitID, &household.InternalNumber,
			&household.HeadResidentID, &household.HeadResidentName, &household.DomicileStatus,
			&household.MoveInDate, &household.MoveOutDate, &household.VerificationStatus,
			&household.CreatedAt, &household.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan household: %w", err)
		}
		households = append(households, household)
	}
	return households, rows.Err()
}

func (s *Service) GetHousehold(ctx context.Context, principal *auth.Principal, id string) (Household, error) {
	var household Household
	var encryptedKK *string
	err := s.db.QueryRow(ctx, `
		SELECT h.id, h.house_unit_id, h.internal_number, h.family_card_number_encrypted,
			h.head_resident_id, r.full_name, h.domicile_status, h.move_in_date::text,
			h.move_out_date::text, h.verification_status, h.created_at, h.updated_at
		FROM households h
		LEFT JOIN residents r ON r.id = h.head_resident_id AND r.organization_id = h.organization_id
		WHERE h.id = $1 AND h.organization_id = $2`, id, principal.OrganizationID,
	).Scan(
		&household.ID, &household.HouseUnitID, &household.InternalNumber, &encryptedKK,
		&household.HeadResidentID, &household.HeadResidentName, &household.DomicileStatus,
		&household.MoveInDate, &household.MoveOutDate, &household.VerificationStatus,
		&household.CreatedAt, &household.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Household{}, ErrHouseholdNotFound
	}
	if err != nil {
		return Household{}, fmt.Errorf("get household: %w", err)
	}
	if encryptedKK != nil {
		masked := maskNationalID("")
		household.FamilyCardNumber = &masked
	}

	members, err := s.listHouseholdMembers(ctx, principal.OrganizationID, id)
	if err != nil {
		return Household{}, err
	}
	household.Members = members
	return household, nil
}

func (s *Service) AddHouseholdMember(ctx context.Context, principal *auth.Principal, householdID string, req HouseholdMemberRequest) error {
	if req.ResidentID == "" || !validRelationship(req.Relationship) {
		return fmt.Errorf("%w: resident_id or relationship", ErrValidation)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin add household member: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM households h
			JOIN residents r ON r.id = $3 AND r.organization_id = h.organization_id
			WHERE h.id = $1 AND h.organization_id = $2
		)`, householdID, principal.OrganizationID, req.ResidentID,
	).Scan(&exists); err != nil {
		return fmt.Errorf("validate household member tenant: %w", err)
	}
	if !exists {
		return ErrConstraint
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO household_members (
			id, organization_id, household_id, resident_id, relationship, is_active, started_at
		) VALUES ($1, $2, $3, $4, $5, true, current_date)`,
		newUUID(), principal.OrganizationID, householdID, req.ResidentID, req.Relationship,
	); err != nil {
		if strings.Contains(err.Error(), "uq_active_household_member_per_resident") ||
			strings.Contains(err.Error(), "uq_active_household_head") {
			return ErrConstraint
		}
		return mapDatabaseError(err, "add household member")
	}

	if req.Relationship == "head" {
		if _, err := tx.Exec(ctx, `
			UPDATE households SET head_resident_id = $1
			WHERE id = $2 AND organization_id = $3`,
			req.ResidentID, householdID, principal.OrganizationID,
		); err != nil {
			return fmt.Errorf("set household head: %w", err)
		}
	}

	if err := s.audit(ctx, tx, principal, "household.member.add", "households", householdID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) listHouseholdMembers(ctx context.Context, organizationID, householdID string) ([]HouseholdMember, error) {
	rows, err := s.db.Query(ctx, `
		SELECT hm.id, hm.resident_id, r.full_name, hm.relationship, hm.is_active,
			hm.started_at::text, hm.ended_at::text
		FROM household_members hm
		JOIN residents r ON r.id = hm.resident_id AND r.organization_id = hm.organization_id
		WHERE hm.organization_id = $1 AND hm.household_id = $2
		ORDER BY hm.is_active DESC, (hm.relationship = 'head') DESC, hm.started_at DESC`,
		organizationID, householdID,
	)
	if err != nil {
		return nil, fmt.Errorf("list household members: %w", err)
	}
	defer rows.Close()

	members := []HouseholdMember{}
	for rows.Next() {
		var member HouseholdMember
		if err := rows.Scan(
			&member.ID, &member.ResidentID, &member.ResidentName, &member.Relationship,
			&member.IsActive, &member.StartedAt, &member.EndedAt,
		); err != nil {
			return nil, fmt.Errorf("scan household member: %w", err)
		}
		members = append(members, member)
	}
	return members, rows.Err()
}

func validDomicileStatus(status string) bool {
	return status == "permanent" || status == "temporary"
}

func validRelationship(relationship string) bool {
	switch relationship {
	case "head", "spouse", "child", "parent", "other":
		return true
	}
	return false
}
