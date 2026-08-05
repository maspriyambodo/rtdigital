package dashboard

type ResidentDashboard struct {
	ActiveInvoices    []InvoiceSummary      `json:"active_invoices"`
	RecentPayments    []PaymentSummary      `json:"recent_payments"`
	RecentLetters     []LetterSummary       `json:"recent_letters"`
	RecentComplaints  []ComplaintSummary    `json:"recent_complaints"`
	Announcements     []AnnouncementSummary `json:"announcements"`
	UpcomingEvents    []EventSummary        `json:"upcoming_events"`
	PublicCashSummary *PublicCashSummary    `json:"public_cash_summary,omitempty"`
}

type AdminDashboard struct {
	ActiveHouseholds  int                `json:"active_households"`
	ActiveResidents   int                `json:"active_residents"`
	ActiveInvoices    int                `json:"active_invoices"`
	OutstandingAmount float64            `json:"outstanding_amount"`
	PendingPayments   int                `json:"pending_payments"`
	CashBalance       float64            `json:"cash_balance"`
	PendingLetters    int                `json:"pending_letters"`
	OpenComplaints    int                `json:"open_complaints"`
	RecentPayments    []PaymentSummary   `json:"recent_payments"`
	RecentLetters     []LetterSummary    `json:"recent_letters"`
	RecentComplaints  []ComplaintSummary `json:"recent_complaints"`
}

type InvoiceSummary struct {
	ID            string  `json:"id"`
	InvoiceNumber string  `json:"invoice_number"`
	DueTypeName   string  `json:"due_type_name"`
	Amount        float64 `json:"amount"`
	PaidAmount    float64 `json:"paid_amount"`
	DueDate       string  `json:"due_date"`
	Status        string  `json:"status"`
}

type PaymentSummary struct {
	ID                 string  `json:"id"`
	PaymentNumber      string  `json:"payment_number"`
	InvoiceNumber      string  `json:"invoice_number"`
	Amount             float64 `json:"amount"`
	PaidAt             string  `json:"paid_at"`
	VerificationStatus string  `json:"verification_status"`
}

type LetterSummary struct {
	ID            string `json:"id"`
	RequestNumber string `json:"request_number"`
	LetterType    string `json:"letter_type"`
	Status        string `json:"status"`
	UpdatedAt     string `json:"updated_at"`
}

type ComplaintSummary struct {
	ID           string `json:"id"`
	TicketNumber string `json:"ticket_number"`
	Title        string `json:"title"`
	Priority     string `json:"priority"`
	Status       string `json:"status"`
	UpdatedAt    string `json:"updated_at"`
}

type AnnouncementSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Category  string `json:"category"`
	Priority  string `json:"priority"`
	PublishAt string `json:"publish_at"`
}

type EventSummary struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	StartsAt string `json:"starts_at"`
	Location string `json:"location"`
}

// PublicCashSummary contains only administrator-published aggregate values.
// It intentionally excludes transaction, payer, and proof-document data.
type PublicCashSummary struct {
	PeriodStart   string                      `json:"period_start"`
	PeriodEnd     string                      `json:"period_end"`
	TotalIncome   float64                     `json:"total_income"`
	TotalExpense  float64                     `json:"total_expense"`
	EndingBalance float64                     `json:"ending_balance"`
	Categories    []PublicCashCategorySummary `json:"categories"`
}

type PublicCashCategorySummary struct {
	CategoryName    string  `json:"category_name"`
	TransactionType string  `json:"transaction_type"`
	TotalAmount     float64 `json:"total_amount"`
}
