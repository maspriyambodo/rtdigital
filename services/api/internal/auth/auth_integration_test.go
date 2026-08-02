package auth_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
)

func TestAuthIntegration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := platform.NewDatabase(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	orgID := testUUID(t)
	userID := testUUID(t)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM organizations WHERE id = $1`, orgID)
	})

	if _, err := pool.Exec(ctx, `
		INSERT INTO organizations (id, name, rt_number, rw_number, status)
		VALUES ($1, 'Test Auth Organization', '001', '001', 'active')`,
		orgID,
	); err != nil {
		t.Fatalf("insert organization: %v", err)
	}

	password := "Password123!"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword(): %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, status)
		VALUES ($1, $2, $3, $4, 'active')`,
		userID, orgID, "auth-"+userID[:8]+"@example.test", hash,
	); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	tokens, err := auth.NewTokenManager("test-secret-must-be-longer-than-thirty-two-bytes")
	if err != nil {
		t.Fatalf("NewTokenManager(): %v", err)
	}
	crypter, err := auth.NewAESCrypter([]byte("01234567890123456789012345678901"))
	if err != nil {
		t.Fatalf("NewAESCrypter(): %v", err)
	}
	service := auth.NewService(pool, tokens, crypter, auth.NoopMailer{}, "http://localhost")

	t.Run("login, rotation, and logout", func(t *testing.T) {
		login, err := service.Login(ctx, "auth-"+userID[:8]+"@example.test", password, "test-agent", "127.0.0.1")
		if err != nil {
			t.Fatalf("Login(): %v", err)
		}
		if login.AccessToken == "" || login.RefreshToken == "" || login.MFARequired {
			t.Fatal("login did not return an active non-MFA session")
		}

		claims, err := tokens.VerifyAccessToken(login.AccessToken)
		if err != nil {
			t.Fatalf("VerifyAccessToken(): %v", err)
		}
		if claims.UserID != userID || claims.OrganizationID != orgID || !claims.MFA {
			t.Fatalf("unexpected access claims: %+v", claims)
		}

		rotated, err := service.Refresh(ctx, login.RefreshToken, "test-agent", "127.0.0.1")
		if err != nil {
			t.Fatalf("Refresh(): %v", err)
		}
		if rotated.RefreshToken == "" || rotated.RefreshToken == login.RefreshToken {
			t.Fatal("refresh token was not rotated")
		}
		if _, err := service.Refresh(ctx, login.RefreshToken, "test-agent", "127.0.0.1"); err != auth.ErrSessionExpired {
			t.Fatalf("reused refresh token error = %v; want ErrSessionExpired", err)
		}

		rotatedClaims, err := tokens.VerifyAccessToken(rotated.AccessToken)
		if err != nil {
			t.Fatalf("VerifyAccessToken(rotated): %v", err)
		}
		if err := service.Logout(ctx, rotatedClaims.SessionID); err != nil {
			t.Fatalf("Logout(): %v", err)
		}
		if _, err := service.Refresh(ctx, rotated.RefreshToken, "test-agent", "127.0.0.1"); err != auth.ErrSessionExpired {
			t.Fatalf("refresh after logout error = %v; want ErrSessionExpired", err)
		}
	})

	t.Run("lockout after failed attempts", func(t *testing.T) {
		if _, err := pool.Exec(ctx, `
			UPDATE users SET failed_login_count = 0, locked_until = NULL WHERE id = $1`,
			userID,
		); err != nil {
			t.Fatalf("reset lockout state: %v", err)
		}

		for range 5 {
			if _, err := service.Login(ctx, "auth-"+userID[:8]+"@example.test", "incorrect-password", "test-agent", "127.0.0.1"); err != auth.ErrInvalidCredentials {
				t.Fatalf("failed login error = %v; want ErrInvalidCredentials", err)
			}
		}
		if _, err := service.Login(ctx, "auth-"+userID[:8]+"@example.test", password, "test-agent", "127.0.0.1"); err != auth.ErrAccountLocked {
			t.Fatalf("locked login error = %v; want ErrAccountLocked", err)
		}
		if _, err := pool.Exec(ctx, `
			UPDATE users SET failed_login_count = 0, locked_until = NULL WHERE id = $1`,
			userID,
		); err != nil {
			t.Fatalf("clear lockout state: %v", err)
		}
	})

	t.Run("MFA verification", func(t *testing.T) {
		enrollment, err := service.GenerateMFASecret(ctx, userID)
		if err != nil {
			t.Fatalf("GenerateMFASecret(): %v", err)
		}
		if enrollment.Secret == "" || enrollment.URI == "" {
			t.Fatal("MFA enrollment is incomplete")
		}

		code, err := auth.GenerateTOTPCode(enrollment.Secret, time.Now().UTC())
		if err != nil {
			t.Fatalf("GenerateTOTPCode(): %v", err)
		}
		if err := service.EnableMFA(ctx, userID, code); err != nil {
			t.Fatalf("EnableMFA(): %v", err)
		}

		login, err := service.Login(ctx, "auth-"+userID[:8]+"@example.test", password, "test-agent", "127.0.0.1")
		if err != nil {
			t.Fatalf("MFA Login(): %v", err)
		}
		if !login.MFARequired {
			t.Fatal("MFA login must require verification")
		}

		pendingClaims, err := tokens.VerifyAccessToken(login.AccessToken)
		if err != nil {
			t.Fatalf("VerifyAccessToken(pending): %v", err)
		}
		if pendingClaims.MFA {
			t.Fatal("pending MFA token unexpectedly has MFA claim")
		}

		code, err = auth.GenerateTOTPCode(enrollment.Secret, time.Now().UTC())
		if err != nil {
			t.Fatalf("GenerateTOTPCode(): %v", err)
		}
		verified, err := service.VerifyMFA(ctx, pendingClaims, code)
		if err != nil {
			t.Fatalf("VerifyMFA(): %v", err)
		}
		verifiedClaims, err := tokens.VerifyAccessToken(verified.AccessToken)
		if err != nil || !verifiedClaims.MFA {
			t.Fatalf("verified token is invalid or lacks MFA claim: %v", err)
		}
	})

	t.Run("inactive account is rejected", func(t *testing.T) {
		inactiveID := testUUID(t)
		if _, err := pool.Exec(ctx, `
			INSERT INTO users (id, organization_id, email, password_hash, status)
			VALUES ($1, $2, $3, $4, 'inactive')`,
			inactiveID, orgID, "inactive-"+inactiveID[:8]+"@example.test", hash,
		); err != nil {
			t.Fatalf("insert inactive user: %v", err)
		}
		if _, err := service.Login(ctx, "inactive-"+inactiveID[:8]+"@example.test", password, "test-agent", "127.0.0.1"); err != auth.ErrAccountInactive {
			t.Fatalf("inactive login error = %v; want ErrAccountInactive", err)
		}
	})
}

func testUUID(t *testing.T) string {
	t.Helper()

	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		t.Fatalf("generate UUID bytes: %v", err)
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(bytes)
	return encoded[:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:]
}
