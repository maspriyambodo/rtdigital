package letters

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrValidation            = errors.New("validation error")
	ErrLetterTypeNotFound    = errors.New("letter type not found")
	ErrLetterRequestNotFound = errors.New("letter request not found")
	ErrInvalidState          = errors.New("invalid state")
	ErrForbidden             = errors.New("forbidden")
	ErrConflict              = errors.New("conflict")
	ErrConstraint            = errors.New("business constraint violated")
	ErrStorage               = errors.New("storage error")
)

type LetterTypeItem struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Requirements  json.RawMessage `json:"requirements"`
	FormSchema    json.RawMessage `json:"form_schema"`
	Template      string          `json:"template"`
	NumberPattern string          `json:"number_pattern"`
	Status        string          `json:"status"`
	SLAHours      *int            `json:"sla_hours,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type CreateLetterTypeRequest struct {
	Name          string          `json:"name"`
	Requirements  json.RawMessage `json:"requirements"`
	FormSchema    json.RawMessage `json:"form_schema"`
	Template      string          `json:"template"`
	NumberPattern string          `json:"number_pattern"`
	Status        string          `json:"status"`
	SLAHours      *int            `json:"sla_hours"`
}

func (r *CreateLetterTypeRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.Template = strings.TrimSpace(r.Template)
	r.NumberPattern = strings.TrimSpace(r.NumberPattern)
	r.Status = strings.TrimSpace(r.Status)
	if r.Name == "" || len(r.Name) > 100 || r.Template == "" || r.NumberPattern == "" || len(r.NumberPattern) > 100 {
		return ErrValidation
	}
	if len(r.Requirements) == 0 {
		r.Requirements = json.RawMessage("[]")
	}
	if len(r.FormSchema) == 0 {
		r.FormSchema = json.RawMessage("{}")
	}
	if !json.Valid(r.Requirements) || !json.Valid(r.FormSchema) {
		return ErrValidation
	}
	if r.Status == "" {
		r.Status = "active"
	}
	if r.Status != "active" && r.Status != "inactive" {
		return ErrValidation
	}
	if r.SLAHours != nil && (*r.SLAHours < 1 || *r.SLAHours > 8_760) {
		return ErrValidation
	}
	return nil
}

type UpdateLetterTypeRequest CreateLetterTypeRequest

func (r *UpdateLetterTypeRequest) Validate() error {
	request := CreateLetterTypeRequest(*r)
	if err := request.Validate(); err != nil {
		return err
	}
	*r = UpdateLetterTypeRequest(request)
	return nil
}

type AttachmentInfo struct {
	AttachmentID string `json:"attachment_id"`
	FileID       string `json:"file_id"`
	OriginalName string `json:"original_name"`
	MIMEType     string `json:"mime_type"`
	SizeBytes    int64  `json:"size_bytes"`
	Purpose      string `json:"purpose"`
}

type PublicLetterVerification struct {
	LetterNumber string     `json:"letter_number,omitempty"`
	LetterType   string     `json:"letter_type"`
	IssuedAt     *time.Time `json:"issued_at,omitempty"`
	Status       string     `json:"status"`
}

type LetterRequestItem struct {
	ID              string           `json:"id"`
	RequesterUserID string           `json:"requester_user_id"`
	RequesterName   string           `json:"requester_name"`
	ResidentID      string           `json:"resident_id"`
	ResidentName    string           `json:"resident_name"`
	LetterTypeID    string           `json:"letter_type_id"`
	LetterTypeName  string           `json:"letter_type_name"`
	RequestNumber   string           `json:"request_number"`
	LetterNumber    *string          `json:"letter_number,omitempty"`
	FormData        json.RawMessage  `json:"form_data"`
	Status          string           `json:"status"`
	ResidentNote    *string          `json:"resident_note,omitempty"`
	InternalNote    *string          `json:"internal_note,omitempty"`
	SubmittedAt     *time.Time       `json:"submitted_at,omitempty"`
	ProcessedBy     *string          `json:"processed_by,omitempty"`
	ApprovedBy      *string          `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time       `json:"approved_at,omitempty"`
	IssuedFileID    *string          `json:"issued_file_id,omitempty"`
	IssuedAt        *time.Time       `json:"issued_at,omitempty"`
	SLAHours        *int             `json:"sla_hours,omitempty"`
	SLADueAt        *time.Time       `json:"sla_due_at,omitempty"`
	SLAEscalatedAt  *time.Time       `json:"sla_escalated_at,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	Attachments     []AttachmentInfo `json:"attachments"`
}

type SubmitLetterRequest struct {
	LetterTypeID      string          `json:"letter_type_id"`
	ResidentID        string          `json:"resident_id"`
	FormData          json.RawMessage `json:"form_data"`
	AttachmentFileIDs []string        `json:"attachment_file_ids"`
	ResidentNote      *string         `json:"resident_note,omitempty"`
}

func (r *SubmitLetterRequest) Validate() error {
	r.LetterTypeID = strings.TrimSpace(r.LetterTypeID)
	r.ResidentID = strings.TrimSpace(r.ResidentID)
	if r.LetterTypeID == "" || r.ResidentID == "" {
		return ErrValidation
	}
	if len(r.FormData) == 0 {
		r.FormData = json.RawMessage("{}")
	}
	if !json.Valid(r.FormData) {
		return ErrValidation
	}
	if r.ResidentNote != nil {
		note := strings.TrimSpace(*r.ResidentNote)
		r.ResidentNote = &note
	}
	return nil
}

type UpdateLetterRequest SubmitLetterRequest

func (r *UpdateLetterRequest) Validate() error {
	request := SubmitLetterRequest(*r)
	if err := request.Validate(); err != nil {
		return err
	}
	*r = UpdateLetterRequest(request)
	return nil
}

type LetterRequestFilter struct {
	Status       string
	LetterTypeID string
	Search       string
}

type ReviewLetterRequest struct {
	ResidentNote *string `json:"resident_note,omitempty"`
	InternalNote *string `json:"internal_note,omitempty"`
}

type CancelLetterRequest struct {
	Reason string `json:"reason"`
}

func (r *CancelLetterRequest) Validate() error {
	r.Reason = strings.TrimSpace(r.Reason)
	if r.Reason == "" {
		return ErrValidation
	}
	return nil
}

func (r *ReviewLetterRequest) Validate(requireResidentNote bool) error {
	if r.ResidentNote != nil {
		note := strings.TrimSpace(*r.ResidentNote)
		r.ResidentNote = &note
	}
	if r.InternalNote != nil {
		note := strings.TrimSpace(*r.InternalNote)
		r.InternalNote = &note
	}
	if requireResidentNote && (r.ResidentNote == nil || *r.ResidentNote == "") {
		return ErrValidation
	}
	return nil
}
