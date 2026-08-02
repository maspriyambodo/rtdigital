package users

import "time"

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
