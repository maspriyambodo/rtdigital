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

type SubmitPaymentRequest struct {
	InvoiceID   string    `json:"invoice_id"`
	Method      string    `json:"method"`
	Amount      float64   `json:"amount"`
	PaidAt      time.Time `json:"paid_at"`
	ProofFileID *string   `json:"proof_file_id"`
}

func (r SubmitPaymentRequest) Validate(now time.Time) error {
	r.Method = strings.ToLower(strings.TrimSpace(r.Method))
	if strings.TrimSpace(r.InvoiceID) == "" ||
		r.Amount <= 0 ||
		r.PaidAt.IsZero() ||
		r.PaidAt.After(now.Add(24*time.Hour)) {
		return ErrValidation
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
	ID                 string    `json:"id"`
	PaymentNumber      string    `json:"payment_number"`
	VerificationStatus string    `json:"verification_status"`
	InvoiceStatus      string    `json:"invoice_status"`
	CreatedAt          time.Time `json:"created_at"`
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
