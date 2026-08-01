package auth_test

import (
	"testing"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

func TestTokenManager(t *testing.T) {
	t.Parallel()

	tokenManager, err := auth.NewTokenManager("a_very_long_secret_key_that_is_at_least_32_bytes_long")
	if err != nil {
		t.Fatalf("NewTokenManager(): %v", err)
	}

	want := auth.TokenClaims{
		UserID:         "user-1",
		OrganizationID: "org-1",
		SessionID:      "session-1",
		MFA:            true,
	}
	token, err := tokenManager.IssueAccessToken(want)
	if err != nil {
		t.Fatalf("IssueAccessToken(): %v", err)
	}

	got, err := tokenManager.VerifyAccessToken(token)
	if err != nil {
		t.Fatalf("VerifyAccessToken(): %v", err)
	}
	if got.UserID != want.UserID || got.OrganizationID != want.OrganizationID || got.SessionID != want.SessionID || got.MFA != want.MFA {
		t.Errorf("VerifyAccessToken() = %+v; want identity %+v", got, want)
	}

	if _, err := tokenManager.VerifyAccessToken(token + "invalid"); err != auth.ErrInvalidToken {
		t.Errorf("VerifyAccessToken(tampered): %v; want ErrInvalidToken", err)
	}

	otherManager, err := auth.NewTokenManager("another_long_secret_key_that_is_at_least_32_bytes_long")
	if err != nil {
		t.Fatalf("NewTokenManager(other): %v", err)
	}
	if _, err := otherManager.VerifyAccessToken(token); err != auth.ErrInvalidToken {
		t.Errorf("VerifyAccessToken(wrong secret): %v; want ErrInvalidToken", err)
	}

	if _, err := auth.NewTokenManager("too-short"); err == nil {
		t.Error("NewTokenManager(short): expected error")
	}
}

func TestOpaqueToken(t *testing.T) {
	t.Parallel()

	raw, hash, err := auth.GenerateOpaqueToken()
	if err != nil {
		t.Fatalf("GenerateOpaqueToken(): %v", err)
	}
	if len(raw) < 43 {
		t.Errorf("raw token length = %d; want >= 43", len(raw))
	}
	if got := auth.HashOpaqueToken(raw); got != hash {
		t.Errorf("HashOpaqueToken() = %q; want %q", got, hash)
	}

	raw2, hash2, err := auth.GenerateOpaqueToken()
	if err != nil {
		t.Fatalf("GenerateOpaqueToken() second call: %v", err)
	}
	if raw == raw2 || hash == hash2 {
		t.Error("opaque tokens must be random")
	}
}