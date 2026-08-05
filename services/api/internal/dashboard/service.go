package dashboard

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) GetResidentDashboard(ctx context.Context, principal *auth.Principal) (ResidentDashboard, error) {
	if principal == nil {
		return ResidentDashboard{}, fmt.Errorf("dashboard principal is required")
	}

	dashboard := ResidentDashboard{
		ActiveInvoices:   []InvoiceSummary{},
		RecentPayments:   []PaymentSummary{},
		RecentLetters:    []LetterSummary{},
		RecentComplaints: []ComplaintSummary{},
		Announcements:    []AnnouncementSummary{},
		UpcomingEvents:   []EventSummary{},
	}

	var residentID *string
	if err := s.db.QueryRow(ctx, `
		SELECT resident_id
		FROM users
		WHERE id = $1 AND organization_id = $2`,
		principal.UserID, principal.OrganizationID,
	).Scan(&residentID); err != nil {
		return ResidentDashboard{}, fmt.Errorf("find linked resident: %w", err)
	}

	if residentID != nil {
		rows, err := s.db.Query(ctx, `
			SELECT i.id, i.invoice_number, dt.name, i.amount, i.paid_amount, i.due_date, i.status
			FROM invoices i
			JOIN due_types dt
			  ON dt.id = i.due_type_id AND dt.organization_id = i.organization_id
			JOIN household_members hm
			  ON hm.household_id = i.household_id AND hm.is_active
			WHERE i.organization_id = $1
			  AND hm.resident_id = $2
			  AND i.status IN ('unpaid', 'partial', 'pending_verification')
			ORDER BY i.due_date ASC, i.created_at ASC
			LIMIT 5`,
			principal.OrganizationID, *residentID,
		)
		if err != nil {
			return ResidentDashboard{}, fmt.Errorf("list resident invoices: %w", err)
		}
		for rows.Next() {
			var item InvoiceSummary
			var dueDate time.Time
			if err := rows.Scan(&item.ID, &item.InvoiceNumber, &item.DueTypeName, &item.Amount, &item.PaidAmount, &dueDate, &item.Status); err != nil {
				rows.Close()
				return ResidentDashboard{}, fmt.Errorf("scan resident invoice: %w", err)
			}
			item.DueDate = dueDate.Format(time.DateOnly)
			dashboard.ActiveInvoices = append(dashboard.ActiveInvoices, item)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return ResidentDashboard{}, fmt.Errorf("iterate resident invoices: %w", err)
		}
		rows.Close()

		rows, err = s.db.Query(ctx, `
			SELECT p.id, p.payment_number, i.invoice_number, p.amount, p.paid_at, p.verification_status
			FROM payments p
			JOIN invoices i
			  ON i.id = p.invoice_id AND i.organization_id = p.organization_id
			JOIN household_members hm
			  ON hm.household_id = i.household_id AND hm.is_active
			WHERE p.organization_id = $1
			  AND hm.resident_id = $2
			ORDER BY p.created_at DESC
			LIMIT 5`,
			principal.OrganizationID, *residentID,
		)
		if err != nil {
			return ResidentDashboard{}, fmt.Errorf("list resident payments: %w", err)
		}
		for rows.Next() {
			var item PaymentSummary
			var paidAt time.Time
			if err := rows.Scan(&item.ID, &item.PaymentNumber, &item.InvoiceNumber, &item.Amount, &paidAt, &item.VerificationStatus); err != nil {
				rows.Close()
				return ResidentDashboard{}, fmt.Errorf("scan resident payment: %w", err)
			}
			item.PaidAt = paidAt.UTC().Format(time.RFC3339)
			dashboard.RecentPayments = append(dashboard.RecentPayments, item)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return ResidentDashboard{}, fmt.Errorf("iterate resident payments: %w", err)
		}
		rows.Close()
	}

	rows, err := s.db.Query(ctx, `
		SELECT lr.id, lr.request_number, lt.name, lr.status, lr.updated_at
		FROM letter_requests lr
		JOIN letter_types lt
		  ON lt.id = lr.letter_type_id AND lt.organization_id = lr.organization_id
		WHERE lr.organization_id = $1 AND lr.requester_user_id = $2
		ORDER BY lr.updated_at DESC
		LIMIT 5`,
		principal.OrganizationID, principal.UserID,
	)
	if err != nil {
		return ResidentDashboard{}, fmt.Errorf("list resident letters: %w", err)
	}
	for rows.Next() {
		var item LetterSummary
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.RequestNumber, &item.LetterType, &item.Status, &updatedAt); err != nil {
			rows.Close()
			return ResidentDashboard{}, fmt.Errorf("scan resident letter: %w", err)
		}
		item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
		dashboard.RecentLetters = append(dashboard.RecentLetters, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return ResidentDashboard{}, fmt.Errorf("iterate resident letters: %w", err)
	}
	rows.Close()

	rows, err = s.db.Query(ctx, `
		SELECT id, ticket_number, title, priority, status, updated_at
		FROM complaints
		WHERE organization_id = $1 AND reporter_user_id = $2
		ORDER BY updated_at DESC
		LIMIT 5`,
		principal.OrganizationID, principal.UserID,
	)
	if err != nil {
		return ResidentDashboard{}, fmt.Errorf("list resident complaints: %w", err)
	}
	for rows.Next() {
		var item ComplaintSummary
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.TicketNumber, &item.Title, &item.Priority, &item.Status, &updatedAt); err != nil {
			rows.Close()
			return ResidentDashboard{}, fmt.Errorf("scan resident complaint: %w", err)
		}
		item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
		dashboard.RecentComplaints = append(dashboard.RecentComplaints, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return ResidentDashboard{}, fmt.Errorf("iterate resident complaints: %w", err)
	}
	rows.Close()

	rows, err = s.db.Query(ctx, `
		SELECT DISTINCT a.id, a.title, a.category, a.priority, a.publish_at
		FROM announcements a
		JOIN announcement_targets at ON at.announcement_id = a.id
		WHERE a.organization_id = $1
		  AND a.status = 'published'
		  AND a.publish_at <= now()
		  AND (a.expire_at IS NULL OR a.expire_at > now())
		  AND (
			at.target_type = 'all'
			OR (at.target_type = 'role' AND at.target_id IN (
				SELECT ur.role_id FROM user_roles ur WHERE ur.user_id = $2
			))
			OR ($3::uuid IS NOT NULL AND at.target_type = 'household' AND at.target_id IN (
				SELECT household_id FROM household_members WHERE resident_id = $3 AND is_active
			))
			OR ($3::uuid IS NOT NULL AND at.target_type = 'house_unit' AND at.target_id IN (
				SELECT h.house_unit_id
				FROM households h
				JOIN household_members hm ON hm.household_id = h.id AND hm.is_active
				WHERE h.organization_id = $1 AND hm.resident_id = $3
			))
		  )
		ORDER BY a.priority DESC, a.publish_at DESC
		LIMIT 5`,
		principal.OrganizationID, principal.UserID, residentID,
	)
	if err != nil {
		return ResidentDashboard{}, fmt.Errorf("list resident announcements: %w", err)
	}
	for rows.Next() {
		var item AnnouncementSummary
		var publishedAt time.Time
		if err := rows.Scan(&item.ID, &item.Title, &item.Category, &item.Priority, &publishedAt); err != nil {
			rows.Close()
			return ResidentDashboard{}, fmt.Errorf("scan resident announcement: %w", err)
		}
		item.PublishAt = publishedAt.UTC().Format(time.RFC3339)
		dashboard.Announcements = append(dashboard.Announcements, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return ResidentDashboard{}, fmt.Errorf("iterate resident announcements: %w", err)
	}
	rows.Close()

	rows, err = s.db.Query(ctx, `
		SELECT id, title, starts_at, COALESCE(location, '')
		FROM events
		WHERE organization_id = $1
		  AND status IN ('planned', 'ongoing')
		  AND starts_at >= now() - interval '1 day'
		ORDER BY starts_at ASC
		LIMIT 5`,
		principal.OrganizationID,
	)
	if err != nil {
		return ResidentDashboard{}, fmt.Errorf("list upcoming events: %w", err)
	}
	for rows.Next() {
		var item EventSummary
		var startsAt time.Time
		if err := rows.Scan(&item.ID, &item.Title, &startsAt, &item.Location); err != nil {
			rows.Close()
			return ResidentDashboard{}, fmt.Errorf("scan upcoming event: %w", err)
		}
		item.StartsAt = startsAt.UTC().Format(time.RFC3339)
		dashboard.UpcomingEvents = append(dashboard.UpcomingEvents, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return ResidentDashboard{}, fmt.Errorf("iterate upcoming events: %w", err)
	}
	rows.Close()

	publicCashSummary, err := s.getPublicCashSummary(ctx, principal.OrganizationID)
	if err != nil {
		return ResidentDashboard{}, err
	}
	dashboard.PublicCashSummary = publicCashSummary

	return dashboard, nil
}

func (s *Service) getPublicCashSummary(ctx context.Context, organizationID string) (*PublicCashSummary, error) {
	var (
		summary                PublicCashSummary
		summaryID              string
		periodStart, periodEnd time.Time
	)

	err := s.db.QueryRow(ctx, `
		SELECT s.id, s.period_start, s.period_end, s.total_income, s.total_expense, s.ending_balance
		FROM public_cash_summaries s
		JOIN cash_publication_policies p
		  ON p.organization_id = s.organization_id
		WHERE s.organization_id = $1
		  AND p.is_public
		  AND (p.public_until IS NULL OR p.public_until >= CURRENT_DATE)
		ORDER BY s.period_end DESC, s.published_at DESC
		LIMIT 1`,
		organizationID,
	).Scan(
		&summaryID, &periodStart, &periodEnd, &summary.TotalIncome,
		&summary.TotalExpense, &summary.EndingBalance,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get public cash summary: %w", err)
	}

	summary.PeriodStart = periodStart.Format(time.DateOnly)
	summary.PeriodEnd = periodEnd.Format(time.DateOnly)
	summary.Categories = []PublicCashCategorySummary{}

	rows, err := s.db.Query(ctx, `
		SELECT category_name, transaction_type, total_amount
		FROM public_cash_summary_categories
		WHERE public_cash_summary_id = $1
		ORDER BY transaction_type, category_name`,
		summaryID,
	)
	if err != nil {
		return nil, fmt.Errorf("list public cash categories: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category PublicCashCategorySummary
		if err := rows.Scan(&category.CategoryName, &category.TransactionType, &category.TotalAmount); err != nil {
			return nil, fmt.Errorf("scan public cash category: %w", err)
		}
		summary.Categories = append(summary.Categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate public cash categories: %w", err)
	}
	return &summary, nil
}

func (s *Service) GetAdminDashboard(ctx context.Context, principal *auth.Principal) (AdminDashboard, error) {
	if principal == nil {
		return AdminDashboard{}, fmt.Errorf("dashboard principal is required")
	}

	dashboard := AdminDashboard{
		RecentPayments:   []PaymentSummary{},
		RecentLetters:    []LetterSummary{},
		RecentComplaints: []ComplaintSummary{},
	}

	if err := s.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM households WHERE organization_id = $1 AND move_out_date IS NULL),
			(SELECT COUNT(*) FROM residents WHERE organization_id = $1 AND resident_status = 'active'),
			(SELECT COUNT(*) FROM invoices WHERE organization_id = $1 AND status IN ('unpaid', 'partial', 'pending_verification')),
			(SELECT COALESCE(SUM(amount - paid_amount - adjustment_amount), 0) FROM invoices WHERE organization_id = $1 AND status IN ('unpaid', 'partial', 'pending_verification')),
			(SELECT COUNT(*) FROM payments WHERE organization_id = $1 AND verification_status = 'pending'),
			(SELECT COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE -amount END), 0) FROM cash_transactions WHERE organization_id = $1 AND status = 'active'),
			(SELECT COUNT(*) FROM letter_requests WHERE organization_id = $1 AND status IN ('submitted', 'under_review', 'needs_revision', 'awaiting_approval')),
			(SELECT COUNT(*) FROM complaints WHERE organization_id = $1 AND status IN ('new', 'reviewed', 'in_progress', 'waiting_information'))`,
		principal.OrganizationID,
	).Scan(
		&dashboard.ActiveHouseholds,
		&dashboard.ActiveResidents,
		&dashboard.ActiveInvoices,
		&dashboard.OutstandingAmount,
		&dashboard.PendingPayments,
		&dashboard.CashBalance,
		&dashboard.PendingLetters,
		&dashboard.OpenComplaints,
	); err != nil {
		return AdminDashboard{}, fmt.Errorf("get admin dashboard counters: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT p.id, p.payment_number, i.invoice_number, p.amount, p.paid_at, p.verification_status
		FROM payments p
		JOIN invoices i ON i.id = p.invoice_id AND i.organization_id = p.organization_id
		WHERE p.organization_id = $1
		ORDER BY p.created_at DESC
		LIMIT 5`,
		principal.OrganizationID,
	)
	if err != nil {
		return AdminDashboard{}, fmt.Errorf("list admin payments: %w", err)
	}
	for rows.Next() {
		var item PaymentSummary
		var paidAt time.Time
		if err := rows.Scan(&item.ID, &item.PaymentNumber, &item.InvoiceNumber, &item.Amount, &paidAt, &item.VerificationStatus); err != nil {
			rows.Close()
			return AdminDashboard{}, fmt.Errorf("scan admin payment: %w", err)
		}
		item.PaidAt = paidAt.UTC().Format(time.RFC3339)
		dashboard.RecentPayments = append(dashboard.RecentPayments, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return AdminDashboard{}, fmt.Errorf("iterate admin payments: %w", err)
	}
	rows.Close()

	rows, err = s.db.Query(ctx, `
		SELECT lr.id, lr.request_number, lt.name, lr.status, lr.updated_at
		FROM letter_requests lr
		JOIN letter_types lt
		  ON lt.id = lr.letter_type_id AND lt.organization_id = lr.organization_id
		WHERE lr.organization_id = $1
		ORDER BY lr.updated_at DESC
		LIMIT 5`,
		principal.OrganizationID,
	)
	if err != nil {
		return AdminDashboard{}, fmt.Errorf("list admin letters: %w", err)
	}
	for rows.Next() {
		var item LetterSummary
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.RequestNumber, &item.LetterType, &item.Status, &updatedAt); err != nil {
			rows.Close()
			return AdminDashboard{}, fmt.Errorf("scan admin letter: %w", err)
		}
		item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
		dashboard.RecentLetters = append(dashboard.RecentLetters, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return AdminDashboard{}, fmt.Errorf("iterate admin letters: %w", err)
	}
	rows.Close()

	rows, err = s.db.Query(ctx, `
		SELECT id, ticket_number, title, priority, status, updated_at
		FROM complaints
		WHERE organization_id = $1
		ORDER BY updated_at DESC
		LIMIT 5`,
		principal.OrganizationID,
	)
	if err != nil {
		return AdminDashboard{}, fmt.Errorf("list admin complaints: %w", err)
	}
	for rows.Next() {
		var item ComplaintSummary
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.TicketNumber, &item.Title, &item.Priority, &item.Status, &updatedAt); err != nil {
			rows.Close()
			return AdminDashboard{}, fmt.Errorf("scan admin complaint: %w", err)
		}
		item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
		dashboard.RecentComplaints = append(dashboard.RecentComplaints, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return AdminDashboard{}, fmt.Errorf("iterate admin complaints: %w", err)
	}
	rows.Close()

	return dashboard, nil
}
