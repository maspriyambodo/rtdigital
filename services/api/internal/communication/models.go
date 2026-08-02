package communication

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrValidation           = errors.New("validation error")
	ErrAnnouncementNotFound = errors.New("announcement not found")
	ErrEventNotFound        = errors.New("event not found")
	ErrInvalidState         = errors.New("invalid state")
	ErrForbidden            = errors.New("forbidden")
	ErrConflict             = errors.New("conflict")
)

type TargetInput struct {
	TargetType string  `json:"target_type"`
	TargetID   *string `json:"target_id,omitempty"`
}

type AttachmentInfo struct {
	AttachmentID string `json:"attachment_id"`
	FileID       string `json:"file_id"`
	OriginalName string `json:"original_name"`
	MIMEType     string `json:"mime_type"`
	SizeBytes    int64  `json:"size_bytes"`
	Purpose      string `json:"purpose"`
}

type AnnouncementTargetInfo struct {
	TargetType string  `json:"target_type"`
	TargetID   *string `json:"target_id,omitempty"`
	TargetName *string `json:"target_name,omitempty"`
}

type AnnouncementItem struct {
	ID           string                   `json:"id"`
	Title        string                   `json:"title"`
	Content      string                   `json:"content"`
	Category     string                   `json:"category"`
	Priority     string                   `json:"priority"`
	PublishAt    *time.Time               `json:"publish_at,omitempty"`
	ExpireAt     *time.Time               `json:"expire_at,omitempty"`
	Status       string                   `json:"status"`
	AuthorUserID string                   `json:"author_user_id"`
	AuthorName   string                   `json:"author_name"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
	IsRead       bool                     `json:"is_read"`
	ReadCount    int                      `json:"read_count"`
	Targets      []AnnouncementTargetInfo `json:"targets"`
	Attachments  []AttachmentInfo         `json:"attachments"`
}

type CreateAnnouncementRequest struct {
	Title             string        `json:"title"`
	Content           string        `json:"content"`
	Category          string        `json:"category"`
	Priority          string        `json:"priority"`
	PublishAt         *time.Time    `json:"publish_at,omitempty"`
	ExpireAt          *time.Time    `json:"expire_at,omitempty"`
	Status            string        `json:"status"`
	Targets           []TargetInput `json:"targets"`
	AttachmentFileIDs []string      `json:"attachment_file_ids"`
}

func (r *CreateAnnouncementRequest) Validate() error {
	r.Title = strings.TrimSpace(r.Title)
	r.Content = strings.TrimSpace(r.Content)
	r.Category = strings.TrimSpace(r.Category)
	r.Priority = strings.TrimSpace(r.Priority)
	r.Status = strings.TrimSpace(r.Status)

	if r.Title == "" || len(r.Title) > 255 || r.Content == "" {
		return ErrValidation
	}
	switch r.Category {
	case "general", "security", "health", "billing", "event", "emergency":
	default:
		return ErrValidation
	}
	if r.Priority == "" {
		r.Priority = "normal"
	}
	if r.Priority != "normal" && r.Priority != "important" {
		return ErrValidation
	}
	if r.Status == "" {
		r.Status = "draft"
	}
	switch r.Status {
	case "draft", "scheduled", "published":
	default:
		return ErrValidation
	}
	if r.Status == "scheduled" && r.PublishAt == nil {
		return ErrValidation
	}
	if r.PublishAt != nil && r.ExpireAt != nil && !r.ExpireAt.After(*r.PublishAt) {
		return ErrValidation
	}
	if len(r.Targets) == 0 {
		r.Targets = []TargetInput{{TargetType: "all"}}
	}
	for i := range r.Targets {
		target := &r.Targets[i]
		target.TargetType = strings.TrimSpace(target.TargetType)
		switch target.TargetType {
		case "all":
			target.TargetID = nil
		case "role", "household", "house_unit":
			if target.TargetID == nil || strings.TrimSpace(*target.TargetID) == "" {
				return ErrValidation
			}
			id := strings.TrimSpace(*target.TargetID)
			target.TargetID = &id
		default:
			return ErrValidation
		}
	}
	return nil
}

type UpdateAnnouncementRequest CreateAnnouncementRequest

func (r *UpdateAnnouncementRequest) Validate() error {
	request := CreateAnnouncementRequest(*r)
	if err := request.Validate(); err != nil {
		return err
	}
	*r = UpdateAnnouncementRequest(request)
	return nil
}

type ReadStats struct {
	TotalAudience int `json:"total_audience"`
	ReadCount     int `json:"read_count"`
}

type EventItem struct {
	ID           string           `json:"id"`
	Title        string           `json:"title"`
	Description  *string          `json:"description,omitempty"`
	Location     *string          `json:"location,omitempty"`
	StartsAt     time.Time        `json:"starts_at"`
	EndsAt       *time.Time       `json:"ends_at,omitempty"`
	Status       string           `json:"status"`
	AuthorUserID string           `json:"author_user_id"`
	AuthorName   string           `json:"author_name"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	Attachments  []AttachmentInfo `json:"attachments"`
}

type CreateEventRequest struct {
	Title             string     `json:"title"`
	Description       *string    `json:"description,omitempty"`
	Location          *string    `json:"location,omitempty"`
	StartsAt          time.Time  `json:"starts_at"`
	EndsAt            *time.Time `json:"ends_at,omitempty"`
	Status            string     `json:"status"`
	AttachmentFileIDs []string   `json:"attachment_file_ids"`
}

func (r *CreateEventRequest) Validate() error {
	r.Title = strings.TrimSpace(r.Title)
	if r.Title == "" || len(r.Title) > 255 || r.StartsAt.IsZero() {
		return ErrValidation
	}
	if r.Description != nil {
		value := strings.TrimSpace(*r.Description)
		if value == "" {
			r.Description = nil
		} else {
			r.Description = &value
		}
	}
	if r.Location != nil {
		value := strings.TrimSpace(*r.Location)
		if value == "" {
			r.Location = nil
		} else {
			r.Location = &value
		}
	}
	if r.EndsAt != nil && !r.EndsAt.After(r.StartsAt) {
		return ErrValidation
	}
	r.Status = strings.TrimSpace(r.Status)
	if r.Status == "" {
		r.Status = "planned"
	}
	switch r.Status {
	case "planned", "ongoing", "completed", "cancelled":
	default:
		return ErrValidation
	}
	return nil
}

type UpdateEventRequest CreateEventRequest

func (r *UpdateEventRequest) Validate() error {
	request := CreateEventRequest(*r)
	if err := request.Validate(); err != nil {
		return err
	}
	*r = UpdateEventRequest(request)
	return nil
}

type AnnouncementFilter struct {
	Status   string
	Category string
	Priority string
	Search   string
}

type EventFilter struct {
	Status   string
	Upcoming bool
	Search   string
}
