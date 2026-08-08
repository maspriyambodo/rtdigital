package savings

import (
	"errors"
	"time"
)

var (
	ErrValidation          = errors.New("savings validation failed")
	ErrProductNotFound     = errors.New("savings product not found")
	ErrAccountNotFound     = errors.New("savings account not found")
	ErrTransactionNotFound = errors.New("savings transaction not found")
	ErrDuplicateData       = errors.New("savings duplicate data")
	ErrConstraint          = errors.New("savings constraint violation")
	ErrInvalidState        = errors.New("savings invalid state")
	ErrForbidden           = errors.New("savings forbidden")
	ErrInsufficientBalance = errors.New("savings insufficient balance")
)

type Product struct {
	ID             string    `json:"id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	Description    *string   `json:"description,omitempty"`
	MinimumDeposit float64   `json:"minimum_deposit"`
	WithdrawalRule string    `json:"withdrawal_rule"`
	Status         string    `json:"status"`
	CreatedBy      string    `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Account struct {
	ID            string    `json:"id"`
	ProductID     string    `json:"product_id"`
	ProductName   string    `json:"product_name,omitempty"`
	HouseholdID   string    `json:"household_id"`
	AccountNumber string    `json:"account_number"`
	Balance       float64   `json:"balance"`
	Status        string    `json:"status"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Transaction struct {
	ID                 string     `json:"id"`
	AccountID          string     `json:"account_id"`
	TransactionNumber  string     `json:"transaction_number"`
	Type               string     `json:"type"`
	Amount             float64    `json:"amount"`
	BalanceAfter       float64    `json:"balance_after"`
	TransactionDate    string     `json:"transaction_date"` // YYYY-MM-DD
	Description        string     `json:"description"`
	ProofFileID        *string    `json:"proof_file_id,omitempty"`
	ReversalOfID       *string    `json:"reversal_of_id,omitempty"`
	VerificationStatus string     `json:"verification_status"`
	VerifiedBy         *string    `json:"verified_by,omitempty"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	RejectionReason    *string    `json:"rejection_reason,omitempty"`
	IdempotencyKey     *string    `json:"idempotency_key,omitempty"`
	CreatedBy          string     `json:"created_by"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type CreateProductReq struct {
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Description    *string `json:"description,omitempty"`
	MinimumDeposit float64 `json:"minimum_deposit"`
	WithdrawalRule string  `json:"withdrawal_rule"`
}

func (r CreateProductReq) Validate() error {
	if r.Code == "" || r.Name == "" {
		return ErrValidation
	}
	if r.MinimumDeposit < 0 {
		return ErrValidation
	}
	if r.WithdrawalRule == "" {
		return ErrValidation
	}
	return nil
}

type UpdateProductReq struct {
	Name           *string  `json:"name,omitempty"`
	Description    *string  `json:"description,omitempty"`
	MinimumDeposit *float64 `json:"minimum_deposit,omitempty"`
	WithdrawalRule *string  `json:"withdrawal_rule,omitempty"`
	Status         *string  `json:"status,omitempty"`
}

func (r UpdateProductReq) Validate() error {
	if r.MinimumDeposit != nil && *r.MinimumDeposit < 0 {
		return ErrValidation
	}
	if r.Status != nil && *r.Status != "active" && *r.Status != "inactive" {
		return ErrValidation
	}
	return nil
}

type CreateAccountReq struct {
	ProductID   string `json:"product_id"`
	HouseholdID string `json:"household_id"`
}

func (r CreateAccountReq) Validate() error {
	if r.ProductID == "" || r.HouseholdID == "" {
		return ErrValidation
	}
	return nil
}

type DepositReq struct {
	AccountID       string  `json:"account_id"`
	Amount          float64 `json:"amount"`
	TransactionDate string  `json:"transaction_date"`
	Description     string  `json:"description"`
	ProofFileID     *string `json:"proof_file_id,omitempty"`
	IdempotencyKey  *string `json:"idempotency_key,omitempty"`
}

func (r DepositReq) Validate() error {
	if r.AccountID == "" || r.Amount <= 0 || r.TransactionDate == "" || r.Description == "" {
		return ErrValidation
	}
	return nil
}

type WithdrawReq struct {
	AccountID       string  `json:"account_id"`
	Amount          float64 `json:"amount"`
	TransactionDate string  `json:"transaction_date"`
	Description     string  `json:"description"`
	ProofFileID     *string `json:"proof_file_id,omitempty"`
	IdempotencyKey  *string `json:"idempotency_key,omitempty"`
}

func (r WithdrawReq) Validate() error {
	if r.AccountID == "" || r.Amount <= 0 || r.TransactionDate == "" || r.Description == "" {
		return ErrValidation
	}
	return nil
}

type VerifyTxReq struct {
	Status          string  `json:"status"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
}

func (r VerifyTxReq) Validate() error {
	if r.Status != "verified" && r.Status != "rejected" {
		return ErrValidation
	}
	if r.Status == "rejected" && (r.RejectionReason == nil || *r.RejectionReason == "") {
		return ErrValidation
	}
	return nil
}
type ProductFilter struct {
	Status string
}

type AccountFilter struct {
	ProductID   string
	HouseholdID string
	Status      string
}


type TransactionFilter struct {
	AccountID          string
	ProductID          string
	HouseholdID        string
	Type               string
	VerificationStatus string
	StartDate          string
	EndDate            string
}

type ReconciliationReport struct {
	ProductID            string  `json:"product_id"`
	ProductName          string  `json:"product_name"`
	TotalAccounts        int     `json:"total_accounts"`
	TotalVerifiedBalance float64 `json:"total_verified_balance"`
	PendingDeposits      float64 `json:"pending_deposits"`
	PendingWithdrawals   float64 `json:"pending_withdrawals"`
}

