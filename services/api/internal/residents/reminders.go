package residents

import (
	"context"
	"fmt"

	"github.com/maspriyambodo/rtdigital/services/api/internal/notifications"
)

type domicileReminder struct {
	organizationID string
	householdID    string
	userID         string
}

// RunDomicileReminders queues one in-app reminder per due household member per
// day. Repeated scheduler runs are idempotent. Dispatcher failures never
// invalidate the reminder ledger.
func (s *Service) RunDomicileReminders(ctx context.Context) (int, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin domicile reminders: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT h.organization_id, h.id, u.id
		FROM households h
		JOIN household_members hm
		  ON hm.organization_id = h.organization_id
		 AND hm.household_id = h.id
		 AND hm.is_active
		JOIN users u
		  ON u.organization_id = hm.organization_id
		 AND u.resident_id = hm.resident_id
		 AND u.status = 'active'
		WHERE h.move_out_date IS NULL
		  AND h.domicile_review_due_at IS NOT NULL
		  AND h.domicile_review_due_at <= CURRENT_DATE`,
	)
	if err != nil {
		return 0, fmt.Errorf("query due domicile reminders: %w", err)
	}

	candidates := []domicileReminder{}
	for rows.Next() {
		var item domicileReminder
		if err := rows.Scan(&item.organizationID, &item.householdID, &item.userID); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan due domicile reminder: %w", err)
		}
		candidates = append(candidates, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, fmt.Errorf("iterate due domicile reminders: %w", err)
	}
	rows.Close()

	deliveries := []domicileReminder{}
	for _, item := range candidates {
		tag, err := tx.Exec(ctx, `
			INSERT INTO domicile_reminder_deliveries (
				id, organization_id, household_id, user_id, reminder_date, status
			) VALUES ($1, $2, $3, $4, CURRENT_DATE, 'sent')
			ON CONFLICT (organization_id, household_id, user_id, reminder_date) DO NOTHING`,
			newUUID(), item.organizationID, item.householdID, item.userID,
		)
		if err != nil {
			return 0, fmt.Errorf("record domicile reminder: %w", err)
		}
		if tag.RowsAffected() > 0 {
			deliveries = append(deliveries, item)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit domicile reminders: %w", err)
	}

	for _, item := range deliveries {
		s.dispatchNotification(notifications.DispatchJob{
			OrganizationID:  item.organizationID,
			RecipientUserID: item.userID,
			Type:            "domicile_confirmation_reminder",
			Title:           "Konfirmasi domisili",
			Body:            "Mohon konfirmasi status tinggal atau pindah untuk pembaruan data RT.",
			ReferenceType:   "household",
			ReferenceID:     item.householdID,
		})
	}
	return len(deliveries), nil
}
