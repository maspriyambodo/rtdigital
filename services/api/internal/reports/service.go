package reports

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) Residents(ctx context.Context, principal *auth.Principal, filter Filter) ([]ResidentReportItem, error) {
	if principal == nil {
		return nil, ErrForbidden
	}
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	conditions := []string{"r.organization_id = $1"}
	args := []any{principal.OrganizationID}
	index := 2
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("r.resident_status = $%d", index))
		args, index = append(args, filter.Status), index+1
	}
	if filter.StartDate != "" {
		conditions = append(conditions, fmt.Sprintf("r.created_at >= $%d::date", index))
		args, index = append(args, filter.StartDate), index+1
	}
	if filter.EndDate != "" {
		conditions = append(conditions, fmt.Sprintf("r.created_at < ($%d::date + interval '1 day')", index))
		args = append(args, filter.EndDate)
	}

	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT r.id, r.full_name, r.gender, r.birth_date, r.resident_status, r.verification_status,
		       hu.code, h.internal_number, hm.relationship, r.created_at
		FROM residents r
		LEFT JOIN household_members hm ON hm.resident_id = r.id AND hm.is_active
		LEFT JOIN households h ON h.id = hm.household_id AND h.organization_id = r.organization_id
		LEFT JOIN house_units hu ON hu.id = h.house_unit_id AND hu.organization_id = r.organization_id
		WHERE %s
		ORDER BY r.full_name ASC`, strings.Join(conditions, " AND ")), args...)
	if err != nil {
		return nil, fmt.Errorf("query residents report: %w", err)
	}
	defer rows.Close()

	items := []ResidentReportItem{}
	for rows.Next() {
		var item ResidentReportItem
		var birthDate *time.Time
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.FullName, &item.Gender, &birthDate, &item.ResidentStatus, &item.VerificationStatus, &item.HouseUnitCode, &item.HouseholdNumber, &item.Relationship, &createdAt); err != nil {
			return nil, fmt.Errorf("scan resident report: %w", err)
		}
		if birthDate != nil {
			value := birthDate.Format(time.DateOnly)
			item.BirthDate = &value
		}
		item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Households(ctx context.Context, principal *auth.Principal, filter Filter) ([]HouseholdReportItem, error) {
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
		conditions = append(conditions, fmt.Sprintf("h.verification_status = $%d", index))
		args = append(args, filter.Status)
	}
	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT h.id, h.internal_number, hu.code, COALESCE(hr.full_name, ''),
		       h.domicile_status, h.verification_status,
		       (SELECT COUNT(*) FROM household_members WHERE household_id = h.id AND is_active),
		       COALESCE(h.move_in_date::text, '')
		FROM households h
		JOIN house_units hu ON hu.id = h.house_unit_id AND hu.organization_id = h.organization_id
		LEFT JOIN residents hr ON hr.id = h.head_resident_id AND hr.organization_id = h.organization_id
		WHERE %s
		ORDER BY hu.code ASC, h.internal_number ASC`, strings.Join(conditions, " AND ")), args...)
	if err != nil {
		return nil, fmt.Errorf("query households report: %w", err)
	}
	defer rows.Close()

	items := []HouseholdReportItem{}
	for rows.Next() {
		var item HouseholdReportItem
		if err := rows.Scan(&item.ID, &item.InternalNumber, &item.HouseUnitCode, &item.HeadResidentName, &item.DomicileStatus, &item.VerificationStatus, &item.ActiveMembersCount, &item.MoveInDate); err != nil {
			return nil, fmt.Errorf("scan household report: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Invoices(ctx context.Context, principal *auth.Principal, filter Filter) ([]InvoiceReportItem, error) {
	if principal == nil {
		return nil, ErrForbidden
	}
	if err := filter.Validate(); err != nil {
		return nil, err
	}
	conditions, args, index := []string{"i.organization_id = $1"}, []any{principal.OrganizationID}, 2
	if filter.Status != "" {
		conditions, args, index = append(conditions, fmt.Sprintf("i.status = $%d", index)), append(args, filter.Status), index+1
	}
	if filter.StartDate != "" {
		conditions, args, index = append(conditions, fmt.Sprintf("i.due_date >= $%d", index)), append(args, filter.StartDate), index+1
	}
	if filter.EndDate != "" {
		conditions, args = append(conditions, fmt.Sprintf("i.due_date <= $%d", index)), append(args, filter.EndDate)
	}
	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT i.id, i.invoice_number, h.internal_number, hu.code, dt.name, i.period_start, i.period_end, i.due_date, i.amount, i.paid_amount, i.adjustment_amount, i.status
		FROM invoices i
		JOIN households h ON h.id = i.household_id AND h.organization_id = i.organization_id
		JOIN house_units hu ON hu.id = h.house_unit_id AND hu.organization_id = i.organization_id
		JOIN due_types dt ON dt.id = i.due_type_id AND dt.organization_id = i.organization_id
		WHERE %s ORDER BY i.due_date DESC, i.invoice_number ASC`, strings.Join(conditions, " AND ")), args...)
	if err != nil {
		return nil, fmt.Errorf("query invoices report: %w", err)
	}
	defer rows.Close()
	items := []InvoiceReportItem{}
	for rows.Next() {
		var item InvoiceReportItem
		var start, end, due time.Time
		if err := rows.Scan(&item.ID, &item.InvoiceNumber, &item.HouseholdNumber, &item.HouseUnitCode, &item.DueTypeName, &start, &end, &due, &item.Amount, &item.PaidAmount, &item.AdjustmentAmount, &item.Status); err != nil {
			return nil, fmt.Errorf("scan invoice report: %w", err)
		}
		item.PeriodStart, item.PeriodEnd, item.DueDate = start.Format(time.DateOnly), end.Format(time.DateOnly), due.Format(time.DateOnly)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Payments(ctx context.Context, principal *auth.Principal, filter Filter) ([]PaymentReportItem, error) {
	if principal == nil {
		return nil, ErrForbidden
	}
	if err := filter.Validate(); err != nil {
		return nil, err
	}
	conditions, args, index := []string{"p.organization_id = $1"}, []any{principal.OrganizationID}, 2
	if filter.Status != "" {
		conditions, args, index = append(conditions, fmt.Sprintf("p.verification_status = $%d", index)), append(args, filter.Status), index+1
	}
	if filter.StartDate != "" {
		conditions, args, index = append(conditions, fmt.Sprintf("p.paid_at >= $%d::date", index)), append(args, filter.StartDate), index+1
	}
	if filter.EndDate != "" {
		conditions, args = append(conditions, fmt.Sprintf("p.paid_at < ($%d::date + interval '1 day')", index)), append(args, filter.EndDate)
	}
	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT p.id, p.payment_number, i.invoice_number, h.internal_number, p.method, p.amount, p.paid_at, p.verification_status, p.verified_at
		FROM payments p
		JOIN invoices i ON i.id = p.invoice_id AND i.organization_id = p.organization_id
		JOIN households h ON h.id = i.household_id AND h.organization_id = p.organization_id
		WHERE %s ORDER BY p.paid_at DESC`, strings.Join(conditions, " AND ")), args...)
	if err != nil {
		return nil, fmt.Errorf("query payments report: %w", err)
	}
	defer rows.Close()
	items := []PaymentReportItem{}
	for rows.Next() {
		var item PaymentReportItem
		var paidAt time.Time
		var verifiedAt *time.Time
		if err := rows.Scan(&item.ID, &item.PaymentNumber, &item.InvoiceNumber, &item.HouseholdNumber, &item.Method, &item.Amount, &paidAt, &item.VerificationStatus, &verifiedAt); err != nil {
			return nil, fmt.Errorf("scan payment report: %w", err)
		}
		item.PaidAt = paidAt.UTC().Format(time.RFC3339)
		if verifiedAt != nil {
			value := verifiedAt.UTC().Format(time.RFC3339)
			item.VerifiedAt = &value
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Cash(ctx context.Context, principal *auth.Principal, filter Filter) ([]CashReportItem, error) {
	if principal == nil {
		return nil, ErrForbidden
	}
	if err := filter.Validate(); err != nil {
		return nil, err
	}
	conditions, args, index := []string{"t.organization_id = $1"}, []any{principal.OrganizationID}, 2
	if filter.Status != "" {
		conditions, args, index = append(conditions, fmt.Sprintf("t.status = $%d", index)), append(args, filter.Status), index+1
	}
	if filter.StartDate != "" {
		conditions, args, index = append(conditions, fmt.Sprintf("t.transaction_date >= $%d", index)), append(args, filter.StartDate), index+1
	}
	if filter.EndDate != "" {
		conditions, args = append(conditions, fmt.Sprintf("t.transaction_date <= $%d", index)), append(args, filter.EndDate)
	}
	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT t.id, t.transaction_number, t.type, c.name, t.amount, t.transaction_date, t.description, t.status
		FROM cash_transactions t LEFT JOIN cash_categories c ON c.id = t.category_id AND c.organization_id = t.organization_id
		WHERE %s ORDER BY t.transaction_date DESC, t.created_at DESC`, strings.Join(conditions, " AND ")), args...)
	if err != nil {
		return nil, fmt.Errorf("query cash report: %w", err)
	}
	defer rows.Close()
	items := []CashReportItem{}
	for rows.Next() {
		var item CashReportItem
		var date time.Time
		if err := rows.Scan(&item.ID, &item.TransactionNumber, &item.Type, &item.CategoryName, &item.Amount, &date, &item.Description, &item.Status); err != nil {
			return nil, fmt.Errorf("scan cash report: %w", err)
		}
		item.TransactionDate = date.Format(time.DateOnly)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Letters(ctx context.Context, principal *auth.Principal, filter Filter) ([]LetterReportItem, error) {
	if principal == nil {
		return nil, ErrForbidden
	}
	if err := filter.Validate(); err != nil {
		return nil, err
	}
	conditions, args, index := []string{"lr.organization_id = $1"}, []any{principal.OrganizationID}, 2
	if filter.Status != "" {
		conditions, args, index = append(conditions, fmt.Sprintf("lr.status = $%d", index)), append(args, filter.Status), index+1
	}
	if filter.StartDate != "" {
		conditions, args, index = append(conditions, fmt.Sprintf("lr.created_at >= $%d::date", index)), append(args, filter.StartDate), index+1
	}
	if filter.EndDate != "" {
		conditions, args = append(conditions, fmt.Sprintf("lr.created_at < ($%d::date + interval '1 day')", index)), append(args, filter.EndDate)
	}
	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT lr.id, lr.request_number, lr.letter_number, lt.name, COALESCE(u.email, u.phone, 'Warga'), r.full_name, lr.status, lr.submitted_at, lr.issued_at
		FROM letter_requests lr
		JOIN letter_types lt ON lt.id = lr.letter_type_id AND lt.organization_id = lr.organization_id
		JOIN users u ON u.id = lr.requester_user_id AND u.organization_id = lr.organization_id
		JOIN residents r ON r.id = lr.resident_id AND r.organization_id = lr.organization_id
		WHERE %s ORDER BY lr.created_at DESC`, strings.Join(conditions, " AND ")), args...)
	if err != nil {
		return nil, fmt.Errorf("query letters report: %w", err)
	}
	defer rows.Close()
	items := []LetterReportItem{}
	for rows.Next() {
		var item LetterReportItem
		var submittedAt, issuedAt *time.Time
		if err := rows.Scan(&item.ID, &item.RequestNumber, &item.LetterNumber, &item.LetterTypeName, &item.RequesterName, &item.ResidentName, &item.Status, &submittedAt, &issuedAt); err != nil {
			return nil, fmt.Errorf("scan letter report: %w", err)
		}
		if submittedAt != nil {
			value := submittedAt.UTC().Format(time.RFC3339)
			item.SubmittedAt = &value
		}
		if issuedAt != nil {
			value := issuedAt.UTC().Format(time.RFC3339)
			item.IssuedAt = &value
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Complaints(ctx context.Context, principal *auth.Principal, filter Filter) ([]ComplaintReportItem, error) {
	if principal == nil {
		return nil, ErrForbidden
	}
	if err := filter.Validate(); err != nil {
		return nil, err
	}
	conditions, args, index := []string{"c.organization_id = $1"}, []any{principal.OrganizationID}, 2
	if filter.Status != "" {
		conditions, args, index = append(conditions, fmt.Sprintf("c.status = $%d", index)), append(args, filter.Status), index+1
	}
	if filter.StartDate != "" {
		conditions, args, index = append(conditions, fmt.Sprintf("c.created_at >= $%d::date", index)), append(args, filter.StartDate), index+1
	}
	if filter.EndDate != "" {
		conditions, args = append(conditions, fmt.Sprintf("c.created_at < ($%d::date + interval '1 day')", index)), append(args, filter.EndDate)
	}
	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT c.id, c.ticket_number, c.category, c.title, c.priority, c.status, COALESCE(ru.email, ru.phone, 'Warga'), NULLIF(COALESCE(au.email, au.phone, ''), ''), c.created_at, c.resolved_at
		FROM complaints c
		JOIN users ru ON ru.id = c.reporter_user_id AND ru.organization_id = c.organization_id
		LEFT JOIN users au ON au.id = c.assigned_to AND au.organization_id = c.organization_id
		WHERE %s ORDER BY c.created_at DESC`, strings.Join(conditions, " AND ")), args...)
	if err != nil {
		return nil, fmt.Errorf("query complaints report: %w", err)
	}
	defer rows.Close()
	items := []ComplaintReportItem{}
	for rows.Next() {
		var item ComplaintReportItem
		var createdAt time.Time
		var resolvedAt *time.Time
		if err := rows.Scan(&item.ID, &item.TicketNumber, &item.Category, &item.Title, &item.Priority, &item.Status, &item.ReporterName, &item.AssignedToName, &createdAt, &resolvedAt); err != nil {
			return nil, fmt.Errorf("scan complaint report: %w", err)
		}
		item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		if resolvedAt != nil {
			value := resolvedAt.UTC().Format(time.RFC3339)
			item.ResolvedAt = &value
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) RecordExportAudit(ctx context.Context, principal *auth.Principal, reportType string, recordCount int) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, metadata)
		VALUES ($1, $2, 'report.export', $3, jsonb_build_object('record_count', $4::int))`,
		principal.OrganizationID, principal.UserID, reportType, recordCount,
	)
	if err != nil {
		return fmt.Errorf("audit report export: %w", err)
	}
	return nil
}

func (f Filter) Validate() error {
	for _, value := range []string{f.StartDate, f.EndDate} {
		if value != "" {
			if _, err := time.Parse(time.DateOnly, value); err != nil {
				return ErrValidation
			}
		}
	}
	if f.StartDate != "" && f.EndDate != "" && f.StartDate > f.EndDate {
		return ErrValidation
	}
	return nil
}