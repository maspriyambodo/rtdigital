package auth

import (
	"errors"
	"testing"
	"time"
)

func TestTokenManagerRejectsExpiredAccessToken(t *testing.T) {
	t.Parallel()

	manager, err := NewTokenManager("a_very_long_secret_key_that_is_at_least_32_bytes_long")
	if err != nil {
		t.Fatalf("NewTokenManager(): %v", err)
	}
	manager.accessExpiry = -time.Minute

	token, err := manager.IssueAccessToken(TokenClaims{
		UserID:         "user-1",
		OrganizationID: "org-1",
		SessionID:      "session-1",
		MFA:            true,
	})
	if err != nil {
		t.Fatalf("IssueAccessToken(): %v", err)
	}

	if _, err := manager.VerifyAccessToken(token); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("VerifyAccessToken(expired) error = %v; want ErrInvalidToken", err)
	}
}
