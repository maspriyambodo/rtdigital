package notifications

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrValidation           = errors.New("validation failed")
	ErrNotificationNotFound = errors.New("notification not found")
)

type Notification struct {
	ID            string     `json:"id"`
	Type          string     `json:"type"`
	Title         string     `json:"title"`
	Body          *string    `json:"body,omitempty"`
	ReferenceType *string    `json:"reference_type,omitempty"`
	ReferenceID   *string    `json:"reference_id,omitempty"`
	ReadAt        *time.Time `json:"read_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type Filter struct {
	UnreadOnly bool
	Limit      int
}

type CreateParams struct {
	UserID        string
	Type          string
	Title         string
	Body          *string
	ReferenceType *string
	ReferenceID   *string
}

func (p CreateParams) Validate() error {
	if strings.TrimSpace(p.UserID) == "" ||
		strings.TrimSpace(p.Type) == "" ||
		strings.TrimSpace(p.Title) == "" {
		return ErrValidation
	}
	return nil
}