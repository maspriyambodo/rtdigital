package invoices

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
	"github.com/maspriyambodo/rtdigital/services/api/internal/notifications"
)

type CreateInvoiceGenerationRunRequest struct {
	DueTypeID   string `json:"due_type_id"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
	DueDate     string `json:"due_date"`
}

// CreateInvoiceGenerationRun is safe for an external scheduler to retry.
// The unique run key plus transaction advisory lock prevent duplicate invoices.
func (s *Service) CreateInvoiceGenerationRun(ctx context.Context, principal *auth.Principal, runKey string, req CreateInvoiceGenerationRunRequest) (InvoiceGenerationRun, error) {
	runKey = strings.TrimSpace(runKey)
	req.DueTypeID = strings.TrimSpace(req.DueTypeID)
	req.PeriodStart = strings.TrimSpace(req.PeriodStart)
	req.PeriodEnd = strings.TrimSpace(req.PeriodEnd)
	req.DueDate = strings.TrimSpace(req.DueDate)
	if runKey == "" || req.DueTypeID == "" || !validDate(req.PeriodStart) || !validDate(req.PeriodEnd) ||
		!validDate(req.DueDate) || req.PeriodEnd < req.PeriodStart {
		return InvoiceGenerationRun{}, fmt.Errorf("%w: run key or generation fields", ErrValidation)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return InvoiceGenerationRun{}, fmt.Errorf("begin invoice generation run: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`,
		principal.OrganizationID+":invoice-generation:"+runKey,
	); err != nil {
		return InvoiceGenerationRun{}, fmt.Errorf("lock invoice generation run: %w", err)
	}

	existing, err := scanInvoiceGenerationRun(tx.QueryRow(ctx, `
		SELECT id, due_type_id, period_start::text, period_end::text, due_date::text, run_key, status,
		       total_targeted, total_created, total_skipped, error_message, started_at, completed_at, created_by::text
		FROM invoice_generation_runs
		WHERE organization_id = $1 AND run_key = $2`,
		principal.OrganizationID, runKey,
	))
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return InvoiceGenerationRun{}, fmt.Errorf("commit existing generation run: %w", err)
		}
		return existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return InvoiceGenerationRun{}, err
	}

	var amount *float64
	var frequency, status string
	var automaticGenerationEnabled bool
	err = tx.QueryRow(ctx, `
		SELECT amount, frequency, status, automatic_generation_enabled
		FROM due_types
		WHERE organization_id = $1 AND id = $2
		FOR UPDATE`,
		principal.OrganizationID, req.DueTypeID,
	).Scan(&amount, &frequency, &status, &automaticGenerationEnabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return InvoiceGenerationRun{}, ErrDueTypeNotFound
	}
	if err != nil {
		return InvoiceGenerationRun{}, fmt.Errorf("lock due type: %w", err)
	}
	if status != "active" || amount == nil || *amount <= 0 || frequency == "once" || !automaticGenerationEnabled {
		return InvoiceGenerationRun{}, fmt.Errorf("%w: due type is not eligible for routine generation", ErrConstraint)
	}

	runID := newUUID()
	var createdBy *string
	if principal.UserID != "" {
		createdBy = &principal.UserID
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO invoice_generation_runs (
			id, organization_id, due_type_id, period_start, period_end, due_date, run_key, status, created_by
		) VALUES ($1, $2, $3, $4::date, $5::date, $6::date, $7, 'running', $8)`,
		runID, principal.OrganizationID, req.DueTypeID, req.PeriodStart, req.PeriodEnd, req.DueDate, runKey, createdBy,
	); err != nil {
		return InvoiceGenerationRun{}, mapDatabaseError(err, "create invoice generation run")
	}

	households, err := s.getTargetHouseholds(ctx, tx, principal.OrganizationID, nil)
	if err != nil {
		return InvoiceGenerationRun{}, err
	}
	result := GenerateInvoicesResult{TotalTargeted: len(households)}
	for _, household := range households {
		var exists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM invoices
				WHERE organization_id = $1 AND household_id = $2 AND due_type_id = $3
				  AND period_start = $4::date AND period_end = $5::date AND status <> 'cancelled'
			)`,
			principal.OrganizationID, household.ID, req.DueTypeID, req.PeriodStart, req.PeriodEnd,
		).Scan(&exists); err != nil {
			return InvoiceGenerationRun{}, fmt.Errorf("check routine invoice: %w", err)
		}
		if exists {
			result.TotalSkipped++
			continue
		}

		key := runKey
		item, err := s.insertInvoice(ctx, tx, principal, household.ID, req.DueTypeID, req.PeriodStart, req.PeriodEnd, req.DueDate, *amount, 0, nil, &key)
		if err != nil {
			return InvoiceGenerationRun{}, err
		}
		result.Invoices = append(result.Invoices, item)
		result.TotalCreated++
		if err := s.audit(ctx, tx, principal, "invoice.generate_routine", "invoices", item.ID); err != nil {
			return InvoiceGenerationRun{}, err
		}
	}

	now := s.now()
	if _, err := tx.Exec(ctx, `
		UPDATE invoice_generation_runs
		SET status = 'completed', total_targeted = $1, total_created = $2, total_skipped = $3, completed_at = $4
		WHERE organization_id = $5 AND id = $6`,
		result.TotalTargeted, result.TotalCreated, result.TotalSkipped, now, principal.OrganizationID, runID,
	); err != nil {
		return InvoiceGenerationRun{}, fmt.Errorf("complete invoice generation run: %w", err)
	}
	if err := s.audit(ctx, tx, principal, "invoice.generation_run.complete", "invoice_generation_runs", runID); err != nil {
		return InvoiceGenerationRun{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return InvoiceGenerationRun{}, fmt.Errorf("commit invoice generation run: %w", err)
	}

	for _, invoice := range result.Invoices {
		s.notifyHousehold(ctx, principal.OrganizationID, invoice.HouseholdID, notifications.DispatchJob{
			Type:          "invoice_created",
			Title:         "Tagihan iuran baru",
			Body:          fmt.Sprintf("Tagihan %s diterbitkan jatuh tempo %s.", invoice.InvoiceNumber, invoice.DueDate),
			ReferenceType: "invoice",
			ReferenceID:   invoice.ID,
		})
	}
	return s.GetInvoiceGenerationRun(ctx, principal, runID)
}

