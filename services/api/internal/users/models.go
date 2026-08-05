package users

import (
	"encoding/json"
	"time"
)

type UserListItem struct {
	ID          string     `json:"id"`
	Email       *string    `json:"email"`
	Phone       *string    `json:"phone"`
	Status      string     `json:"status"`
	Roles       []string   `json:"roles"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type UserDetail struct {
	ID           string     `json:"id"`
	Email        *string    `json:"email"`
	Phone        *string    `json:"phone"`
	Status       string     `json:"status"`
	Roles        []RoleInfo `json:"roles"`
	MFAEnabledAt *time.Time `json:"mfa_enabled_at"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type RoleInfo struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type InviteUserRequest struct {
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	RoleCodes []string `json:"role_codes"`
}

type InviteUserResponse struct {
	UserID        string    `json:"user_id"`
	InviteToken   string    `json:"invite_token,omitempty"`
	ActivationURL string    `json:"activation_url"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type OfficeHandover struct {
	ID             string          `json:"id"`
	OutgoingUserID string          `json:"outgoing_user_id"`
	IncomingUserID *string         `json:"incoming_user_id,omitempty"`
	Status         string          `json:"status"`
	Checklist      json.RawMessage `json:"checklist"`
	Notes          *string         `json:"notes,omitempty"`
	CompletedBy    *string         `json:"completed_by,omitempty"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type CreateOfficeHandoverRequest struct {
	OutgoingUserID string          `json:"outgoing_user_id"`
	Checklist      json.RawMessage `json:"checklist"`
	Notes          *string         `json:"notes,omitempty"`
}

type CompleteOfficeHandoverRequest struct {
	IncomingUserID string          `json:"incoming_user_id"`
	Checklist      json.RawMessage `json:"checklist"`
	Notes          *string         `json:"notes,omitempty"`
}
