package reports

import "errors"

var (
	ErrValidation = errors.New("validation failed")
	ErrForbidden  = errors.New("forbidden")
)

type Filter struct {
	StartDate string
	EndDate   string
	Status    string
}

type ResidentReportItem struct {
	ID                 string  `json:"id"`
	FullName           string  `json:"full_name"`
	Gender             *string `json:"gender"`
	BirthDate          *string `json:"birth_date"`
	ResidentStatus     string  `json:"resident_status"`
	VerificationStatus string  `json:"verification_status"`
	HouseUnitCode      *string `json:"house_unit_code"`
	HouseholdNumber    *string `json:"household_number"`
	Relationship       *string `json:"relationship"`
	CreatedAt          string  `json:"created_at"`
}

type MutationReportItem struct {
	ID              string `json:"id"`
	FullName        string `json:"full_name"`
	HouseholdNumber string `json:"household_number"`
	HouseUnitCode   string `json:"house_unit_code"`
	Relationship    string `json:"relationship"`
	StartedAt       string `json:"started_at"`
	EndedAt         string `json:"ended_at"`
	Status          string `json:"status"`
}

type ArrearsReportItem struct {
	HouseholdNumber  string  `json:"household_number"`
	HouseUnitCode    string  `json:"house_unit_code"`
	HeadResidentName string  `json:"head_resident_name"`
	InvoiceCount     int     `json:"invoice_count"`
	TotalArrears     float64 `json:"total_arrears"`
}

type HouseholdReportItem struct {
	ID                 string `json:"id"`
	InternalNumber     string `json:"internal_number"`
	HouseUnitCode      string `json:"house_unit_code"`
	HeadResidentName   string `json:"head_resident_name"`
	DomicileStatus     string `json:"domicile_status"`
	VerificationStatus string `json:"verification_status"`
	ActiveMembersCount int    `json:"active_members_count"`
	MoveInDate         string `json:"move_in_date"`
}

type InvoiceReportItem struct {
	ID               string  `json:"id"`
	InvoiceNumber    string  `json:"invoice_number"`
	HouseholdNumber  string  `json:"household_number"`
	HouseUnitCode    string  `json:"house_unit_code"`
	DueTypeName      string  `json:"due_type_name"`
	PeriodStart      string  `json:"period_start"`
	PeriodEnd        string  `json:"period_end"`
	DueDate          string  `json:"due_date"`
	Amount           float64 `json:"amount"`
	PaidAmount       float64 `json:"paid_amount"`
	AdjustmentAmount float64 `json:"adjustment_amount"`
	Status           string  `json:"status"`
}

type PaymentReportItem struct {
	ID                 string  `json:"id"`
	PaymentNumber      string  `json:"payment_number"`
	InvoiceNumber      string  `json:"invoice_number"`
	HouseholdNumber    string  `json:"household_number"`
	Method             string  `json:"method"`
	Amount             float64 `json:"amount"`
	PaidAt             string  `json:"paid_at"`
	VerificationStatus string  `json:"verification_status"`
	VerifiedAt         *string `json:"verified_at,omitempty"`
}

type CashReportItem struct {
	ID                string  `json:"id"`
	TransactionNumber string  `json:"transaction_number"`
	Type              string  `json:"type"`
	CategoryName      *string `json:"category_name"`
	Amount            float64 `json:"amount"`
	TransactionDate   string  `json:"transaction_date"`
	Description       string  `json:"description"`
	Status            string  `json:"status"`
}

type LetterReportItem struct {
	ID             string  `json:"id"`
	RequestNumber  string  `json:"request_number"`
	LetterNumber   *string `json:"letter_number,omitempty"`
	LetterTypeName string  `json:"letter_type_name"`
	RequesterName  string  `json:"requester_name"`
	ResidentName   string  `json:"resident_name"`
	Status         string  `json:"status"`
	SubmittedAt    *string `json:"submitted_at,omitempty"`
	IssuedAt       *string `json:"issued_at,omitempty"`
}

type ComplaintReportItem struct {
	ID             string  `json:"id"`
	TicketNumber   string  `json:"ticket_number"`
	Category       string  `json:"category"`
	Title          string  `json:"title"`
	Priority       string  `json:"priority"`
	Status         string  `json:"status"`
	ReporterName   string  `json:"reporter_name"`
	AssignedToName *string `json:"assigned_to_name,omitempty"`
	CreatedAt      string  `json:"created_at"`
	ResolvedAt     *string `json:"resolved_at,omitempty"`
}