package savings

import (
	"time"
)

type Product struct {
	ID               string    `json:"id"`
	Code             string    `json:"code"`
	Name             string    `json:"name"`
	Description      *string   `json:"description,omitempty"`
	MinimumDeposit   float64   `json:"minimum_deposit"`
	WithdrawalRule   string    `json:"withdrawal_rule"`
	Status           string    `json:"status"`
	CreatedBy        string    `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Account struct {
	ID            string    `json:"id"`
	ProductID     string    `json:"product_id"`
	HouseholdID   string    `json:"household_id"`
	AccountNumber string    `json:"account_number"`
	Balance       float64   `json:"balance"`
	Status        string    `json:"status"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Transaction struct {
	ID                 string    `json:"id"`
	AccountID          string    `json:"account_id"`
	TransactionNumber  string    `json:"transaction_number"`
	Type               string    `json:"type"`
	Amount             float64   `json:"amount"`
	BalanceAfter       float64   `json:"balance_after"`
	TransactionDate    string    `json:"transaction_date"` // YYYY-MM-DD
	Description        string    `json:"description"`
	ProofFileID        *string   `json:"proof_file_id,omitempty"`
	ReversalOfID       *string   `json:"reversal_of_id,omitempty"`
	VerificationStatus string    `json:"verification_status"`
	VerifiedBy         *string   `json:"verified_by,omitempty"`
	VerifiedAt         *time.Time`json:"verified_at,omitempty"`
	RejectionReason    *string   `json:"rejection_reason,omitempty"`
	IdempotencyKey     *string   `json:"idempotency_key,omitempty"`
	CreatedBy          string    `json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type CreateProductReq struct {
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Description    *string `json:"description,omitempty"`
	MinimumDeposit float64 `json:"minimum_deposit"`
	WithdrawalRule string  `json:"withdrawal_rule"`
}

type UpdateProductReq struct {
	Name           *string  `json:"name,omitempty"`
	Description    *string  `json:"description,omitempty"`
	MinimumDeposit *float64 `json:"minimum_deposit,omitempty"`
	WithdrawalRule *string  `json:"withdrawal_rule,omitempty"`
	Status         *string  `json:"status,omitempty"`
}

type CreateAccountReq struct {
	ProductID   string `json:"product_id"`
	HouseholdID string `json:"household_id"`
}

type DepositReq struct {
	AccountID       string  `json:"account_id"`
	Amount          float64 `json:"amount"`
	TransactionDate string  `json:"transaction_date"`
	Description     string  `json:"description"`
	ProofFileID     *string `json:"proof_file_id,omitempty"`
}

type WithdrawReq struct {
	AccountID       string  `json:"account_id"`
	Amount          float64 `json:"amount"`
	TransactionDate string  `json:"transaction_date"`
	Description     string  `json:"description"`
	ProofFileID     *string `json:"proof_file_id,omitempty"`
}

type VerifyTxReq struct {
	Status          string  `json:"status"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
}
