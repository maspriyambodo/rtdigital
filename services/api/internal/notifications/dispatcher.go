package notifications

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailSender interface {
	SendEmail(context.Context, string, string, string) error
}

type NoopEmailSender struct{}

func (NoopEmailSender) SendEmail(context.Context, string, string, string) error {
	return nil
}

type Dispatcher struct {
	db       *pgxpool.Pool
	service  *Service
	mailer   EmailSender
	whatsapp WhatsAppSender
	logger   *slog.Logger
}

type DispatchJob struct {
	OrganizationID  string
	RecipientUserID string
	Type            string
	Title           string
	Body            string
	ReferenceType   string
	ReferenceID     string
	IsDueReminder   bool
	Channels        map[string]bool
}

func NewDispatcher(
	db *pgxpool.Pool,
	service *Service,
	mailer EmailSender,
	whatsapp WhatsAppSender,
	logger *slog.Logger,
) *Dispatcher {
	if mailer == nil {
		mailer = NoopEmailSender{}
	}
	if whatsapp == nil {
		whatsapp = NoopWhatsAppSender{}
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Dispatcher{
		db:       db,
		service:  service,
		mailer:   mailer,
		whatsapp: whatsapp,
		logger:   logger,
	}
}

// DispatchNotification adapts domain services without creating package dependencies.
func (d *Dispatcher) DispatchNotification(organizationID, recipientUserID, notificationType, title, body, referenceType, referenceID string) {
	d.Dispatch(DispatchJob{
		OrganizationID:  organizationID,
		RecipientUserID: recipientUserID,
		Type:            notificationType,
		Title:           title,
		Body:            body,
		ReferenceType:   referenceType,
		ReferenceID:     referenceID,
	})
}

// Dispatch schedules in-app and provider delivery after the caller's main transaction commits.
// ponytail: in-process best-effort only; replace with a durable outbox when delivery reliability is required.
func (d *Dispatcher) Dispatch(job DispatchJob) {
	if d == nil || d.db == nil || d.service == nil ||
		strings.TrimSpace(job.OrganizationID) == "" ||
		strings.TrimSpace(job.RecipientUserID) == "" ||
		strings.TrimSpace(job.Type) == "" ||
		strings.TrimSpace(job.Title) == "" {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		var email, phone *string
		var inAppEnabled, emailEnabled, whatsappEnabled, dueReminderEnabled bool
		if err := d.db.QueryRow(ctx, `
			SELECT
				NULLIF(u.email, ''),
				NULLIF(u.phone, ''),
				COALESCE(np.in_app_enabled, true),
				COALESCE(np.email_enabled, true),
				COALESCE(np.whatsapp_enabled, false),
				COALESCE(np.due_reminder_enabled, true)
			FROM users u
			LEFT JOIN notification_preferences np
			  ON np.organization_id = u.organization_id
			 AND np.user_id = u.id
			WHERE u.organization_id = $1
			  AND u.id = $2
			  AND u.status IN ('active', 'invited')`,
			job.OrganizationID,
			job.RecipientUserID,
		).Scan(&email, &phone, &inAppEnabled, &emailEnabled, &whatsappEnabled, &dueReminderEnabled); err != nil {
			d.logFailure(ctx, "load_recipient", job, err)
			return
		}

		if job.IsDueReminder && !dueReminderEnabled {
			return
		}
		if inAppEnabled && d.allowsChannel(job, "in_app") {
			if err := d.createInApp(ctx, job); err != nil {
				d.logFailure(ctx, "create_in_app", job, err)
			}
		}
		if email != nil && emailEnabled && d.allowsChannel(job, "email") {
			if err := d.mailer.SendEmail(ctx, *email, job.Title, emailHTML(job.Title, job.Body)); err != nil {
				d.logFailure(ctx, "send_email", job, err)
			}
		}
		if phone != nil && whatsappEnabled && d.allowsChannel(job, "whatsapp") {
			if err := d.whatsapp.SendMessage(ctx, *phone, whatsappText(job.Title, job.Body)); err != nil {
				d.logFailure(ctx, "send_whatsapp", job, err)
			}
		}
	}()
}

func (d *Dispatcher) allowsChannel(job DispatchJob, channel string) bool {
	return len(job.Channels) == 0 || job.Channels[channel]
}

func (d *Dispatcher) createInApp(ctx context.Context, job DispatchJob) error {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin notification transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var body, referenceType, referenceID *string
	if job.Body != "" {
		body = &job.Body
	}
	if job.ReferenceType != "" {
		referenceType = &job.ReferenceType
	}
	if job.ReferenceID != "" {
		referenceID = &job.ReferenceID
	}
	if _, err := d.service.Create(ctx, tx, job.OrganizationID, CreateParams{
		UserID:        job.RecipientUserID,
		Type:          job.Type,
		Title:         job.Title,
		Body:          body,
		ReferenceType: referenceType,
		ReferenceID:   referenceID,
	}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit notification transaction: %w", err)
	}
	return nil
}

func (d *Dispatcher) logFailure(ctx context.Context, operation string, job DispatchJob, err error) {
	d.logger.WarnContext(ctx, "notification delivery failed",
		"operation", operation,
		"error", err,
		"organization_id", job.OrganizationID,
		"recipient_user_id", job.RecipientUserID,
		"type", job.Type,
	)
}

func emailHTML(title, body string) string {
	return "<h2>" + html.EscapeString(title) + "</h2><p>" +
		strings.ReplaceAll(html.EscapeString(body), "\n", "<br>") + "</p>"
}

func whatsappText(title, body string) string {
	if body == "" {
		return "*" + title + "*"
	}
	return "*" + title + "*\n\n" + body
}
