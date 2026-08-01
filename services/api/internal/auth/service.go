package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountInactive    = errors.New("account is inactive")
	ErrAccountLocked      = errors.New("account is temporarily locked")
	ErrAccountInvited     = errors.New("account activation required")
	ErrSessionExpired     = errors.New("session expired or revoked")
	ErrTokenNotFound      = errors.New("token not found or expired")
	ErrMFANotConfigured   = errors.New("MFA not configured")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
)

const (
	maxFailedAttempts = 5
	lockoutDuration   = 15 * time.Minute
)

type Service struct {
	db         *pgxpool.Pool
	tokens     *TokenManager
	crypter    Crypter
	mailer     Mailer
	appBaseURL string
	now        func() time.Time
}

func NewService(db *pgxpool.Pool, tokens *TokenManager, crypter Crypter, mailer Mailer, appBaseURL string) *Service {
	if mailer == nil {
		mailer = NoopMailer{}
	}
	return &Service{
		db:         db,
		tokens:     tokens,
		crypter:    crypter,
		mailer:     mailer,
		appBaseURL: strings.TrimRight(appBaseURL, "/"),
		now:        func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Login(ctx context.Context, identifier, password, userAgent, ipAddress string) (LoginResult, error) {
	email, _ := NormalizeEmail(identifier)
	phone, _ := NormalizePhone(identifier)
	if email == "" && phone == "" {
		return LoginResult{}, ErrInvalidCredentials
	}

	var user User
	err := s.db.QueryRow(ctx, `
		SELECT id, organization_id, email, phone, password_hash, status,
		       failed_login_count, locked_until, mfa_secret_encrypted, mfa_enabled_at
		FROM users
		WHERE ($1 <> '' AND lower(email) = $1) OR ($2 <> '' AND phone = $2)
		ORDER BY created_at
		LIMIT 1`, email, phone,
	).Scan(
		&user.ID, &user.OrganizationID, &user.Email, &user.Phone, &user.PasswordHash,
		&user.Status, &user.FailedLoginCount, &user.LockedUntil,
		&user.MFASecretEncrypted, &user.MFAEnabledAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return LoginResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, fmt.Errorf("find login user: %w", err)
	}

	now := s.now()
	if user.LockedUntil != nil && user.LockedUntil.After(now) {
		return LoginResult{}, ErrAccountLocked
	}
	switch user.Status {
	case "inactive":
		return LoginResult{}, ErrAccountInactive
	case "invited":
		return LoginResult{}, ErrAccountInvited
	case "locked":
		return LoginResult{}, ErrAccountLocked
	}

	ok, err := VerifyPassword(password, user.PasswordHash)
	if err != nil || !ok {
		_ = s.recordFailedLogin(ctx, user.ID)
		return LoginResult{}, ErrInvalidCredentials
	}

	sessionID := newUUID()
	rawRefresh, refreshHash, err := GenerateOpaqueToken()
	if err != nil {
		return LoginResult{}, err
	}
	refreshExpires := now.Add(s.tokens.RefreshExpiry())

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return LoginResult{}, fmt.Errorf("begin login transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		UPDATE users
		SET failed_login_count = 0, locked_until = NULL, last_login_at = now()
		WHERE id = $1`, user.ID); err != nil {
		return LoginResult{}, fmt.Errorf("record login: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO sessions (id, user_id, organization_id, refresh_token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		sessionID, user.ID, user.OrganizationID, refreshHash, userAgent, validIP(ipAddress), refreshExpires,
	); err != nil {
		return LoginResult{}, fmt.Errorf("create session: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (
			organization_id, actor_user_id, action, entity_type, entity_id,
			metadata, ip_address, user_agent
		)
		VALUES ($1, $2, 'auth.login.success', 'users', $2, '{"channel":"password"}'::jsonb, $3, $4)`,
		user.OrganizationID, user.ID, validIP(ipAddress), userAgent,
	); err != nil {
		return LoginResult{}, fmt.Errorf("write login audit: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return LoginResult{}, fmt.Errorf("commit login: %w", err)
	}

	mfaRequired := user.MFAEnabledAt != nil && user.MFASecretEncrypted != nil
	accessToken, err := s.tokens.IssueAccessToken(TokenClaims{
		UserID: user.ID, OrganizationID: user.OrganizationID, SessionID: sessionID, MFA: !mfaRequired,
	})
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		AccessToken: accessToken, RefreshToken: rawRefresh,
		AccessExpires: now.Add(s.tokens.AccessExpiry()), RefreshExpires: refreshExpires,
		MFARequired: mfaRequired,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, rawToken, userAgent, ipAddress string) (LoginResult, error) {
	now := s.now()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return LoginResult{}, fmt.Errorf("begin refresh transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var session Session
	var status string
	var mfaEnabledAt *time.Time
	err = tx.QueryRow(ctx, `
		SELECT s.id, s.user_id, s.organization_id, s.refresh_token_hash, s.expires_at, s.revoked_at,
		       u.status, u.mfa_enabled_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.refresh_token_hash = $1
		FOR UPDATE`, HashOpaqueToken(rawToken),
	).Scan(
		&session.ID, &session.UserID, &session.OrganizationID, &session.RefreshHash, &session.ExpiresAt, &session.RevokedAt,
		&status, &mfaEnabledAt,
	)
	if errors.Is(err, pgx.ErrNoRows) || session.RevokedAt != nil || !session.ExpiresAt.After(now) || status != "active" {
		return LoginResult{}, ErrSessionExpired
	}
	if err != nil {
		return LoginResult{}, fmt.Errorf("find session: %w", err)
	}

	if _, err := tx.Exec(ctx, `UPDATE sessions SET revoked_at = $1, revoke_reason = 'rotated' WHERE id = $2`, now, session.ID); err != nil {
		return LoginResult{}, fmt.Errorf("revoke prior session: %w", err)
	}

	newSessionID := newUUID()
	rawRefresh, refreshHash, err := GenerateOpaqueToken()
	if err != nil {
		return LoginResult{}, err
	}
	refreshExpires := now.Add(s.tokens.RefreshExpiry())
	if _, err := tx.Exec(ctx, `
		INSERT INTO sessions (id, user_id, organization_id, refresh_token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		newSessionID, session.UserID, session.OrganizationID, refreshHash, userAgent, validIP(ipAddress), refreshExpires,
	); err != nil {
		return LoginResult{}, fmt.Errorf("create rotated session: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return LoginResult{}, fmt.Errorf("commit token rotation: %w", err)
	}

	mfaRequired := mfaEnabledAt != nil
	accessToken, err := s.tokens.IssueAccessToken(TokenClaims{
		UserID: session.UserID, OrganizationID: session.OrganizationID, SessionID: newSessionID, MFA: !mfaRequired,
	})
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{
		AccessToken: accessToken, RefreshToken: rawRefresh,
		AccessExpires: now.Add(s.tokens.AccessExpiry()), RefreshExpires: refreshExpires,
		MFARequired: mfaRequired,
	}, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE sessions SET revoked_at = now(), revoke_reason = 'user_logout'
		WHERE id = $1 AND revoked_at IS NULL`, sessionID)
	return err
}

func (s *Service) LogoutAll(ctx context.Context, userID string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE sessions SET revoked_at = now(), revoke_reason = 'logout_all'
		WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}

func (s *Service) ActivateAccount(ctx context.Context, rawToken, password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}
	return s.consumePasswordToken(ctx, "activation_tokens", rawToken, password, false)
}

func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	email, err := NormalizeEmail(email)
	if err != nil {
		return nil // Prevent account enumeration.
	}

	var userID string
	err = s.db.QueryRow(ctx, `SELECT id FROM users WHERE lower(email) = $1 AND status = 'active'`, email).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("find reset user: %w", err)
	}

	raw, tokenHash, err := GenerateOpaqueToken()
	if err != nil {
		return err
	}
	if _, err := s.db.Exec(ctx, `
		INSERT INTO password_reset_tokens (token_hash, user_id, expires_at)
		VALUES ($1, $2, now() + interval '1 hour')`, tokenHash, userID); err != nil {
		return fmt.Errorf("create reset token: %w", err)
	}

	link := fmt.Sprintf("%s/reset-password?token=%s", s.appBaseURL, raw)
	_ = s.mailer.SendEmail(ctx, email, "Reset Kata Sandi RT Digital",
		fmt.Sprintf(`<p>Atur ulang kata sandi melalui <a href="%s">tautan ini</a>.</p>`, link))
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, rawToken, password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}
	return s.consumePasswordToken(ctx, "password_reset_tokens", rawToken, password, true)
}

