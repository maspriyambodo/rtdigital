package reports

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

// Mutations reports household-membership history. Resident-status changes cannot be
// reconstructed because existing audit records do not retain before/after values.
func (s *Service) Mutations(ctx context.Context, principal *auth.Principal, filter Filter) ([]MutationReportItem, error) {
	if principal == nil {
		return nil, ErrForbidden
	}
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	conditions := []string{"h.organization_id = $1"}
	args := []any{principal.OrganizationID}
	index := 2
	if filter.Status != "" {
		switch filter.Status {
		case "active":
			conditions = append(conditions, "hm.is_active")
		case "ended":
			conditions = append(conditions, "NOT hm.is_active")
		default:
			return nil, ErrValidation
		}
	}
	if filter.StartDate != "" {
		conditions = append(conditions, fmt.Sprintf("COALESCE(hm.ended_at, hm.started_at) >= $%d::date", index))
		args, index = append(args, filter.StartDate), index+1
	}
	if filter.EndDate != "" {
		conditions = append(conditions, fmt.Sprintf("COALESCE(hm.ended_at, hm.started_at) <= $%d::date", index))
		args = append(args, filter.EndDate)
	}

	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT hm.id, r.full_name, h.internal_number, hu.code, hm.relationship,
		       hm.started_at, hm.ended_at,
		       CASE WHEN hm.is_active THEN 'active' ELSE 'ended' END
		FROM household_members hm
		JOIN residents r ON r.id = hm.resident_id AND r.organization_id = $1
		JOIN households h ON h.id = hm.household_id AND h.organization_id = $1
		JOIN house_units hu ON hu.id = h.house_unit_id AND hu.organization_id = $1
		WHERE %s
		ORDER BY COALESCE(hm.ended_at, hm.started_at) DESC, r.full_name ASC`,
		strings.Join(conditions, " AND "),
	), args...)
	if err != nil {
		return nil, fmt.Errorf("query resident mutations report: %w", err)
	}
	defer rows.Close()

	items := []MutationReportItem{}
	for rows.Next() {
		var item MutationReportItem
		var startedAt time.Time
		var endedAt *time.Time
		if err := rows.Scan(
			&item.ID,
			&item.FullName,
			&item.HouseholdNumber,
			&item.HouseUnitCode,
			&item.Relationship,
			&startedAt,
			&endedAt,
			&item.Status,
		); err != nil {
			return nil, fmt.Errorf("scan resident mutation report: %w", err)
		}
		item.StartedAt = startedAt.Format(time.DateOnly)
		if endedAt != nil {
			item.EndedAt = endedAt.Format(time.DateOnly)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}