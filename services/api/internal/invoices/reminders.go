package invoices

import (
	"context"
	"fmt"
	"time"

	"github.com/maspriyambodo/rtdigital/services/api/internal/notifications"
)

type ReminderResult struct {
	Eligible int
	Queued   int
	Skipped  int
	Failures int
}

// RunDueReminders creates one idempotent delivery record per enabled channel.
// The unique database key suppresses repeats from hourly scheduler runs.
func (s *Service) RunDueReminders(ctx context.Context) (ReminderResult, error) {
	now := s.now().UTC()
	today := now.Format("2006-01-02")
	rows, err := s.db.Query(ctx, `
		SELECT i.id, i.organization_id, i.household_id, i.invoice_number, i.due_date::text,
		       CASE
		         WHEN i.due_date < $1::date THEN 'overdue'
		         WHEN i.due_date = $1::date THEN 'due_today'
		         ELSE 'before_due'
		       END,
		       dt.reminder_lead_days
		FROM invoices i
		JOIN due_types dt
		  ON dt.organization_id = i.organization_id
		 AND dt.id = i.due_type_id
		WHERE i.status IN ('unpaid', 'partial', 'pending_verification')
		  AND (
		    i.due_date < $1::date
		    OR i.due_date = $1::date
		    OR i.due_date = ($1::date + dt.reminder_lead_days)
		  )
		ORDER BY i.organization_id, i.due_date, i.id`,
		today,
	)
	if err != nil {
		return ReminderResult{}, fmt.Errorf("list reminder invoices: %w", err)
	}
	defer rows.Close()

	var result ReminderResult
	for rows.Next() {
		var invoiceID, organizationID, householdID, invoiceNumber, dueDate, kind string
		var leadDays int
		if err := rows.Scan(
			&invoiceID, &organizationID, &householdID, &invoiceNumber, &dueDate, &kind, &leadDays,
		); err != nil {
			return ReminderResult{}, fmt.Errorf("scan reminder invoice: %w", err)
		}
		_ = leadDays
		result.Eligible++

		recipients, err := s.reminderRecipients(ctx, organizationID, householdID)
		if err != nil {
			result.Failures++
			continue
		}
		for _, recipient := range recipients {
			channels := map[string]bool{
				"in_app":   recipient.InAppEnabled,
				"email":    recipient.EmailEnabled && recipient.Email != nil,
				"whatsapp": recipient.WhatsAppEnabled && recipient.Phone != nil,
			}
			if !recipient.DueReminderEnabled {
				result.Skipped += len(channels)
				continue
			}

			queuedChannels := make(map[string]bool)
			for channel, enabled := range channels {
				if !enabled {
					result.Skipped++
					continue
				}
				claimed, err := s.claimReminderDelivery(
					ctx, organizationID, invoiceID, recipient.UserID, channel, kind, today,
				)
				if err != nil {
					result.Failures++
					continue
				}
				if !claimed {
					result.Skipped++
					continue
				}
				queuedChannels[channel] = true
				result.Queued++
			}
			if len(queuedChannels) == 0 || s.dispatcher == nil {
				continue
			}
			s.dispatcher.Dispatch(notifications.DispatchJob{
				OrganizationID:  organizationID,
				RecipientUserID: recipient.UserID,
				Type:            "invoice_reminder_" + kind,
				Title:           reminderTitle(kind),
				Body:            fmt.Sprintf("Tagihan %s jatuh tempo %s.", invoiceNumber, dueDate),
				ReferenceType:   "invoice",
				ReferenceID:     invoiceID,
				IsDueReminder:   true,
				Channels:        queuedChannels,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return ReminderResult{}, fmt.Errorf("iterate reminder invoices: %w", err)
	}
	return result, nil
}

type reminderRecipient struct {
	UserID             string
	Email              *string
	Phone              *string
	InAppEnabled       bool
	EmailEnabled       bool
	WhatsAppEnabled    bool
	DueReminderEnabled bool
}

func (s *Service) reminderRecipients(ctx context.Context, organizationID, householdID string) ([]reminderRecipient, error) {
	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT u.id, NULLIF(u.email, ''), NULLIF(u.phone, ''),
		       COALESCE(np.in_app_enabled, true),
		       COALESCE(np.email_enabled, true),
		       COALESCE(np.whatsapp_enabled, false),
		       COALESCE(np.due_reminder_enabled, true)
		FROM household_members hm
		JOIN users u
		  ON u.organization_id = hm.organization_id
		 AND u.resident_id = hm.resident_id
		LEFT JOIN notification_preferences np
		  ON np.organization_id = u.organization_id
		 AND np.user_id = u.id
		WHERE hm.organization_id = $1
		  AND hm.household_id = $2
		  AND hm.is_active
		  AND u.status = 'active'`,
		organizationID, householdID,
	)
	if err != nil {
		return nil, fmt.Errorf("list reminder recipients: %w", err)
	}
	defer rows.Close()

	items := []reminderRecipient{}
	for rows.Next() {
		var item reminderRecipient
		if err := rows.Scan(
			&item.UserID, &item.Email, &item.Phone, &item.InAppEnabled,
			&item.EmailEnabled, &item.WhatsAppEnabled, &item.DueReminderEnabled,
		); err != nil {
			return nil, fmt.Errorf("scan reminder recipient: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) claimReminderDelivery(
	ctx context.Context,
	organizationID, invoiceID, userID, channel, kind, scheduledFor string,
) (bool, error) {
	tag, err := s.db.Exec(ctx, `
		INSERT INTO invoice_reminder_deliveries (
			id, organization_id, invoice_id, user_id, channel, reminder_kind,
			scheduled_for, status, sent_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7::date, 'sent', $8)
		ON CONFLICT (organization_id, invoice_id, user_id, channel, reminder_kind, scheduled_for)
		DO NOTHING`,
		newUUID(), organizationID, invoiceID, userID, channel, kind, scheduledFor, s.now(),
	)
	if err != nil {
		return false, fmt.Errorf("claim reminder delivery: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

func reminderTitle(kind string) string {
	switch kind {
	case "overdue":
		return "Pengingat tagihan tertunggak"
	case "due_today":
		return "Tagihan jatuh tempo hari ini"
	default:
		return "Pengingat jatuh tempo tagihan"
	}
}

// ponytail: dispatch is in-process best-effort; use an outbox and provider receipt
// states when provider-level delivery guarantees become necessary.
var _ = time.Time{}
