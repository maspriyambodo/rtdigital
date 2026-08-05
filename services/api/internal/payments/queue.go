package payments

import (
	"context"
	"fmt"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

var standardRejectReasons = []string{
	"Bukti pembayaran tidak terbaca atau tidak sesuai.",
	"Nominal pembayaran tidak sesuai tagihan.",
	"Rekening tujuan atau tanggal transfer tidak sesuai.",
	"Pembayaran terduplikasi.",
}

// Queue returns pending payments with their allocations, current remaining
// invoice balances, and related payment history for one-screen verification.
func (s *Service) Queue(ctx context.Context, principal *auth.Principal) ([]PaymentQueueItem, error) {
	if principal == nil {
		return nil, ErrValidation
	}

	rows, err := s.db.Query(ctx, `
		SELECT p.id, p.invoice_id, i.invoice_number, p.payment_number, p.method, p.amount,
		       p.paid_at, p.proof_file_id, p.verification_status, p.verified_by, p.verified_at,
		       p.rejection_reason, p.cancelled_by, p.cancelled_at, p.cancellation_reason,
		       p.created_by, p.created_at, p.updated_at, i.status
		FROM payments p
		JOIN invoices i
		  ON i.organization_id = p.organization_id
		 AND i.id = p.invoice_id
		WHERE p.organization_id = $1
		  AND p.verification_status = 'pending'
		ORDER BY p.created_at ASC, p.id ASC`,
		principal.OrganizationID,
	)
	if err != nil {
		return nil, fmt.Errorf("list payment queue: %w", err)
	}
	defer rows.Close()

	items := []PaymentQueueItem{}
	for rows.Next() {
		var item PaymentQueueItem
		if err := rows.Scan(
			&item.ID, &item.InvoiceID, &item.InvoiceNumber, &item.PaymentNumber,
			&item.Method, &item.Amount, &item.PaidAt, &item.ProofFileID,
			&item.VerificationStatus, &item.VerifiedBy, &item.VerifiedAt,
			&item.RejectionReason, &item.CancelledBy, &item.CancelledAt,
			&item.CancellationReason, &item.CreatedBy, &item.CreatedAt,
			&item.UpdatedAt, &item.InvoiceStatus,
		); err != nil {
			return nil, fmt.Errorf("scan queued payment: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate payment queue: %w", err)
	}

	for index := range items {
		allocations, err := s.paymentAllocations(
			ctx, s.db, principal.OrganizationID, items[index].ID, items[index].InvoiceID, items[index].Amount,
		)
		if err != nil {
			return nil, err
		}
		for allocationIndex := range allocations {
			if err := s.db.QueryRow(ctx, `
				SELECT GREATEST(amount - adjustment_amount - paid_amount, 0)
				FROM invoices
				WHERE organization_id = $1 AND id = $2`,
				principal.OrganizationID, allocations[allocationIndex].InvoiceID,
			).Scan(&allocations[allocationIndex].RemainingAmount); err != nil {
				return nil, fmt.Errorf("get remaining invoice amount: %w", err)
			}
		}
		history, err := s.queueHistory(ctx, principal.OrganizationID, allocations, items[index].ID)
		if err != nil {
			return nil, err
		}
		items[index].Allocations = allocations
		items[index].RelevantHistory = history
		items[index].StandardRejectReasons = standardRejectReasons
	}
	return items, nil
}

func (s *Service) queueHistory(
	ctx context.Context, organizationID string, allocations []PaymentAllocation, excludePaymentID string,
) ([]Payment, error) {
	invoiceIDs := make([]string, 0, len(allocations))
	for _, allocation := range allocations {
		invoiceIDs = append(invoiceIDs, allocation.InvoiceID)
	}

	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT p.id, p.invoice_id, i.invoice_number, p.payment_number, p.method, p.amount,
		       p.paid_at, p.proof_file_id, p.verification_status, p.verified_by, p.verified_at,
		       p.rejection_reason, p.cancelled_by, p.cancelled_at, p.cancellation_reason,
		       p.created_by, p.created_at, p.updated_at, i.status
		FROM payments p
		JOIN invoices i
		  ON i.organization_id = p.organization_id
		 AND i.id = p.invoice_id
		LEFT JOIN payment_allocations pa
		  ON pa.organization_id = p.organization_id
		 AND pa.payment_id = p.id
		WHERE p.organization_id = $1
		  AND p.id <> $2
		  AND (p.invoice_id = ANY($3::uuid[]) OR pa.invoice_id = ANY($3::uuid[]))
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT 10`,
		organizationID, excludePaymentID, invoiceIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("list related payment history: %w", err)
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
			return nil, fmt.Errorf("scan related payment: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