func (s *Service) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrWeakPassword
	}
	var oldHash string
	if err := s.db.QueryRow(ctx, `SELECT password_hash FROM users WHERE id = $1 AND status = 'active'`, userID).Scan(&oldHash); err != nil {
		return ErrInvalidCredentials
	}
	ok, err := VerifyPassword(oldPassword, oldHash)
	if err != nil || !ok {
		return ErrInvalidCredentials
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `
		UPDATE users SET password_hash = $1, failed_login_count = 0, locked_until = NULL WHERE id = $2`, hash, userID)
	return err
}

func (s *Service) VerifyMFA(ctx context.Context, claims *TokenClaims, code string) (LoginResult, error) {
	if claims == nil || claims.MFA {
		return LoginResult{}, ErrInvalidToken
	}

	var secret string
	err := s.db.QueryRow(ctx, `
		SELECT mfa_secret_encrypted FROM users
		WHERE id = $1 AND organization_id = $2 AND mfa_enabled_at IS NOT NULL`,
		claims.UserID, claims.OrganizationID,
	).Scan(&secret)
	if err != nil || s.crypter == nil {
		return LoginResult{}, ErrMFANotConfigured
	}
	decrypted, err := s.crypter.Decrypt(secret)
	if err != nil {
		return LoginResult{}, ErrInvalidTOTPCode // Do not expose encryption failures.
	}
	ok, err := VerifyTOTP(code, decrypted, s.now())
	if err != nil || !ok {
		return LoginResult{}, ErrInvalidTOTPCode
	}
	accessToken, err := s.tokens.IssueAccessToken(TokenClaims{
		UserID: claims.UserID, OrganizationID: claims.OrganizationID, SessionID: claims.SessionID, MFA: true,
	})
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{AccessToken: accessToken, AccessExpires: s.now().Add(s.tokens.AccessExpiry())}, nil
}

