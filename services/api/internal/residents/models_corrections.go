package residents

import "time"

type ResidentCorrection struct {
	ID               string         `json:"id"`
	ResidentID       string         `json:"resident_id"`
	ResidentName     string         `json:"resident_name"`
	RequesterUserID  string         `json:"requester_user_id"`
	RequesterName    *string        `json:"requester_name,omitempty"`
	RequestedChanges map[string]any `json:"requested_changes"`
	Reason           string         `json:"reason"`
	Status           string         `json:"status"`
	ReviewerUserID   *string        `json:"reviewer_user_id,omitempty"`
	ReviewerName     *string        `json:"reviewer_name,omitempty"`
	ReviewerNote     *string        `json:"reviewer_note,omitempty"`
	ReviewedAt       *time.Time     `json:"reviewed_at,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type CreateResidentCorrectionRequest struct {
	RequestedChanges map[string]any `json:"requested_changes"`
	Reason           string         `json:"reason"`
}

type ReviewResidentCorrectionRequest struct {
	Note string `json:"note"`
}
