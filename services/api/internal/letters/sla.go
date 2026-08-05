package letters

import (
	"context"
	"fmt"

	"github.com/maspriyambodo/rtdigital/services/api/internal/notifications"
)

type letterEscalation struct {
	organizationID  string
	letterID        string
	requesterUserID string
	requestNumber   string
}

// EscalateLetters marks each overdue active request once. Safe for repeated
// scheduler runs; provider failures remain best-effort after commit.
func (s *Service) EscalateLetters(ctx context.Context) (int, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin letter escalation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := s.now()
	rows, err := tx.Query(ctx, `
		UPDATE letter_requests
		SET sla_escalated_at = $1
		WHERE sla_due_at IS NOT NULL
		  AND sla_escalated_at IS NULL
		  AND sla_due_at <= $1
		  AND status IN ('submitted', 'under_review', 'needs_revision', 'awaiting_approval')
		RETURNING organization_id, id, requester_user_id, request_number`,
		now,
	)
	if err != nil {
		return 0, fmt.Errorf("escalate overdue letters: %w", err)
	}

	escalations := []letterEscalation{}
	for rows.Next() {
		var item letterEscalation
		if err := rows.Scan(&item.organizationID, &item.letterID, &item.requesterUserID, &item.requestNumber); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan escalated letter: %w", err)
		}
		escalations = append(escalations, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, fmt.Errorf("iterate escalated letters: %w", err)
	}
	rows.Close()

	for _, item := range escalations {
		if _, err := tx.Exec(ctx, `
			INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id, metadata)
			VALUES ($1, NULL, 'letter_request.escalate', 'letter_request', $2,
			        jsonb_build_object('sla_escalated_at', $3))`,
			item.organizationID, item.letterID, now,
		); err != nil {
			return 0, fmt.Errorf("audit letter escalation: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit letter escalation: %w", err)
	}

	for _, item := range escalations {
		s.notifySecretariesAndRT(item.organizationID, notifications.DispatchJob{
			Type:          "letter_request_escalated",
			Title:         "Eskalasi pengajuan surat",
			Body:          fmt.Sprintf("Pengajuan %s telah melewati SLA.", item.requestNumber),
			ReferenceType: "letter_request",
			ReferenceID:   item.letterID,
		})
		s.dispatchNotification(notifications.DispatchJob{
			OrganizationID:  item.organizationID,
			RecipientUserID: item.requesterUserID,
			Type:            "letter_request_sla_overdue",
			Title:           "Status pengajuan surat",
			Body:            fmt.Sprintf("Pengajuan %s melewati estimasi waktu proses.", item.requestNumber),
			ReferenceType:   "letter_request",
			ReferenceID:     item.letterID,
		})
	}
	return len(escalations), nil
}
