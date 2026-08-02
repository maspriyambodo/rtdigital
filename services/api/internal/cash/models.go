package cash

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrValidation          = errors.New("invalid request data")
	ErrCategoryNotFound    = errors.New("cash category not found")
	ErrTransactionNotFound = errors.New("cash transaction not found")
	ErrDuplicateData       = errors.New("duplicate data")
	ErrConstraint          = errors.New("database constraint violation")
	ErrForbidden           = errors.New("forbidden")
	ErrInvalidState        = errors.New("invalid cash transaction state")
)

type Category struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (r *CreateCategoryRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" || (r.Type != "income" && r.Type != "expense") {
		return ErrValidation
	}
	return nil
}

type UpdateCategoryRequest struct {
	Name   *string `json:"name,omitempty"`
	Status *string `json:"status,omitempty"`
}

func (r *UpdateCategoryRequest) Validate() error {
	if r.Name == nil && r.Status == nil {
		return ErrValidation
	}
	if r.Name != nil {
		name := strings.TrimSpace(*r.Name)
		if name == "" {
			return ErrValidation
		}
		r.Name = &name
	}
	if r.Status != nil && *r.Status != "active" && *r.Status != "inactive" {
		return ErrValidation
	}
	return nil
}

type Transaction struct {
	ID                string    `json:"id"`
	TransactionNumber string    `json:"transaction_number"`
	Type              string    `json:"type"`
	CategoryID        *string   `json:"category_id,omitempty"`
	CategoryName      *string   `json:"category_name,omitempty"`
	Amount            float64   `json:"amount"`
	TransactionDate   string    `json:"transaction_date"`
	Description       string    `json:"description"`
	ProofFileID       *string   `json:"proof_file_id,omitempty"`
	ReferenceType     *string   `json:"reference_type,omitempty"`
	ReferenceID       *string   `json:"reference_id,omitempty"`
	ReversalOfID      *string   `json:"reversal_of_id,omitempty"`
	Status            string    `json:"status"`
	CreatedBy         string    `json:"created_by"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	RunningBalance    float64   `json:"running_balance"`
}

type TransactionFilter struct {
	StartDate  string
	EndDate    string
	Type       string
	CategoryID string
}

func (f TransactionFilter) Validate() error {
	if f.Type != "" && f.Type != "income" && f.Type != "expense" {
		return ErrValidation
	}
	for _, date := range []string{f.StartDate, f.EndDate} {
		if date != "" {
			if _, err := time.Parse(time.DateOnly, date); err != nil {
				return ErrValidation
			}
		}
	}
	if f.StartDate != "" && f.EndDate != "" && f.StartDate > f.EndDate {
		return ErrValidation
	}
	return nil
}

type RecordTransactionRequest struct {
	Type            string  `json:"type"`
	CategoryID      string  `json:"category_id"`
	Amount          float64 `json:"amount"`
	TransactionDate string  `json:"transaction_date"`
	Description     string  `json:"description"`
	ProofFileID     *string `json:"proof_file_id,omitempty"`
}

func (r *RecordTransactionRequest) Validate() error {
	r.CategoryID = strings.TrimSpace(r.CategoryID)
	r.Description = strings.TrimSpace(r.Description)
	if r.Type != "income" && r.Type != "expense" {
		return ErrValidation
	}
	if r.CategoryID == "" || r.Amount <= 0 || r.Description == "" {
		return ErrValidation
	}
	if _, err := time.Parse(time.DateOnly, r.TransactionDate); err != nil {
		return ErrValidation
	}
	return nil
}

type ReverseTransactionRequest struct {
	Reason string `json:"reason"`
}

func (r *ReverseTransactionRequest) Validate() error {
	r.Reason = strings.TrimSpace(r.Reason)
	if r.Reason == "" {
		return ErrValidation
	}
	return nil
}

type CashBook struct {
	Transactions []Transaction `json:"transactions"`
	TotalIncome  float64       `json:"total_income"`
	TotalExpense float64       `json:"total_expense"`
	Balance      float64       `json:"balance"`
}
