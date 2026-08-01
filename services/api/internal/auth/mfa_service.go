package auth

import (
	"context"
	"fmt"
)

type MFAEnrollment struct {
	URI    string `json:"uri"`
	Secret string `json:"secret"`
}

func (s *Service) GenerateMFASecret(ctx context.Context, userID string) (MFAEnrollment, error) {
	if s.crypter == nil {
		return MFAEnrollment{}, ErrMFANotConfigured
	}

	var email *string
	if err := s.db.QueryRow(ctx, `
		SELECT email FROM users WHERE id = $1 AND status = 'active'`,
		userID,
	).Scan(&email); err != nil {
		return MFAEnrollment{}, ErrInvalidCredentials
	}

	accountName := userID
	if email != nil && *email != "" {
		accountName = *email
	}

	secret, err := GenerateTOTPSecret()
	if err != nil {
		return MFAEnrollment{}, err
	}
	encrypted, err := s.crypter.Encrypt(secret)
	if err != nil {
		return MFAEnrollment{}, err
	}

	if _, err := s.db.Exec(ctx, `
		UPDATE users
		SET mfa_secret_encrypted = $1, mfa_enabled_at = NULL
		WHERE id = $2 AND status = 'active'`,
		encrypted, userID,
	); err != nil {
		return MFAEnrollment{}, fmt.Errorf("save MFA secret: %w", err)
	}

	return MFAEnrollment{
		URI:    GenerateTOTPURI(secret, accountName, "RT Digital"),
		Secret: secret,
	}, nil
}

func (s *Service) EnableMFA(ctx context.Context, userID, code string) error {
	if s.crypter == nil {
		return ErrMFANotConfigured
	}

	var encrypted string
	if err := s.db.QueryRow(ctx, `
		SELECT mfa_secret_encrypted
		FROM users
		WHERE id = $1 AND status = 'active' AND mfa_enabled_at IS NULL`,
		userID,
	).Scan(&encrypted); err != nil {
		return ErrInvalidTOTPCode
	}

	secret, err := s.crypter.Decrypt(encrypted)
	if err != nil {
		return ErrInvalidTOTPCode
	}
	ok, err := VerifyTOTP(code, secret, s.now())
	if err != nil || !ok {
		return ErrInvalidTOTPCode
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin MFA enable: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var organizationID string
	if err := tx.QueryRow(ctx, `
		UPDATE users SET mfa_enabled_at = now() WHERE id = $1
		RETURNING organization_id`,
		userID,
	).Scan(&organizationID); err != nil {
		return fmt.Errorf("enable MFA: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id)
		VALUES ($1, $2, 'auth.mfa.enable', 'users', $2)`,
		organizationID, userID,
	); err != nil {
		return fmt.Errorf("audit MFA enable: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit MFA enable: %w", err)
	}
	return nil
}

func (s *Service) DisableMFA(ctx context.Context, userID, code string) error {
	if s.crypter == nil {
		return ErrMFANotConfigured
	}

	var encrypted string
	if err := s.db.QueryRow(ctx, `
		SELECT mfa_secret_encrypted
		FROM users
		WHERE id = $1 AND status = 'active' AND mfa_enabled_at IS NOT NULL`,
		userID,
	).Scan(&encrypted); err != nil {
		return ErrMFANotConfigured
	}

	secret, err := s.crypter.Decrypt(encrypted)
	if err != nil {
		return ErrInvalidTOTPCode
	}
	ok, err := VerifyTOTP(code, secret, s.now())
	if err != nil || !ok {
		return ErrInvalidTOTPCode
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin MFA disable: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var organizationID string
	if err := tx.QueryRow(ctx, `
		UPDATE users
		SET mfa_enabled_at = NULL, mfa_secret_encrypted = NULL
		WHERE id = $1
		RETURNING organization_id`,
		userID,
	).Scan(&organizationID); err != nil {
		return fmt.Errorf("disable MFA: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id)
		VALUES ($1, $2, 'auth.mfa.disable', 'users', $2)`,
		organizationID, userID,
	); err != nil {
		return fmt.Errorf("audit MFA disable: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit MFA disable: %w", err)
	}
	return nil
}
