package complaints

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrValidation        = errors.New("validation error")
	ErrComplaintNotFound = errors.New("complaint not found")
	ErrInvalidState      = errors.New("invalid state")
	ErrForbidden         = errors.New("forbidden")
	ErrConflict          = errors.New("conflict")
)

type AttachmentInfo struct {
	AttachmentID string `json:"attachment_id"`
	FileID       string `json:"file_id"`
	OriginalName string `json:"original_name"`
	MIMEType     string `json:"mime_type"`
	SizeBytes    int64  `json:"size_bytes"`
	Purpose      string `json:"purpose"`
}

type CommentItem struct {
	ID           string    `json:"id"`
	ComplaintID  string    `json:"complaint_id"`
	AuthorUserID string    `json:"author_user_id"`
	AuthorName   string    `json:"author_name"`
	Body         string    `json:"body"`
	IsInternal   bool      `json:"is_internal"`
	CreatedAt    time.Time `json:"created_at"`
}

type ComplaintItem struct {
	ID                  string           `json:"id"`
	ReporterUserID      string           `json:"reporter_user_id"`
	ReporterName        string           `json:"reporter_name"`
	TicketNumber        string           `json:"ticket_number"`
	Category            string           `json:"category"`
	Title               string           `json:"title"`
	Description         string           `json:"description"`
	LocationDescription *string          `json:"location_description,omitempty"`
	Priority            string           `json:"priority"`
	Status              string           `json:"status"`
	AssignedTo          *string          `json:"assigned_to,omitempty"`
	AssignedToName      *string          `json:"assigned_to_name,omitempty"`
	ResolutionNote      *string          `json:"resolution_note,omitempty"`
	ResolvedAt          *time.Time       `json:"resolved_at,omitempty"`
	ClosedAt            *time.Time       `json:"closed_at,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
	Attachments         []AttachmentInfo `json:"attachments"`
	Comments            []CommentItem    `json:"comments"`
}

type CreateComplaintRequest struct {
	Category            string   `json:"category"`
	Title               string   `json:"title"`
	Description         string   `json:"description"`
	LocationDescription *string  `json:"location_description,omitempty"`
	Priority            string   `json:"priority"`
	AttachmentFileIDs   []string `json:"attachment_file_ids"`
}

func (r *CreateComplaintRequest) Validate() error {
	r.Category = strings.TrimSpace(r.Category)
	r.Title = strings.TrimSpace(r.Title)
	r.Description = strings.TrimSpace(r.Description)
	r.Priority = strings.ToLower(strings.TrimSpace(r.Priority))
	if r.LocationDescription != nil {
		location := strings.TrimSpace(*r.LocationDescription)
		r.LocationDescription = &location
	}
	if r.Category == "" || len(r.Category) > 50 || r.Title == "" || len(r.Title) > 255 || r.Description == "" {
		return ErrValidation
	}
	if r.Priority == "" {
		r.Priority = "normal"
	}
	if r.Priority != "low" && r.Priority != "normal" && r.Priority != "high" {
		return ErrValidation
	}
	return nil
}

type UpdateComplaintRequest CreateComplaintRequest

func (r *UpdateComplaintRequest) Validate() error {
	request := CreateComplaintRequest(*r)
	if err := request.Validate(); err != nil {
		return err
	}
	*r = UpdateComplaintRequest(request)
	return nil
}

type ComplaintFilter struct {
	Status     string
	Category   string
	AssignedTo string
	Search     string
}

type AssignComplaintRequest struct {
	AssignedTo string `json:"assigned_to"`
}

func (r *AssignComplaintRequest) Validate() error {
	r.AssignedTo = strings.TrimSpace(r.AssignedTo)
	if r.AssignedTo == "" {
		return ErrValidation
	}
	return nil
}

type UpdateStatusRequest struct {
	Status         string  `json:"status"`
	ResolutionNote *string `json:"resolution_note,omitempty"`
}

func (r *UpdateStatusRequest) Validate() error {
	r.Status = strings.ToLower(strings.TrimSpace(r.Status))
	if r.ResolutionNote != nil {
		note := strings.TrimSpace(*r.ResolutionNote)
		r.ResolutionNote = &note
	}
	switch r.Status {
	case "reviewed", "in_progress", "waiting_information", "resolved", "rejected", "closed":
	default:
		return ErrValidation
	}
	if (r.Status == "resolved" || r.Status == "closed") && (r.ResolutionNote == nil || *r.ResolutionNote == "") {
		return ErrValidation
	}
	return nil
}

type AddCommentRequest struct {
	Body       string `json:"body"`
	IsInternal bool   `json:"is_internal"`
}

func (r *AddCommentRequest) Validate() error {
	r.Body = strings.TrimSpace(r.Body)
	if r.Body == "" {
		return ErrValidation
	}
	return nil
}
