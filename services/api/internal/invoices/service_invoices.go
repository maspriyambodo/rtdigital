package invoices

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
	"github.com/maspriyambodo/rtdigital/services/api/internal/notifications"
)

func (s *Service) ListInvoices(ctx context.Context, principal *auth.Principal, filter InvoiceFilter) ([]Invoice, error) {
	filter.Status = strings.TrimSpace(filter.Status)
	filter.DueTypeID = strings.TrimSpace(filter.DueTypeID)
	filter.HouseholdID = strings.TrimSpace(filter.HouseholdID)
	filter.PeriodStart = strings.TrimSpace(filter.PeriodStart)
	filter.PeriodEnd = strings.TrimSpace(filter.PeriodEnd)

	if filter.Status != "" && !validInvoiceStatus(filter.Status) {
		return nil, fmt.Errorf("%w: status", ErrValidation)
	}
	if filter.PeriodStart != "" && !validDate(filter.PeriodStart) {
		return nil, fmt.Errorf("%w: period_start", ErrValidation)
	}
	if filter.PeriodEnd != "" && !validDate(filter.PeriodEnd) {
		return nil, fmt.Errorf("%w: period_end", ErrValidation)
	}

	residentHouseholdID, residentOnly, err := s.resolveResidentHouseholdScope(ctx, principal)
	if err != nil {
		return nil, err
	}
	if residentOnly {
		if residentHouseholdID == "" {
			return []Invoice{}, nil
		}
		filter.HouseholdID = residentHouseholdID
	}

	rows, err := s.db.Query(ctx, `
		SELECT i.id, i.household_id, h.internal_number, hu.code, i.due_type_id, dt.name,
		       i.invoice_number, i.period_start::text, i.period_end::text, i.due_date::text,
		       i.amount, i.paid_amount, i.adjustment_amount, i.adjustment_reason,
		       i.status, i.cancelled_by::text, i.cancelled_at, i.cancellation_reason,
		       i.created_at, i.updated_at
		FROM invoices i
		JOIN households h ON h.id = i.household_id AND h.organization_id = i.organization_id
		JOIN house_units hu ON hu.id = h.house_unit_id AND hu.organization_id = i.organization_id
		JOIN due_types dt ON dt.id = i.due_type_id AND dt.organization_id = i.organization_id
		WHERE i.organization_id = $1
		  AND ($2 = '' OR i.status = $2)
		  AND ($3 = '' OR i.due_type_id = $3::uuid)
		  AND ($4 = '' OR i.household_id = $4::uuid)
		  AND ($5 = '' OR i.period_start >= $5::date)
		  AND ($6 = '' OR i.period_end <= $6::date)
		  AND (NOT $7::boolean OR (i.status IN ('unpaid', 'pending_verification', 'partial') AND i.due_date < $8::date))
		ORDER BY i.due_date DESC, i.invoice_number DESC`,
		principal.OrganizationID, filter.Status, filter.DueTypeID, filter.HouseholdID,
		filter.PeriodStart, filter.PeriodEnd, filter.OnlyArrears, s.now().Format("2006-01-02"),
	)
	if err != nil {
		return nil, fmt.Errorf("list invoices: %w", err)
	}
	defer rows.Close()

	items := []Invoice{}
	for rows.Next() {
		item, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) GetInvoice(ctx context.Context, principal *auth.Principal, id string) (Invoice, error) {
	residentHouseholdID, residentOnly, err := s.resolveResidentHouseholdScope(ctx, principal)
	if err != nil {
		return Invoice{}, err
	}

	item, err := s.getInvoice(ctx, s.db, principal.OrganizationID, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Invoice{}, ErrInvoiceNotFound
	}
	if err != nil {
		return Invoice{}, err
	}
	if residentOnly && item.HouseholdID != residentHouseholdID {
		return Invoice{}, ErrInvoiceNotFound
	}
	return item, nil
}

func (s *Service) CreateInvoice(ctx context.Context, principal *auth.Principal, req CreateInvoiceRequest) (Invoice, error) {
	req.HouseholdID = strings.TrimSpace(req.HouseholdID)
	req.DueTypeID = strings.TrimSpace(req.DueTypeID)
	req.PeriodStart = strings.TrimSpace(req.PeriodStart)
	req.PeriodEnd = strings.TrimSpace(req.PeriodEnd)
	req.DueDate = strings.TrimSpace(req.DueDate)
	if req.HouseholdID == "" || req.DueTypeID == "" || !validDate(req.PeriodStart) || !validDate(req.PeriodEnd) || !validDate(req.DueDate) || req.PeriodEnd < req.PeriodStart {
		return Invoice{}, fmt.Errorf("%w: invoice fields", ErrValidation)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Invoice{}, fmt.Errorf("begin create invoice: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	amount, err := s.resolveInvoiceAmount(ctx, tx, principal.OrganizationID, req.DueTypeID, req.Amount)
	if err != nil {
		return Invoice{}, err
	}
	adjustmentAmount, adjustmentReason, err := validateAdjustment(req.AdjustmentAmount, req.AdjustmentReason, amount)
	if err != nil {
		return Invoice{}, err
	}
	if err := s.assertHousehold(ctx, tx, principal.OrganizationID, req.HouseholdID); err != nil {
		return Invoice{}, err
	}

	item, err := s.insertInvoice(ctx, tx, principal, req.HouseholdID, req.DueTypeID, req.PeriodStart, req.PeriodEnd, req.DueDate, amount, adjustmentAmount, adjustmentReason, nil)
	if err != nil {
		return Invoice{}, err
	}
	if err := s.audit(ctx, tx, principal, "invoice.create", "invoices", item.ID); err != nil {
		return Invoice{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Invoice{}, fmt.Errorf("commit create invoice: %w", err)
	}
	s.notifyHousehold(ctx, principal.OrganizationID, req.HouseholdID, notifications.DispatchJob{
		Type:          "invoice_created",
		Title:         "Tagihan iuran baru",
		Body:          fmt.Sprintf("Tagihan %s diterbitkan jatuh tempo %s.", item.InvoiceNumber, item.DueDate),
		ReferenceType: "invoice",
		ReferenceID:   item.ID,
	})
	return s.GetInvoice(ctx, principal, item.ID)
}

func (s *Service) GenerateInvoices(ctx context.Context, principal *auth.Principal, idempotencyKey string, req GenerateInvoicesRequest) (GenerateInvoicesResult, error) {
	req.DueTypeID = strings.TrimSpace(req.DueTypeID)
	req.PeriodStart = strings.TrimSpace(req.PeriodStart)
	req.PeriodEnd = strings.TrimSpace(req.PeriodEnd)
	req.DueDate = strings.TrimSpace(req.DueDate)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" || req.DueTypeID == "" || !validDate(req.PeriodStart) || !validDate(req.PeriodEnd) || !validDate(req.DueDate) || req.PeriodEnd < req.PeriodStart {
		return GenerateInvoicesResult{}, fmt.Errorf("%w: idempotency key or invoice fields", ErrValidation)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return GenerateInvoicesResult{}, fmt.Errorf("begin bulk generation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, principal.OrganizationID+":"+idempotencyKey); err != nil {
		return GenerateInvoicesResult{}, fmt.Errorf("lock bulk generation: %w", err)
	}
	existing, err := s.getBulkGeneratedInvoices(ctx, tx, principal.OrganizationID, idempotencyKey)
	if err != nil {
		return GenerateInvoicesResult{}, err
	}
	if len(existing) > 0 {
		if err := tx.Commit(ctx); err != nil {
			return GenerateInvoicesResult{}, fmt.Errorf("commit existing bulk result: %w", err)
		}
		return GenerateInvoicesResult{TotalTargeted: len(existing), TotalCreated: len(existing), Invoices: existing}, nil
	}

	amount, err := s.resolveInvoiceAmount(ctx, tx, principal.OrganizationID, req.DueTypeID, req.Amount)
	if err != nil {
		return GenerateInvoicesResult{}, err
	}
	adjustmentAmount, adjustmentReason, err := validateAdjustment(req.AdjustmentAmount, req.AdjustmentReason, amount)
	if err != nil {
		return GenerateInvoicesResult{}, err
	}
	households, err := s.getTargetHouseholds(ctx, tx, principal.OrganizationID, req.HouseholdIDs)
	if err != nil {
		return GenerateInvoicesResult{}, err
	}
	if len(households) == 0 {
		return GenerateInvoicesResult{}, fmt.Errorf("%w: no target households", ErrValidation)
	}

	result := GenerateInvoicesResult{TotalTargeted: len(households), Invoices: make([]Invoice, 0, len(households))}
	for _, household := range households {
		var exists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM invoices
				WHERE organization_id = $1 AND household_id = $2 AND due_type_id = $3
				  AND period_start = $4::date AND period_end = $5::date AND status <> 'cancelled'
			)`, principal.OrganizationID, household.ID, req.DueTypeID, req.PeriodStart, req.PeriodEnd,
		).Scan(&exists); err != nil {
			return GenerateInvoicesResult{}, fmt.Errorf("check duplicate invoice: %w", err)
		}
		if exists {
			result.TotalSkipped++
			continue
		}
		key := idempotencyKey
		item, err := s.insertInvoice(ctx, tx, principal, household.ID, req.DueTypeID, req.PeriodStart, req.PeriodEnd, req.DueDate, amount, adjustmentAmount, adjustmentReason, &key)
		if err != nil {
			return GenerateInvoicesResult{}, err
		}
		item.HouseholdNumber, item.HouseUnitCode, item.DueTypeName = household.InternalNumber, household.HouseUnitCode, household.DueTypeName
		result.Invoices = append(result.Invoices, item)
		result.TotalCreated++
		if err := s.audit(ctx, tx, principal, "invoice.generate", "invoices", item.ID); err != nil {
			return GenerateInvoicesResult{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return GenerateInvoicesResult{}, fmt.Errorf("commit bulk generation: %w", err)
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
	return result, nil
}

func (s *Service) UpdateInvoice(ctx context.Context, principal *auth.Principal, id string, req UpdateInvoiceRequest) (Invoice, error) {
	if req.Amount == nil && req.AdjustmentAmount == nil && req.AdjustmentReason == nil && req.DueDate == nil {
		return Invoice{}, fmt.Errorf("%w: no changes", ErrValidation)
	}
	if req.DueDate != nil && !validDate(*req.DueDate) {
		return Invoice{}, fmt.Errorf("%w: due_date", ErrValidation)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Invoice{}, fmt.Errorf("begin update invoice: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var current Invoice
	err = tx.QueryRow(ctx, `
		SELECT id, amount, paid_amount, adjustment_amount, adjustment_reason, status, due_date::text
		FROM invoices WHERE id = $1 AND organization_id = $2 FOR UPDATE`,
		id, principal.OrganizationID,
	).Scan(&current.ID, &current.Amount, &current.PaidAmount, &current.AdjustmentAmount, &current.AdjustmentReason, &current.Status, &current.DueDate)
	if errors.Is(err, pgx.ErrNoRows) {
		return Invoice{}, ErrInvoiceNotFound
	}
	if err != nil {
		return Invoice{}, fmt.Errorf("lock invoice: %w", err)
	}
	if current.PaidAmount > 0 || current.Status != "unpaid" {
		return Invoice{}, ErrInvalidInvoiceStatus
	}

	amount := current.Amount
	if req.Amount != nil {
		if *req.Amount <= 0 {
			return Invoice{}, fmt.Errorf("%w: amount", ErrValidation)
		}
		amount = *req.Amount
	}
	adjustment := current.AdjustmentAmount
	reason := current.AdjustmentReason
	if req.AdjustmentAmount != nil || req.AdjustmentReason != nil {
		value := req.AdjustmentAmount
		if value == nil {
			value = &current.AdjustmentAmount
		}
		inputReason := req.AdjustmentReason
		if inputReason == nil {
			inputReason = current.AdjustmentReason
		}
		adjustment, reason, err = validateAdjustment(value, inputReason, amount)
		if err != nil {
			return Invoice{}, err
		}
	}
	dueDate := current.DueDate
	if req.DueDate != nil {
		dueDate = *req.DueDate
	}
	if _, err := tx.Exec(ctx, `
		UPDATE invoices SET amount = $1, adjustment_amount = $2, adjustment_reason = $3, due_date = $4::date
		WHERE id = $5 AND organization_id = $6`, amount, adjustment, reason, dueDate, id, principal.OrganizationID); err != nil {
		return Invoice{}, mapDatabaseError(err, "update invoice")
	}
	if err := s.audit(ctx, tx, principal, "invoice.update", "invoices", id); err != nil {
		return Invoice{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Invoice{}, fmt.Errorf("commit update invoice: %w", err)
	}
	return s.GetInvoice(ctx, principal, id)
}

func (s *Service) CancelInvoice(ctx context.Context, principal *auth.Principal, id string, req CancelInvoiceRequest) (Invoice, error) {
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" {
		return Invoice{}, fmt.Errorf("%w: reason", ErrValidation)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Invoice{}, fmt.Errorf("begin cancel invoice: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var amount float64
	var status string
	if err := tx.QueryRow(ctx, `
		SELECT paid_amount, status FROM invoices
		WHERE id = $1 AND organization_id = $2 FOR UPDATE`, id, principal.OrganizationID,
	).Scan(&amount, &status); errors.Is(err, pgx.ErrNoRows) {
		return Invoice{}, ErrInvoiceNotFound
	} else if err != nil {
		return Invoice{}, fmt.Errorf("lock invoice: %w", err)
	}
	if amount > 0 || status != "unpaid" {
		return Invoice{}, ErrInvalidInvoiceStatus
	}
	if _, err := tx.Exec(ctx, `
		UPDATE invoices
		SET status = 'cancelled', cancelled_by = $1, cancelled_at = $2, cancellation_reason = $3
		WHERE id = $4 AND organization_id = $5`, principal.UserID, s.now(), req.Reason, id, principal.OrganizationID); err != nil {
		return Invoice{}, mapDatabaseError(err, "cancel invoice")
	}
	if err := s.audit(ctx, tx, principal, "invoice.cancel", "invoices", id); err != nil {
		return Invoice{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Invoice{}, fmt.Errorf("commit cancel invoice: %w", err)
	}
	return s.GetInvoice(ctx, principal, id)
}

type invoiceScanner interface {
	Scan(...any) error
}

func scanInvoice(row invoiceScanner) (Invoice, error) {
	var item Invoice
	err := row.Scan(
		&item.ID, &item.HouseholdID, &item.HouseholdNumber, &item.HouseUnitCode,
		&item.DueTypeID, &item.DueTypeName, &item.InvoiceNumber, &item.PeriodStart,
		&item.PeriodEnd, &item.DueDate, &item.Amount, &item.PaidAmount,
		&item.AdjustmentAmount, &item.AdjustmentReason, &item.Status,
		&item.CancelledBy, &item.CancelledAt, &item.CancellationReason,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return Invoice{}, fmt.Errorf("scan invoice: %w", err)
	}
	return item, nil
}

func (s *Service) getInvoice(ctx context.Context, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, organizationID, id string) (Invoice, error) {
	return scanInvoice(db.QueryRow(ctx, `
		SELECT i.id, i.household_id, h.internal_number, hu.code, i.due_type_id, dt.name,
		       i.invoice_number, i.period_start::text, i.period_end::text, i.due_date::text,
		       i.amount, i.paid_amount, i.adjustment_amount, i.adjustment_reason,
		       i.status, i.cancelled_by::text, i.cancelled_at, i.cancellation_reason,
		       i.created_at, i.updated_at
		FROM invoices i
		JOIN households h ON h.id = i.household_id AND h.organization_id = i.organization_id
		JOIN house_units hu ON hu.id = h.house_unit_id AND hu.organization_id = i.organization_id
		JOIN due_types dt ON dt.id = i.due_type_id AND dt.organization_id = i.organization_id
		WHERE i.id = $1 AND i.organization_id = $2`, id, organizationID))
}

func (s *Service) insertInvoice(ctx context.Context, tx pgx.Tx, principal *auth.Principal, householdID, dueTypeID, periodStart, periodEnd, dueDate string, amount, adjustmentAmount float64, adjustmentReason, bulkKey *string) (Invoice, error) {
	var item Invoice
	if err := tx.QueryRow(ctx, `
		SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`,
		principal.OrganizationID+":invoice-number",
	).Scan(new(any)); err != nil {
		return Invoice{}, fmt.Errorf("lock invoice sequence: %w", err)
	}
	prefix := fmt.Sprintf("INV-%s-", s.now().Format("0601"))
	var count int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM invoices
		WHERE organization_id = $1 AND invoice_number LIKE $2 || '%'`,
		principal.OrganizationID, prefix,
	).Scan(&count); err != nil {
		return Invoice{}, fmt.Errorf("count invoice sequence: %w", err)
	}
	number := fmt.Sprintf("%s%04d", prefix, count+1)
	row := tx.QueryRow(ctx, `
		INSERT INTO invoices (
			id, organization_id, household_id, due_type_id, invoice_number, period_start, period_end,
			due_date, amount, paid_amount, adjustment_amount, adjustment_reason, status, bulk_generation_key
		) VALUES ($1, $2, $3, $4, $5, $6::date, $7::date, $8::date, $9, 0, $10, $11, 'unpaid', $12)
		RETURNING id, household_id, ''::text, ''::text, due_type_id, ''::text, invoice_number,
		          period_start::text, period_end::text, due_date::text, amount, paid_amount,
		          adjustment_amount, adjustment_reason, status, cancelled_by::text, cancelled_at,
		          cancellation_reason, created_at, updated_at`,
		newUUID(), principal.OrganizationID, householdID, dueTypeID, number, periodStart, periodEnd,
		dueDate, amount, adjustmentAmount, adjustmentReason, bulkKey,
	)
	item, err := scanInvoice(row)
	if err != nil {
		return Invoice{}, mapDatabaseError(err, "insert invoice")
	}
	return item, nil
}

func (s *Service) assertHousehold(ctx context.Context, tx pgx.Tx, organizationID, householdID string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM households WHERE id = $1 AND organization_id = $2)`,
		householdID, organizationID,
	).Scan(&exists); err != nil {
		return fmt.Errorf("check household: %w", err)
	}
	if !exists {
		return ErrHouseholdNotFound
	}
	return nil
}

type targetHousehold struct {
	ID             string
	InternalNumber string
	HouseUnitCode  string
	DueTypeName    string
}

func (s *Service) getTargetHouseholds(ctx context.Context, tx pgx.Tx, organizationID string, selectedIDs []string) ([]targetHousehold, error) {
	cleaned := make([]string, 0, len(selectedIDs))
	seen := map[string]struct{}{}
	for _, id := range selectedIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		cleaned = append(cleaned, id)
	}

	query := `
		SELECT h.id, h.internal_number, hu.code
		FROM households h
		JOIN house_units hu ON hu.id = h.house_unit_id AND hu.organization_id = h.organization_id
		WHERE h.organization_id = $1
		  AND EXISTS (
			SELECT 1
			FROM household_members hm
			WHERE hm.organization_id = h.organization_id
			  AND hm.household_id = h.id
			  AND hm.is_active
		  )`
	args := []any{organizationID}
	if len(selectedIDs) > 0 {
		if len(cleaned) == 0 {
			return nil, fmt.Errorf("%w: household_ids", ErrValidation)
		}
		query += ` AND h.id = ANY($2::uuid[])`
		args = append(args, cleaned)
	}
	query += ` ORDER BY hu.code, h.internal_number`

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get target households: %w", err)
	}
	defer rows.Close()

	items := []targetHousehold{}
	for rows.Next() {
		var item targetHousehold
		if err := rows.Scan(&item.ID, &item.InternalNumber, &item.HouseUnitCode); err != nil {
			return nil, fmt.Errorf("scan target household: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) getBulkGeneratedInvoices(ctx context.Context, db interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, organizationID, key string) ([]Invoice, error) {
	rows, err := db.Query(ctx, `
		SELECT i.id, i.household_id, h.internal_number, hu.code, i.due_type_id, dt.name,
		       i.invoice_number, i.period_start::text, i.period_end::text, i.due_date::text,
		       i.amount, i.paid_amount, i.adjustment_amount, i.adjustment_reason,
		       i.status, i.cancelled_by::text, i.cancelled_at, i.cancellation_reason,
		       i.created_at, i.updated_at
		FROM invoices i
		JOIN households h ON h.id = i.household_id AND h.organization_id = i.organization_id
		JOIN house_units hu ON hu.id = h.house_unit_id AND hu.organization_id = i.organization_id
		JOIN due_types dt ON dt.id = i.due_type_id AND dt.organization_id = i.organization_id
		WHERE i.organization_id = $1 AND i.bulk_generation_key = $2
		ORDER BY i.invoice_number`, organizationID, key)
	if err != nil {
		return nil, fmt.Errorf("find bulk invoices: %w", err)
	}
	defer rows.Close()

	items := []Invoice{}
	for rows.Next() {
		item, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) resolveResidentHouseholdScope(ctx context.Context, principal *auth.Principal) (string, bool, error) {
	if principal.HasPermission("invoice.read_all") || principal.HasPermission("invoice.create") {
		return "", false, nil
	}
	var residentID *string
	err := s.db.QueryRow(ctx, `
		SELECT resident_id FROM users WHERE id = $1 AND organization_id = $2`,
		principal.UserID, principal.OrganizationID,
	).Scan(&residentID)
	if errors.Is(err, pgx.ErrNoRows) || residentID == nil {
		return "", true, nil
	}
	if err != nil {
		return "", true, fmt.Errorf("resolve resident scope: %w", err)
	}
	var householdID string
	err = s.db.QueryRow(ctx, `
		SELECT household_id FROM household_members
		WHERE organization_id = $1 AND resident_id = $2 AND is_active
		LIMIT 1`, principal.OrganizationID, *residentID,
	).Scan(&householdID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", true, nil
	}
	if err != nil {
		return "", true, fmt.Errorf("resolve household scope: %w", err)
	}
	return householdID, true, nil
}

func (s *Service) resolveInvoiceAmount(ctx context.Context, tx pgx.Tx, organizationID, dueTypeID string, customAmount *float64) (float64, error) {
	var defaultAmount *float64
	var status string
	err := tx.QueryRow(ctx, `
		SELECT amount, status FROM due_types
		WHERE id = $1 AND organization_id = $2`, dueTypeID, organizationID,
	).Scan(&defaultAmount, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrDueTypeNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("get due type: %w", err)
	}
	if status != "active" {
		return 0, fmt.Errorf("%w: inactive due type", ErrConstraint)
	}
	if customAmount != nil {
		if *customAmount <= 0 {
			return 0, fmt.Errorf("%w: amount", ErrValidation)
		}
		return *customAmount, nil
	}
	if defaultAmount == nil || *defaultAmount <= 0 {
		return 0, fmt.Errorf("%w: amount required", ErrValidation)
	}
	return *defaultAmount, nil
}

func validateAdjustment(amount *float64, reason *string, invoiceAmount float64) (float64, *string, error) {
	if amount == nil || *amount == 0 {
		return 0, nil, nil
	}
	if *amount < 0 || *amount >= invoiceAmount {
		return 0, nil, fmt.Errorf("%w: adjustment_amount", ErrValidation)
	}
	reason = nullableTrim(reason)
	if reason == nil {
		return 0, nil, fmt.Errorf("%w: adjustment_reason", ErrValidation)
	}
	return *amount, reason, nil
}

func validInvoiceStatus(value string) bool {
	switch value {
	case "unpaid", "pending_verification", "partial", "paid", "cancelled":
		return true
	}
	return false
}
