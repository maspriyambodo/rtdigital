package reports

import (
	"context"
	"fmt"
	"strings"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

func (s *Service) Arrears(ctx context.Context, principal *auth.Principal, filter Filter) ([]ArrearsReportItem, error) {
	if principal == nil {
		return nil, ErrForbidden
	}
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	conditions := []string{
		"i.organization_id = $1",
		"i.status IN ('unpaid', 'partial', 'pending_verification')",
	}
	args := []any{principal.OrganizationID}
	index := 2
	if filter.StartDate != "" {
		conditions = append(conditions, fmt.Sprintf("i.due_date >= $%d", index))
		args, index = append(args, filter.StartDate), index+1
	}
	if filter.EndDate != "" {
		conditions = append(conditions, fmt.Sprintf("i.due_date <= $%d", index))
		args = append(args, filter.EndDate)
	}

	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT h.internal_number, hu.code, COALESCE(r.full_name, ''), COUNT(i.id),
		       SUM(i.amount - i.paid_amount - i.adjustment_amount)
		FROM invoices i
		JOIN households h ON h.id = i.household_id AND h.organization_id = i.organization_id
		JOIN house_units hu ON hu.id = h.house_unit_id AND hu.organization_id = i.organization_id
		LEFT JOIN residents r ON r.id = h.head_resident_id AND r.organization_id = i.organization_id
		WHERE %s
		GROUP BY h.id, h.internal_number, hu.code, r.full_name
		HAVING SUM(i.amount - i.paid_amount - i.adjustment_amount) > 0
		ORDER BY SUM(i.amount - i.paid_amount - i.adjustment_amount) DESC`,
		strings.Join(conditions, " AND "),
	), args...)
	if err != nil {
		return nil, fmt.Errorf("query arrears report: %w", err)
	}
	defer rows.Close()

	items := []ArrearsReportItem{}
	for rows.Next() {
		var item ArrearsReportItem
		if err := rows.Scan(
			&item.HouseholdNumber,
			&item.HouseUnitCode,
			&item.HeadResidentName,
			&item.InvoiceCount,
			&item.TotalArrears,
		); err != nil {
			return nil, fmt.Errorf("scan arrears report: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}