func (s *Service) recordFailedLogin(ctx context.Context, userID string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE users
		SET failed_login_count = failed_login_count + 1,
		    locked_until = CASE
		        WHEN failed_login_count + 1 >= $1 THEN now() + $2::interval
		        ELSE locked_until
		    END
		WHERE id = $3`, maxFailedAttempts, lockoutDuration.String(), userID)
	return err
}

func (s *Service) consumePasswordToken(ctx context.Context, table, rawToken, password string, revokeSessions bool) error {
	tokenHash := HashOpaqueToken(rawToken)
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin token transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID string
	err = tx.QueryRow(ctx, fmt.Sprintf(`
		UPDATE %s
		SET used_at = now()
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now()
		RETURNING user_id`, table), tokenHash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrTokenNotFound
	}
	if err != nil {
		return fmt.Errorf("consume token: %w", err)
	}

	status := "active"
	if table == "password_reset_tokens" {
		if _, err := tx.Exec(ctx, `
			UPDATE users
			SET password_hash = $1, failed_login_count = 0, locked_until = NULL
			WHERE id = $2`, hash, userID); err != nil {
			return fmt.Errorf("reset password: %w", err)
		}
	} else if _, err := tx.Exec(ctx, `
		UPDATE users
		SET password_hash = $1, status = $2, failed_login_count = 0, locked_until = NULL
		WHERE id = $3`, hash, status, userID); err != nil {
		return fmt.Errorf("activate account: %w", err)
	}

	if revokeSessions {
		if _, err := tx.Exec(ctx, `
			UPDATE sessions SET revoked_at = now(), revoke_reason = 'password_reset'
			WHERE user_id = $1 AND revoked_at IS NULL`, userID); err != nil {
			return fmt.Errorf("revoke sessions: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func validIP(value string) any {
	address, err := netip.ParseAddr(value)
	if err != nil {
		return nil
	}
	return address.String()
}

func newUUID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic("secure random source unavailable")
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}