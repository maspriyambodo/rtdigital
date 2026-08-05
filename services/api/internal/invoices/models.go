package invoices

import "time"

type DueType struct {
	ID                         string    `json:"id"`
	Name                       string    `json:"name"`
	Description                *string   `json:"description"`
	Amount                     *float64  `json:"amount"`
	Frequency                  string    `json:"frequency"`
	DueDay                     *int      `json:"due_day"`
	Status                     string    `json:"status"`
	AutomaticGenerationEnabled bool      `json:"automatic_generation_enabled"`
	GenerationLeadDays         int       `json:"generation_lead_days"`
	ReminderLeadDays           int       `json:"reminder_lead_days"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
}

type Invoice struct {
	ID                 string     `json:"id"`
	HouseholdID        string     `json:"household_id"`
	HouseholdNumber    string     `json:"household_number"`
	HouseUnitCode      string     `json:"house_unit_code"`
	DueTypeID          string     `json:"due_type_id"`
	DueTypeName        string     `json:"due_type_name"`
	InvoiceNumber      string     `json:"invoice_number"`
	PeriodStart        string     `json:"period_start"`
	PeriodEnd          string     `json:"period_end"`
	DueDate            string     `json:"due_date"`
	Amount             float64    `json:"amount"`
	PaidAmount         float64    `json:"paid_amount"`
	AdjustmentAmount   float64    `json:"adjustment_amount"`
	AdjustmentReason   *string    `json:"adjustment_reason"`
	Status             string     `json:"status"`
	CancelledBy        *string    `json:"cancelled_by,omitempty"`
	CancelledAt        *time.Time `json:"cancelled_at,omitempty"`
	CancellationReason *string    `json:"cancellation_reason,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type CreateDueTypeRequest struct {
	Name                       string   `json:"name"`
	Description                *string  `json:"description"`
	Amount                     *float64 `json:"amount"`
	Frequency                  string   `json:"frequency"`
	DueDay                     *int     `json:"due_day"`
	AutomaticGenerationEnabled *bool    `json:"automatic_generation_enabled"`
	GenerationLeadDays         *int     `json:"generation_lead_days"`
	ReminderLeadDays           *int     `json:"reminder_lead_days"`
}

type UpdateDueTypeRequest struct {
	Name                       *string  `json:"name"`
	Description                *string  `json:"description"`
	Amount                     *float64 `json:"amount"`
	Frequency                  *string  `json:"frequency"`
	DueDay                     *int     `json:"due_day"`
	AutomaticGenerationEnabled *bool    `json:"automatic_generation_enabled"`
	GenerationLeadDays         *int     `json:"generation_lead_days"`
	ReminderLeadDays           *int     `json:"reminder_lead_days"`
}

type CreateInvoiceRequest struct {
	HouseholdID      string   `json:"household_id"`
	DueTypeID        string   `json:"due_type_id"`
	PeriodStart      string   `json:"period_start"`
	PeriodEnd        string   `json:"period_end"`
	DueDate          string   `json:"due_date"`
	Amount           *float64 `json:"amount"`
	AdjustmentAmount *float64 `json:"adjustment_amount"`
	AdjustmentReason *string  `json:"adjustment_reason"`
}

type GenerateInvoicesRequest struct {
	DueTypeID        string   `json:"due_type_id"`
	HouseholdIDs     []string `json:"household_ids"`
	PeriodStart      string   `json:"period_start"`
	PeriodEnd        string   `json:"period_end"`
	DueDate          string   `json:"due_date"`
	Amount           *float64 `json:"amount"`
	AdjustmentAmount *float64 `json:"adjustment_amount"`
	AdjustmentReason *string  `json:"adjustment_reason"`
}

type GenerateInvoicesResult struct {
	TotalTargeted int       `json:"total_targeted"`
	TotalCreated  int       `json:"total_created"`
	TotalSkipped  int       `json:"total_skipped"`
	Invoices      []Invoice `json:"invoices"`
}

type UpdateInvoiceRequest struct {
	Amount           *float64 `json:"amount"`
	AdjustmentAmount *float64 `json:"adjustment_amount"`
	AdjustmentReason *string  `json:"adjustment_reason"`
	DueDate          *string  `json:"due_date"`
}

type CancelInvoiceRequest struct {
	Reason string `json:"reason"`
}

type InvoiceFilter struct {
	Status      string
	DueTypeID   string
	HouseholdID string
	PeriodStart string
	PeriodEnd   string
	OnlyArrears bool
}

type InvoiceGenerationRun struct {
	ID            string     `json:"id"`
	DueTypeID     string     `json:"due_type_id"`
	PeriodStart   string     `json:"period_start"`
	PeriodEnd     string     `json:"period_end"`
	DueDate       string     `json:"due_date"`
	RunKey        string     `json:"run_key"`
	Status        string     `json:"status"`
	TotalTargeted int        `json:"total_targeted"`
	TotalCreated  int        `json:"total_created"`
	TotalSkipped  int        `json:"total_skipped"`
	ErrorMessage  *string    `json:"error_message,omitempty"`
	StartedAt     time.Time  `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CreatedBy     *string    `json:"created_by,omitempty"`
}
