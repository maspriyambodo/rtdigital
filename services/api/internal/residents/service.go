package residents

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

var (
	ErrValidation        = errors.New("validation failed")
	ErrHouseUnitNotFound = errors.New("house unit not found")
	ErrResidentNotFound  = errors.New("resident not found")
	ErrHouseholdNotFound = errors.New("household not found")
	ErrDuplicateData     = errors.New("duplicate data")
	ErrConstraint        = errors.New("business constraint violated")
)

type Service struct {
	db       *pgxpool.Pool
	crypter  auth.Crypter
	blindKey string
	now      func() time.Time
}

func NewService(db *pgxpool.Pool, crypter auth.Crypter, blindKey string) *Service {
	return &Service{db: db, crypter: crypter, blindKey: blindKey, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) CreateHouseUnit(ctx context.Context, principal *auth.Principal, req CreateHouseUnitRequest) (HouseUnit, error) {
	req.Code = strings.TrimSpace(req.Code)
	if req.Code == "" || !validOccupancyStatus(req.OccupancyStatus) {
		return HouseUnit{}, fmt.Errorf("%w: code or occupancy_status", ErrValidation)
	}

	var unit HouseUnit
	err := s.db.QueryRow(ctx, `
		INSERT INTO house_units (id, organization_id, code, address_detail, occupancy_status, status)
		VALUES ($1, $2, $3, $4, $5, 'active')
		RETURNING id, code, address_detail, occupancy_status, status, created_at, updated_at`,
		newUUID(), principal.OrganizationID, req.Code, nullableTrim(req.AddressDetail), req.OccupancyStatus,
	).Scan(&unit.ID, &unit.Code, &unit.AddressDetail, &unit.OccupancyStatus, &unit.Status, &unit.CreatedAt, &unit.UpdatedAt)
	if err != nil {
		return HouseUnit{}, mapDatabaseError(err, "create house unit")
	}
	return unit, nil
}

func (s *Service) ListHouseUnits(ctx context.Context, principal *auth.Principal) ([]HouseUnit, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, code, address_detail, occupancy_status, status, created_at, updated_at
		FROM house_units WHERE organization_id = $1 ORDER BY code`, principal.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("list house units: %w", err)
	}
	defer rows.Close()

	units := []HouseUnit{}
	for rows.Next() {
		var unit HouseUnit
		if err := rows.Scan(&unit.ID, &unit.Code, &unit.AddressDetail, &unit.OccupancyStatus, &unit.Status, &unit.CreatedAt, &unit.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan house unit: %w", err)
		}
		units = append(units, unit)
	}
	return units, rows.Err()
}

func (s *Service) GetHouseUnit(ctx context.Context, principal *auth.Principal, id string) (HouseUnit, error) {
	var unit HouseUnit
	err := s.db.QueryRow(ctx, `
		SELECT id, code, address_detail, occupancy_status, status, created_at, updated_at
		FROM house_units WHERE id = $1 AND organization_id = $2`, id, principal.OrganizationID,
	).Scan(&unit.ID, &unit.Code, &unit.AddressDetail, &unit.OccupancyStatus, &unit.Status, &unit.CreatedAt, &unit.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return HouseUnit{}, ErrHouseUnitNotFound
	}
	if err != nil {
		return HouseUnit{}, fmt.Errorf("get house unit: %w", err)
	}
	return unit, nil
}

func (s *Service) UpdateHouseUnit(ctx context.Context, principal *auth.Principal, id string, req UpdateHouseUnitRequest) (HouseUnit, error) {
	if req.Code == nil && req.AddressDetail == nil && req.OccupancyStatus == nil {
		return HouseUnit{}, fmt.Errorf("%w: no changes", ErrValidation)
	}
	if req.Code != nil && strings.TrimSpace(*req.Code) == "" {
		return HouseUnit{}, fmt.Errorf("%w: code", ErrValidation)
	}
	if req.OccupancyStatus != nil && !validOccupancyStatus(*req.OccupancyStatus) {
		return HouseUnit{}, fmt.Errorf("%w: occupancy_status", ErrValidation)
	}

	var unit HouseUnit
	err := s.db.QueryRow(ctx, `
		UPDATE house_units
		SET code = COALESCE(NULLIF($1, ''), code),
		    address_detail = COALESCE($2, address_detail),
		    occupancy_status = COALESCE($3, occupancy_status)
		WHERE id = $4 AND organization_id = $5
		RETURNING id, code, address_detail, occupancy_status, status, created_at, updated_at`,
		nullableTrim(req.Code), nullableTrim(req.AddressDetail), req.OccupancyStatus, id, principal.OrganizationID,
	).Scan(&unit.ID, &unit.Code, &unit.AddressDetail, &unit.OccupancyStatus, &unit.Status, &unit.CreatedAt, &unit.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return HouseUnit{}, ErrHouseUnitNotFound
	}
	if err != nil {
		return HouseUnit{}, mapDatabaseError(err, "update house unit")
	}
	return unit, nil
}

func (s *Service) DeactivateHouseUnit(ctx context.Context, principal *auth.Principal, id string) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE house_units SET status = 'inactive'
		WHERE id = $1 AND organization_id = $2 AND status = 'active'`, id, principal.OrganizationID)
	if err != nil {
		return fmt.Errorf("deactivate house unit: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrHouseUnitNotFound
	}
	return nil
}

func validOccupancyStatus(status string) bool {
	switch status {
	case "owned", "rented", "contract", "empty":
		return true
	}
	return false
}

func nullableTrim(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func mapDatabaseError(err error, operation string) error {
	if strings.Contains(err.Error(), "unique") {
		return ErrDuplicateData
	}
	return fmt.Errorf("%s: %w", operation, err)
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