func (s *Service) ListInvoiceGenerationRuns(ctx context.Context, principal *auth.Principal) ([]InvoiceGenerationRun, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, due_type_id, period_start::text, period_end::text, due_date::text, run_key, status,
		       total_targeted, total_created, total_skipped, error_message, started_at, completed_at, created_by::text
		FROM invoice_generation_runs
		WHERE organization_id = $1
		ORDER BY started_at DESC, id DESC`,
		principal.OrganizationID,
	)
	if err != nil {
		return nil, fmt.Errorf("list invoice generation runs: %w", err)
	}
	defer rows.Close()

	items := []InvoiceGenerationRun{}
	for rows.Next() {
		item, err := scanInvoiceGenerationRun(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) GetInvoiceGenerationRun(ctx context.Context, principal *auth.Principal, id string) (InvoiceGenerationRun, error) {
	item, err := scanInvoiceGenerationRun(s.db.QueryRow(ctx, `
		SELECT id, due_type_id, period_start::text, period_end::text, due_date::text, run_key, status,
		       total_targeted, total_created, total_skipped, error_message, started_at, completed_at, created_by::text
		FROM invoice_generation_runs
		WHERE organization_id = $1 AND id = $2`,
		principal.OrganizationID, id,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return InvoiceGenerationRun{}, ErrInvoiceNotFound
	}
	if err != nil {
		return InvoiceGenerationRun{}, fmt.Errorf("get invoice generation run: %w", err)
	}
	return item, nil
}

type invoiceGenerationRunScanner interface {
	Scan(...any) error
}

func scanInvoiceGenerationRun(row invoiceGenerationRunScanner) (InvoiceGenerationRun, error) {
	var item InvoiceGenerationRun
	err := row.Scan(
		&item.ID, &item.DueTypeID, &item.PeriodStart, &item.PeriodEnd, &item.DueDate,
		&item.RunKey, &item.Status, &item.TotalTargeted, &item.TotalCreated, &item.TotalSkipped,
		&item.ErrorMessage, &item.StartedAt, &item.CompletedAt, &item.CreatedBy,
	)
	return item, err
}

// ponytail: external scheduler calls the authenticated endpoint; add a dedicated
// service credential only when scheduler infrastructure is provisioned.
var _ = time.Time{}
