package auth_test

import (
	"strings"
	"testing"
	"time"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

func TestTOTP(t *testing.T) {
	t.Parallel()

	secret, err := auth.GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret(): %v", err)
	}
	if len(secret) != 32 {
		t.Errorf("secret length = %d; want 32", len(secret))
	}

	uri := auth.GenerateTOTPURI(secret, "user@example.com", "RT Digital")
	if !strings.HasPrefix(uri, "otpauth://totp/RT%20Digital:user@example.com?secret=") {
		t.Errorf("unexpected URI format: %q", uri)
	}

	at := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	code, err := auth.GenerateTOTPCode(secret, at)
	if err != nil {
		t.Fatalf("GenerateTOTPCode(): %v", err)
	}
	if len(code) != 6 {
		t.Errorf("code length = %d; want 6", len(code))
	}

	for _, verificationTime := range []time.Time{at, at.Add(29 * time.Second), at.Add(31 * time.Second)} {
		ok, err := auth.VerifyTOTP(code, secret, verificationTime)
		if err != nil || !ok {
			t.Errorf("VerifyTOTP(%s) = %v, %v; want true, nil", verificationTime, ok, err)
		}
	}

	ok, err := auth.VerifyTOTP(code, secret, at.Add(90*time.Second))
	if err != nil || ok {
		t.Errorf("VerifyTOTP(outside window) = %v, %v; want false, nil", ok, err)
	}

	for _, invalid := range []string{"", "12345", "1234567", "abcdef", "      "} {
		ok, err := auth.VerifyTOTP(invalid, secret, at)
		if err != auth.ErrInvalidTOTPCode || ok {
			t.Errorf("VerifyTOTP(%q) = %v, %v; want false, ErrInvalidTOTPCode", invalid, ok, err)
		}
	}

	if _, err := auth.VerifyTOTP("123456", "invalid-secret", at); err == nil {
		t.Error("VerifyTOTP(invalid secret): expected error")
	}
}

func TestTOTPMatchesRFC6238Vector(t *testing.T) {
	t.Parallel()

	code, err := auth.GenerateTOTPCode("GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", time.Unix(59, 0))
	if err != nil {
		t.Fatalf("GenerateTOTPCode(): %v", err)
	}
	if code != "287082" {
		t.Errorf("GenerateTOTPCode() = %q; want RFC 6238 vector %q", code, "287082")
	}
}