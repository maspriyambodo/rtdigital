package auth

import "time"

type User struct {
	ID                 string
	OrganizationID     string
	Email              *string
	Phone              *string
	PasswordHash       string
	Status             string
	FailedLoginCount   int
	LockedUntil        *time.Time
	MFASecretEncrypted *string
	MFAEnabledAt       *time.Time
}

type Session struct {
	ID             string
	UserID         string
	OrganizationID string
	RefreshHash    string
	ExpiresAt      time.Time
	RevokedAt      *time.Time
}

type LoginResult struct {
	AccessToken    string    `json:"access_token"`
	RefreshToken   string    `json:"-"`
	AccessExpires  time.Time `json:"expires_at"`
	RefreshExpires time.Time `json:"-"`
	MFARequired    bool      `json:"mfa_required"`
}
