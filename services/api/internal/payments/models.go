package payments

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrValidation          = errors.New("validation failed")
	ErrPaymentNotFound     = errors.New("payment not found")
	ErrInvoiceNotFound     = errors.New("invoice not found")
	ErrFileNotFound        = errors.New("file not found")
	ErrFileNotConfirmed    = errors.New("file not confirmed")
	ErrDuplicateData       = errors.New("duplicate data")
	ErrConstraint          = errors.New("business constraint violated")
	ErrInvalidPaymentState = errors.New("invalid payment state")
	ErrInvalidInvoiceState = errors.New("invalid invoice state")
	ErrForbidden           = errors.New("forbidden")
)

type Payment struct {
	ID                 string     `json:"id"`
	InvoiceID          string     `json:"invoice_id"`
	InvoiceNumber      string     `json:"invoice_number"`
	PaymentNumber      string     `json:"payment_number"`
	Method             string     `json:"method"`
	Amount             float64    `json:"amount"`
	PaidAt             time.Time  `json:"paid_at"`
	ProofFileID        *string    `json:"proof_file_id,omitempty"`
	VerificationStatus string     `json:"verification_status"`
	VerifiedBy         *string    `json:"verified_by,omitempty"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	RejectionReason    *string    `json:"rejection_reason,omitempty"`
	CancelledBy        *string    `json:"cancelled_by,omitempty"`
	CancelledAt        *time.Time `json:"cancelled_at,omitempty"`
	CancellationReason *string    `json:"cancellation_reason,omitempty"`
	CreatedBy          string     `json:"created_by"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	InvoiceStatus      string     `json:"invoice_status,omitempty"`
}

type PaymentAllocation struct {
	ID              string    `json:"id,omitempty"`
	InvoiceID       string    `json:"invoice_id"`
	InvoiceNumber   string    `json:"invoice_number,omitempty"`
	Amount          float64   `json:"amount"`
	RemainingAmount float64   `json:"remaining_amount,omitempty"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
}

type PaymentQueueItem struct {
	Payment
	Allocations           []PaymentAllocation `json:"allocations"`
	RelevantHistory       []Payment           `json:"relevant_history"`
	StandardRejectReasons []string            `json:"standard_reject_reasons"`
}

type SubmitPaymentRequest struct {
	// InvoiceID is retained for existing single-invoice clients.
	InvoiceID   string              `json:"invoice_id,omitempty"`
	Allocations []PaymentAllocation `json:"allocations,omitempty"`
	Method      string              `json:"method"`
	Amount      float64             `json:"amount"`
	PaidAt      time.Time           `json:"paid_at"`
	ProofFileID *string             `json:"proof_file_id,omitempty"`
}

func (r SubmitPaymentRequest) Validate(now time.Time) error {
	r.Method = strings.ToLower(strings.TrimSpace(r.Method))
	if r.Amount <= 0 || r.PaidAt.IsZero() || r.PaidAt.After(now.Add(24*time.Hour)) {
		return ErrValidation
	}
	if strings.TrimSpace(r.InvoiceID) == "" && len(r.Allocations) == 0 {
		return ErrValidation
	}
	if strings.TrimSpace(r.InvoiceID) != "" && len(r.Allocations) > 0 {
		return ErrValidation
	}
	if len(r.Allocations) > 0 {
		seen := make(map[string]struct{}, len(r.Allocations))
		var total float64
		for _, allocation := range r.Allocations {
			invoiceID := strings.TrimSpace(allocation.InvoiceID)
			if invoiceID == "" || allocation.Amount <= 0 {
				return ErrValidation
			}
			if _, duplicate := seen[invoiceID]; duplicate {
				return ErrValidation
			}
			seen[invoiceID] = struct{}{}
			total += allocation.Amount
		}
		if total != r.Amount {
			return ErrValidation
		}
	}

	switch r.Method {
	case "cash", "transfer", "other":
	default:
		return ErrValidation
	}

	if r.Method == "transfer" && (r.ProofFileID == nil || strings.TrimSpace(*r.ProofFileID) == "") {
		return ErrValidation
	}
	return nil
}

type SubmitPaymentResponse struct {
	ID                 string              `json:"id"`
	PaymentNumber      string              `json:"payment_number"`
	VerificationStatus string              `json:"verification_status"`
	InvoiceStatus      string              `json:"invoice_status,omitempty"`
	Allocations        []PaymentAllocation `json:"allocations,omitempty"`
	CreatedAt          time.Time           `json:"created_at"`
}

type VerifyPaymentRequest struct {
	Note string `json:"note"`
}

type RejectPaymentRequest struct {
	Reason string `json:"reason"`
}

func (r RejectPaymentRequest) Validate() error {
	if strings.TrimSpace(r.Reason) == "" {
		return ErrValidation
	}
	return nil
}

type CancelPaymentRequest struct {
	Reason string `json:"reason"`
}

func (r CancelPaymentRequest) Validate() error {
	if strings.TrimSpace(r.Reason) == "" {
		return ErrValidation
	}
	return nil
}

type PaymentActionResponse struct {
	ID                 string     `json:"id"`
	VerificationStatus string     `json:"verification_status"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	InvoiceStatus      string     `json:"invoice_status"`
}

type PaymentFilter struct {
	InvoiceID          string
	VerificationStatus string
}
