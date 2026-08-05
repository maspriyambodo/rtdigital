package payments

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
	"github.com/maspriyambodo/rtdigital/services/api/internal/cash"
	"github.com/maspriyambodo/rtdigital/services/api/internal/notifications"
)

type Service struct {
	db         *pgxpool.Pool
	cash       *cash.Service
	dispatcher *notifications.Dispatcher
	now        func() time.Time
}

func NewService(db *pgxpool.Pool, cashService *cash.Service) *Service {
	return &Service{
		db:   db,
		cash: cashService,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) SetNotificationDispatcher(dispatcher *notifications.Dispatcher) {
	s.dispatcher = dispatcher
}

func (s *Service) List(ctx context.Context, principal *auth.Principal, filter PaymentFilter) ([]Payment, error) {
	if principal == nil {
		return nil, ErrValidation
	}

	conditions := []string{"p.organization_id = $1"}
	args := []any{principal.OrganizationID}
	argument := 2

	if filter.InvoiceID != "" {
		conditions = append(conditions, fmt.Sprintf("p.invoice_id = $%d", argument))
		args = append(args, filter.InvoiceID)
		argument++
	}
	if filter.VerificationStatus != "" {
		switch filter.VerificationStatus {
		case "pending", "verified", "rejected", "cancelled":
		default:
			return nil, ErrValidation
		}
		conditions = append(conditions, fmt.Sprintf("p.verification_status = $%d", argument))
		args = append(args, filter.VerificationStatus)
		argument++
	}
	if !principal.HasPermission("payment.read") {
		conditions = append(conditions, fmt.Sprintf(`
			EXISTS (
				SELECT 1
				FROM household_members hm
				JOIN users u
				  ON u.organization_id = hm.organization_id
				 AND u.resident_id = hm.resident_id
				WHERE hm.organization_id = p.organization_id
				  AND hm.household_id = i.household_id
				  AND hm.is_active
				  AND u.id = $%d
			)`, argument))
		args = append(args, principal.UserID)
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.invoice_id, i.invoice_number, p.payment_number, p.method, p.amount,
		       p.paid_at, p.proof_file_id, p.verification_status, p.verified_by, p.verified_at,
		       p.rejection_reason, p.cancelled_by, p.cancelled_at, p.cancellation_reason,
		       p.created_by, p.created_at, p.updated_at, i.status
		FROM payments p
		JOIN invoices i
		  ON i.organization_id = p.organization_id
		 AND i.id = p.invoice_id
		WHERE %s
		ORDER BY p.created_at DESC`, strings.Join(conditions, " AND "))

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()

	items := []Payment{}
	for rows.Next() {
		var item Payment
		if err := rows.Scan(
			&item.ID, &item.InvoiceID, &item.InvoiceNumber, &item.PaymentNumber,
			&item.Method, &item.Amount, &item.PaidAt, &item.ProofFileID,
			&item.VerificationStatus, &item.VerifiedBy, &item.VerifiedAt,
			&item.RejectionReason, &item.CancelledBy, &item.CancelledAt,
			&item.CancellationReason, &item.CreatedBy, &item.CreatedAt,
			&item.UpdatedAt, &item.InvoiceStatus,
		); err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate payments: %w", err)
	}
	return items, nil
}

func (s *Service) Get(ctx context.Context, principal *auth.Principal, paymentID string) (Payment, error) {
	items, err := s.List(ctx, principal, PaymentFilter{})
	if err != nil {
		return Payment{}, err
	}
	for _, item := range items {
		if item.ID == paymentID {
			return item, nil
		}
	}
	return Payment{}, ErrPaymentNotFound
}

func (s *Service) Submit(ctx context.Context, principal *auth.Principal, idempotencyKey string, request SubmitPaymentRequest) (SubmitPaymentResponse, error) {
	now := s.now()
	if principal == nil || strings.TrimSpace(idempotencyKey) == "" || request.Validate(now) != nil {
		return SubmitPaymentResponse{}, ErrValidation
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return SubmitPaymentResponse{}, fmt.Errorf("begin payment submission: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		existingID, existingNumber, existingStatus, existingInvoiceStatus string
		existingCreatedAt                                                 time.Time
	)
	err = tx.QueryRow(ctx, `
		SELECT p.id, p.payment_number, p.verification_status, i.status, p.created_at
		FROM payments p
		JOIN invoices i
		  ON i.organization_id = p.organization_id
		 AND i.id = p.invoice_id
		WHERE p.organization_id = $1
		  AND p.idempotency_key = $2`,
		principal.OrganizationID,
		idempotencyKey,
	).Scan(&existingID, &existingNumber, &existingStatus, &existingInvoiceStatus, &existingCreatedAt)
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return SubmitPaymentResponse{}, fmt.Errorf("commit idempotent payment: %w", err)
		}
		return SubmitPaymentResponse{
			ID:                 existingID,
			PaymentNumber:      existingNumber,
			VerificationStatus: existingStatus,
			InvoiceStatus:      existingInvoiceStatus,
			CreatedAt:          existingCreatedAt,
		}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return SubmitPaymentResponse{}, fmt.Errorf("find idempotent payment: %w", err)
	}

	allocations := request.Allocations
	if len(allocations) == 0 {
		allocations = []PaymentAllocation{{InvoiceID: strings.TrimSpace(request.InvoiceID), Amount: request.Amount}}
	}

	// Lock in stable order to avoid deadlocks between concurrent multi-invoice reports.
	sort.Slice(allocations, func(i, j int) bool { return allocations[i].InvoiceID < allocations[j].InvoiceID })

	var householdID string
	for _, allocation := range allocations {
		var invoiceHouseholdID, invoiceStatus string
		var invoiceAmount, paidAmount, pendingAmount float64
		err = tx.QueryRow(ctx, `
			SELECT i.household_id, i.status, i.amount, i.paid_amount,
			       COALESCE((
			         SELECT SUM(CASE WHEN EXISTS (
			           SELECT 1 FROM payment_allocations pa WHERE pa.payment_id = p.id
			         ) THEN 0 ELSE p.amount END)
			         FROM payments p
			         WHERE p.organization_id = i.organization_id
			           AND p.invoice_id = i.id
			           AND p.verification_status = 'pending'
			       ), 0)
			       + COALESCE((
			         SELECT SUM(pa.amount)
			         FROM payment_allocations pa
			         JOIN payments p
			           ON p.organization_id = pa.organization_id
			          AND p.id = pa.payment_id
			         WHERE pa.organization_id = i.organization_id
			           AND pa.invoice_id = i.id
			           AND p.verification_status = 'pending'
			       ), 0)
			FROM invoices i
			WHERE i.organization_id = $1 AND i.id = $2
			FOR UPDATE`,
			principal.OrganizationID, allocation.InvoiceID,
		).Scan(&invoiceHouseholdID, &invoiceStatus, &invoiceAmount, &paidAmount, &pendingAmount)
		if errors.Is(err, pgx.ErrNoRows) {
			return SubmitPaymentResponse{}, ErrInvoiceNotFound
		}
		if err != nil {
			return SubmitPaymentResponse{}, fmt.Errorf("lock allocated invoice: %w", err)
		}
		if invoiceStatus == "paid" || invoiceStatus == "cancelled" {
			return SubmitPaymentResponse{}, ErrInvalidInvoiceState
		}
		if allocation.Amount > invoiceAmount-paidAmount-pendingAmount {
			return SubmitPaymentResponse{}, ErrConstraint
		}
		if householdID == "" {
			householdID = invoiceHouseholdID
		} else if householdID != invoiceHouseholdID {
			return SubmitPaymentResponse{}, ErrForbidden
		}
	}

	var belongsToHousehold bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM household_members hm
			JOIN users u
			  ON u.organization_id = hm.organization_id
			 AND u.resident_id = hm.resident_id
			WHERE hm.organization_id = $1
			  AND hm.household_id = $2
			  AND hm.is_active
			  AND u.id = $3
		)`,
		principal.OrganizationID, householdID, principal.UserID,
	).Scan(&belongsToHousehold); err != nil {
		return SubmitPaymentResponse{}, fmt.Errorf("authorize payment household: %w", err)
	}
	if !belongsToHousehold {
		return SubmitPaymentResponse{}, ErrForbidden
	}

	if request.ProofFileID != nil {
		var confirmed bool
		if err := tx.QueryRow(ctx, `
			SELECT confirmed_at IS NOT NULL
			FROM file_objects
			WHERE id = $1
			  AND organization_id = $2
			  AND uploaded_by = $3
			  AND deleted_at IS NULL
			FOR UPDATE`,
			*request.ProofFileID,
			principal.OrganizationID,
			principal.UserID,
		).Scan(&confirmed); errors.Is(err, pgx.ErrNoRows) {
			return SubmitPaymentResponse{}, ErrFileNotFound
		} else if err != nil {
			return SubmitPaymentResponse{}, fmt.Errorf("lock proof file: %w", err)
		} else if !confirmed {
			return SubmitPaymentResponse{}, ErrFileNotConfirmed
		}
	}

	paymentID := newUUID()
	paymentNumber := fmt.Sprintf("PAY-%s-%s", now.Format("0601"), paymentID[:8])
	primaryInvoiceID := allocations[0].InvoiceID
	if _, err := tx.Exec(ctx, `
		INSERT INTO payments (
			id, organization_id, invoice_id, payment_number, method, amount, paid_at,
			proof_file_id, verification_status, idempotency_key, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending', $9, $10, $11, $11)`,
		paymentID, principal.OrganizationID, primaryInvoiceID, paymentNumber, request.Method,
		request.Amount, request.PaidAt, request.ProofFileID, idempotencyKey, principal.UserID, now,
	); err != nil {
		return SubmitPaymentResponse{}, mapDatabaseError(err, "insert payment")
	}
	for index, allocation := range allocations {
		if _, err := tx.Exec(ctx, `
			INSERT INTO payment_allocations (id, organization_id, payment_id, invoice_id, amount)
			VALUES ($1, $2, $3, $4, $5)`,
			newUUID(), principal.OrganizationID, paymentID, allocation.InvoiceID, allocation.Amount,
		); err != nil {
			return SubmitPaymentResponse{}, mapDatabaseError(err, "insert payment allocation")
		}
		allocations[index].ID = ""
	}

	if request.ProofFileID != nil {
		if _, err := tx.Exec(ctx, `
			INSERT INTO file_attachments (
				id, organization_id, file_id, entity_type, entity_id, purpose
			) VALUES ($1, $2, $3, 'payment', $4, 'payment_proof')`,
			newUUID(),
			principal.OrganizationID,
			*request.ProofFileID,
			paymentID,
		); err != nil {
			return SubmitPaymentResponse{}, mapDatabaseError(err, "attach payment proof")
		}
	}

	invoiceStatus := ""
	for _, allocation := range allocations {
		status, err := s.syncInvoiceStatus(ctx, tx, principal.OrganizationID, allocation.InvoiceID)
		if err != nil {
			return SubmitPaymentResponse{}, err
		}
		if allocation.InvoiceID == primaryInvoiceID {
			invoiceStatus = status
		}
	}
	if err := s.audit(ctx, tx, principal, "payment.submit", paymentID); err != nil {
		return SubmitPaymentResponse{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return SubmitPaymentResponse{}, fmt.Errorf("commit payment submission: %w", err)
	}
	s.notifyTreasurers(principal.OrganizationID, notifications.DispatchJob{
		Type:          "payment_submitted",
		Title:         "Laporan pembayaran baru",
		Body:          fmt.Sprintf("Pembayaran %s sebesar Rp%.0f menunggu verifikasi.", paymentNumber, request.Amount),
		ReferenceType: "payment",
		ReferenceID:   paymentID,
	})

	return SubmitPaymentResponse{
		ID:                 paymentID,
		PaymentNumber:      paymentNumber,
		VerificationStatus: "pending",
		InvoiceStatus:      invoiceStatus,
		Allocations:        allocations,
		CreatedAt:          now,
	}, nil
}

func (s *Service) Verify(ctx context.Context, principal *auth.Principal, paymentID string, _ VerifyPaymentRequest) (PaymentActionResponse, error) {
	if principal == nil {
		return PaymentActionResponse{}, ErrValidation
	}
	return s.resolve(ctx, principal, paymentID, "verified", "")
}

func (s *Service) Reject(ctx context.Context, principal *auth.Principal, paymentID string, request RejectPaymentRequest) (PaymentActionResponse, error) {
	if principal == nil || request.Validate() != nil {
		return PaymentActionResponse{}, ErrValidation
	}
	return s.resolve(ctx, principal, paymentID, "rejected", strings.TrimSpace(request.Reason))
}

func (s *Service) Cancel(ctx context.Context, principal *auth.Principal, paymentID string, request CancelPaymentRequest) (PaymentActionResponse, error) {
	if principal == nil || request.Validate() != nil {
		return PaymentActionResponse{}, ErrValidation
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return PaymentActionResponse{}, fmt.Errorf("begin payment cancellation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status, invoiceID string
	var amount float64
	err = tx.QueryRow(ctx, `
		SELECT verification_status, invoice_id, amount
		FROM payments
		WHERE id = $1 AND organization_id = $2
		FOR UPDATE`,
		paymentID,
		principal.OrganizationID,
	).Scan(&status, &invoiceID, &amount)
	if errors.Is(err, pgx.ErrNoRows) {
		return PaymentActionResponse{}, ErrPaymentNotFound
	}
	if err != nil {
		return PaymentActionResponse{}, fmt.Errorf("lock payment cancellation: %w", err)
	}
	if status == "cancelled" || status == "rejected" {
		return PaymentActionResponse{}, ErrInvalidPaymentState
	}

	now := s.now()
	if _, err := tx.Exec(ctx, `
		UPDATE payments
		SET verification_status = 'cancelled',
		    cancelled_by = $1,
		    cancelled_at = $2,
		    cancellation_reason = $3
		WHERE id = $4 AND organization_id = $5`,
		principal.UserID,
		now,
		strings.TrimSpace(request.Reason),
		paymentID,
		principal.OrganizationID,
	); err != nil {
		return PaymentActionResponse{}, fmt.Errorf("cancel payment: %w", err)
	}
	allocations, err := s.paymentAllocations(ctx, tx, principal.OrganizationID, paymentID, invoiceID, amount)
	if err != nil {
		return PaymentActionResponse{}, err
	}
	if status == "verified" {
		for _, allocation := range allocations {
			if _, err := tx.Exec(ctx, `
				UPDATE invoices
				SET paid_amount = paid_amount - $1
				WHERE id = $2 AND organization_id = $3`,
				allocation.Amount, allocation.InvoiceID, principal.OrganizationID,
			); err != nil {
				return PaymentActionResponse{}, fmt.Errorf("reverse allocated invoice payment: %w", err)
			}
		}
	}

	invoiceStatus := ""
	for _, allocation := range allocations {
		status, err := s.syncInvoiceStatus(ctx, tx, principal.OrganizationID, allocation.InvoiceID)
		if err != nil {
			return PaymentActionResponse{}, err
		}
		if allocation.InvoiceID == invoiceID {
			invoiceStatus = status
		}
	}
	if err := s.audit(ctx, tx, principal, "payment.cancel", paymentID); err != nil {
		return PaymentActionResponse{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return PaymentActionResponse{}, fmt.Errorf("commit payment cancellation: %w", err)
	}
	return PaymentActionResponse{
		ID:                 paymentID,
		VerificationStatus: "cancelled",
		InvoiceStatus:      invoiceStatus,
	}, nil
}

func (s *Service) resolve(ctx context.Context, principal *auth.Principal, paymentID, outcome, reason string) (PaymentActionResponse, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return PaymentActionResponse{}, fmt.Errorf("begin payment resolution: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status, createdBy, invoiceID, paymentNumber string
	var amount float64
	var paidAt time.Time
	err = tx.QueryRow(ctx, `
		SELECT verification_status, created_by, invoice_id, amount, payment_number, paid_at
		FROM payments
		WHERE id = $1 AND organization_id = $2
		FOR UPDATE`,
		paymentID,
		principal.OrganizationID,
	).Scan(&status, &createdBy, &invoiceID, &amount, &paymentNumber, &paidAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return PaymentActionResponse{}, ErrPaymentNotFound
	}
	if err != nil {
		return PaymentActionResponse{}, fmt.Errorf("lock payment: %w", err)
	}
	if status != "pending" {
		return PaymentActionResponse{}, ErrInvalidPaymentState
	}
	if createdBy == principal.UserID {
		return PaymentActionResponse{}, ErrForbidden
	}

	now := s.now()
	allocations, err := s.paymentAllocations(ctx, tx, principal.OrganizationID, paymentID, invoiceID, amount)
	if err != nil {
		return PaymentActionResponse{}, err
	}
	if outcome == "verified" {
		if _, err := tx.Exec(ctx, `
			UPDATE payments
			SET verification_status = 'verified', verified_by = $1, verified_at = $2
			WHERE id = $3 AND organization_id = $4`,
			principal.UserID, now, paymentID, principal.OrganizationID,
		); err != nil {
			return PaymentActionResponse{}, fmt.Errorf("verify payment: %w", err)
		}
		for _, allocation := range allocations {
			if _, err := tx.Exec(ctx, `
				UPDATE invoices
				SET paid_amount = paid_amount + $1
				WHERE id = $2 AND organization_id = $3`,
				allocation.Amount, allocation.InvoiceID, principal.OrganizationID,
			); err != nil {
				return PaymentActionResponse{}, fmt.Errorf("update allocated invoice amount: %w", err)
			}
		}
		if s.cash == nil {
			return PaymentActionResponse{}, fmt.Errorf("cash service is required")
		}
		if err := s.cash.RecordVerifiedPayment(ctx, tx, principal, paymentID, paymentNumber, amount, paidAt); err != nil {
			return PaymentActionResponse{}, err
		}
	} else {
		if _, err := tx.Exec(ctx, `
			UPDATE payments
			SET verification_status = 'rejected',
			    verified_by = $1,
			    verified_at = $2,
			    rejection_reason = $3
			WHERE id = $4 AND organization_id = $5`,
			principal.UserID,
			now,
			reason,
			paymentID,
			principal.OrganizationID,
		); err != nil {
			return PaymentActionResponse{}, fmt.Errorf("reject payment: %w", err)
		}
	}

	invoiceStatus := ""
	for _, allocation := range allocations {
		status, err := s.syncInvoiceStatus(ctx, tx, principal.OrganizationID, allocation.InvoiceID)
		if err != nil {
			return PaymentActionResponse{}, err
		}
		if allocation.InvoiceID == invoiceID {
			invoiceStatus = status
		}
	}
	action := "payment.verify"
	if outcome == "rejected" {
		action = "payment.reject"
	}
	if err := s.audit(ctx, tx, principal, action, paymentID); err != nil {
		return PaymentActionResponse{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return PaymentActionResponse{}, fmt.Errorf("commit payment resolution: %w", err)
	}
	title := "Pembayaran diverifikasi"
	body := fmt.Sprintf("Pembayaran %s sebesar Rp%.0f telah diverifikasi.", paymentNumber, amount)
	if outcome == "rejected" {
		title = "Pembayaran ditolak"
		body = fmt.Sprintf("Pembayaran %s ditolak: %s", paymentNumber, reason)
	}
	s.dispatchNotification(notifications.DispatchJob{
		OrganizationID:  principal.OrganizationID,
		RecipientUserID: createdBy,
		Type:            "payment_" + outcome,
		Title:           title,
		Body:            body,
		ReferenceType:   "payment",
		ReferenceID:     paymentID,
	})

	return PaymentActionResponse{
		ID:                 paymentID,
		VerificationStatus: outcome,
		VerifiedAt:         &now,
		InvoiceStatus:      invoiceStatus,
	}, nil
}

func (s *Service) notifyTreasurers(organizationID string, job notifications.DispatchJob) {
	if s.dispatcher == nil {
		return
	}
	rows, err := s.db.Query(context.Background(), `
		SELECT DISTINCT u.id
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles r ON r.id = ur.role_id
		WHERE u.organization_id = $1
		  AND u.status = 'active'
		  AND r.code = 'bendahara'`,
		organizationID,
	)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		if rows.Scan(&userID) == nil {
			job.OrganizationID = organizationID
			job.RecipientUserID = userID
			s.dispatchNotification(job)
		}
	}
}

func (s *Service) dispatchNotification(job notifications.DispatchJob) {
	if s.dispatcher != nil {
		s.dispatcher.Dispatch(job)
	}
}

func (s *Service) paymentAllocations(
	ctx context.Context,
	db interface {
		Query(context.Context, string, ...any) (pgx.Rows, error)
	},
	organizationID, paymentID, legacyInvoiceID string,
	legacyAmount float64,
) ([]PaymentAllocation, error) {
	rows, err := db.Query(ctx, `
		SELECT pa.id, pa.invoice_id, i.invoice_number, pa.amount, pa.created_at
		FROM payment_allocations pa
		JOIN invoices i
		  ON i.organization_id = pa.organization_id
		 AND i.id = pa.invoice_id
		WHERE pa.organization_id = $1 AND pa.payment_id = $2
		ORDER BY pa.invoice_id`,
		organizationID, paymentID,
	)
	if err != nil {
		return nil, fmt.Errorf("list payment allocations: %w", err)
	}
	defer rows.Close()

	allocations := []PaymentAllocation{}
	for rows.Next() {
		var allocation PaymentAllocation
		if err := rows.Scan(
			&allocation.ID, &allocation.InvoiceID, &allocation.InvoiceNumber,
			&allocation.Amount, &allocation.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan payment allocation: %w", err)
		}
		allocations = append(allocations, allocation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate payment allocations: %w", err)
	}
	if len(allocations) == 0 {
		return []PaymentAllocation{{InvoiceID: legacyInvoiceID, Amount: legacyAmount}}, nil
	}
	return allocations, nil
}

func (s *Service) syncInvoiceStatus(ctx context.Context, tx pgx.Tx, organizationID, invoiceID string) (string, error) {
	var (
		currentStatus string
		amount        float64
		paidAmount    float64
	)
	err := tx.QueryRow(ctx, `
		SELECT status, amount, paid_amount
		FROM invoices
		WHERE id = $1 AND organization_id = $2
		FOR UPDATE`,
		invoiceID,
		organizationID,
	).Scan(&currentStatus, &amount, &paidAmount)
	if err != nil {
		return "", fmt.Errorf("lock invoice status: %w", err)
	}
	if currentStatus == "cancelled" {
		return currentStatus, nil
	}

	var pending bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM payments p
			WHERE p.organization_id = $1
			  AND p.verification_status = 'pending'
			  AND (
			    (p.invoice_id = $2 AND NOT EXISTS (
			      SELECT 1 FROM payment_allocations pa
			      WHERE pa.organization_id = p.organization_id
			        AND pa.payment_id = p.id
			    ))
			    OR EXISTS (
			      SELECT 1 FROM payment_allocations pa
			      WHERE pa.organization_id = p.organization_id
			        AND pa.payment_id = p.id
			        AND pa.invoice_id = $2
			    )
			  )
		)`,
		organizationID, invoiceID,
	).Scan(&pending); err != nil {
		return "", fmt.Errorf("check pending payments: %w", err)
	}

	status := "unpaid"
	switch {
	case paidAmount >= amount:
		status = "paid"
	case pending:
		status = "pending_verification"
	case paidAmount > 0:
		status = "partial"
	}
	if _, err := tx.Exec(ctx, `
		UPDATE invoices
		SET status = $1
		WHERE id = $2 AND organization_id = $3`,
		status,
		invoiceID,
		organizationID,
	); err != nil {
		return "", fmt.Errorf("sync invoice status: %w", err)
	}
	return status, nil
}

func (s *Service) audit(ctx context.Context, tx pgx.Tx, principal *auth.Principal, action, paymentID string) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id)
		VALUES ($1, $2, $3, 'payment', $4)`,
		principal.OrganizationID,
		principal.UserID,
		action,
		paymentID,
	); err != nil {
		return fmt.Errorf("audit %s: %w", action, err)
	}
	return nil
}

func mapDatabaseError(err error, operation string) error {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "unique"):
		return ErrDuplicateData
	case strings.Contains(message, "check"), strings.Contains(message, "foreign key"):
		return ErrConstraint
	default:
		return fmt.Errorf("%s: %w", operation, err)
	}
}

func newUUID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic("secure random source unavailable")
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}
