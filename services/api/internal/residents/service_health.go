package residents

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

type healthResident struct {
	MissingContact bool `json:"missing_contact"`
	Unverified     bool `json:"unverified"`
	Stale          bool `json:"stale"`
}

// ListHouseholdHealthScores returns an operational worklist. It never exposes
// resident identities, contact values, NIK/KK values, or sanctions.
func (s *Service) ListHouseholdHealthScores(ctx context.Context, principal *auth.Principal) ([]HouseholdHealthScore, error) {
	if principal == nil || !principal.HasPermission("resident.read") {
		return nil, ErrForbidden
	}

	rows, err := s.db.Query(ctx, `
		SELECT h.id, h.internal_number, h.domicile_review_due_at::text, h.updated_at,
		       COALESCE(
		         jsonb_agg(
		           jsonb_build_object(
		             'missing_contact', (r.phone IS NULL AND r.email IS NULL),
		             'unverified', (r.verification_status <> 'verified'),
		             'stale', (r.updated_at < now() - interval '1 year')
		           )
		         ) FILTER (WHERE r.id IS NOT NULL),
		         '[]'::jsonb
		       )
		FROM households h
		LEFT JOIN household_members hm
		  ON hm.organization_id = h.organization_id
		 AND hm.household_id = h.id
		 AND hm.is_active
		LEFT JOIN residents r
		  ON r.organization_id = hm.organization_id
		 AND r.id = hm.resident_id
		WHERE h.organization_id = $1
		  AND h.move_out_date IS NULL
		GROUP BY h.id
		ORDER BY h.internal_number`,
		principal.OrganizationID,
	)
	if err != nil {
		return nil, fmt.Errorf("list household health scores: %w", err)
	}
	defer rows.Close()

	items := []HouseholdHealthScore{}
	for rows.Next() {
		var (
			item HouseholdHealthScore
			raw  json.RawMessage
		)
		if err := rows.Scan(
			&item.HouseholdID, &item.InternalNumber, &item.DomicileDueAt,
			&item.UpdatedAt, &raw,
		); err != nil {
			return nil, fmt.Errorf("scan household health score: %w", err)
		}

		var residents []healthResident
		if err := json.Unmarshal(raw, &residents); err != nil {
			return nil, fmt.Errorf("decode household health score: %w", err)
		}

		if len(residents) == 0 {
			item.Score = 0
			item.MissingItems = []string{"active_members"}
		} else {
			total, missing := len(residents)*3, 0
			for _, resident := range residents {
				if resident.MissingContact {
					item.MissingItems = append(item.MissingItems, "contact")
					missing++
				}
				if resident.Unverified {
					item.MissingItems = append(item.MissingItems, "verification")
					missing++
				}
				if resident.Stale {
					item.MissingItems = append(item.MissingItems, "data_update")
					missing++
				}
			}
			if item.DomicileDueAt != nil {
				due, err := time.Parse(time.DateOnly, *item.DomicileDueAt)
				if err == nil && !due.After(s.now().UTC()) {
					item.MissingItems = append(item.MissingItems, "domicile_confirmation")
					missing++
					total++
				}
			}
			item.Score = (total - missing) * 100 / total
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate household health scores: %w", err)
	}
	return items, nil
}
