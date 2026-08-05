package users

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

var (
	ErrValidation            = errors.New("validation failed")
	ErrUserNotFound          = errors.New("user not found")
	ErrRoleNotFound          = errors.New("role not found")
	ErrDuplicateContact      = errors.New("email or phone already in use")
	ErrCannotEscalate        = errors.New("cannot escalate privileges")
	ErrCannotModifySelf      = errors.New("cannot modify own roles")
	ErrMFAEnrollmentRequired = errors.New("MFA enrollment required")
	ErrForbidden             = errors.New("forbidden")
)

type NotificationDispatcher interface {
	DispatchNotification(organizationID, recipientUserID, notificationType, title, body, referenceType, referenceID string)
}

type Service struct {
	db         *pgxpool.Pool
	mailer     auth.Mailer
	dispatcher NotificationDispatcher
	appBaseURL string
	now        func() time.Time
}

func NewService(db *pgxpool.Pool, mailer auth.Mailer, appBaseURL string) *Service {
	if mailer == nil {
		mailer = auth.NoopMailer{}
	}
	return &Service{
		db:         db,
		mailer:     mailer,
		appBaseURL: strings.TrimRight(appBaseURL, "/"),
		now:        func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) SetNotificationDispatcher(dispatcher NotificationDispatcher) {
	s.dispatcher = dispatcher
}

func (s *Service) ListRoles(ctx context.Context, principal *auth.Principal) ([]RoleInfo, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, code, name, description
		FROM roles
		WHERE organization_id IS NULL OR organization_id = $1
		ORDER BY name`, principal.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("query roles: %w", err)
	}
	defer rows.Close()

	roles := []RoleInfo{}
	for rows.Next() {
		var role RoleInfo
		if err := rows.Scan(&role.ID, &role.Code, &role.Name, &role.Description); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (s *Service) ListUsers(ctx context.Context, principal *auth.Principal) ([]UserListItem, error) {
	rows, err := s.db.Query(ctx, `
		SELECT u.id, u.email, u.phone, u.status, u.last_login_at, u.created_at,
		       COALESCE(array_agg(r.code ORDER BY r.code) FILTER (WHERE r.code IS NOT NULL), ARRAY[]::varchar[])
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN roles r ON r.id = ur.role_id
		WHERE u.organization_id = $1
		GROUP BY u.id
		ORDER BY u.created_at DESC`, principal.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	users := []UserListItem{}
	for rows.Next() {
		var user UserListItem
		if err := rows.Scan(&user.ID, &user.Email, &user.Phone, &user.Status, &user.LastLoginAt, &user.CreatedAt, &user.Roles); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *Service) GetUser(ctx context.Context, principal *auth.Principal, userID string) (UserDetail, error) {
	var user UserDetail
	err := s.db.QueryRow(ctx, `
		SELECT id, email, phone, status, mfa_enabled_at, last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1 AND organization_id = $2`, userID, principal.OrganizationID,
	).Scan(&user.ID, &user.Email, &user.Phone, &user.Status, &user.MFAEnabledAt, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserDetail{}, ErrUserNotFound
	}
	if err != nil {
		return UserDetail{}, fmt.Errorf("query user: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT r.id, r.code, r.name, r.description
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name`, userID)
	if err != nil {
		return UserDetail{}, fmt.Errorf("query user roles: %w", err)
	}
	defer rows.Close()

	user.Roles = []RoleInfo{}
	for rows.Next() {
		var role RoleInfo
		if err := rows.Scan(&role.ID, &role.Code, &role.Name, &role.Description); err != nil {
			return UserDetail{}, fmt.Errorf("scan user role: %w", err)
		}
		user.Roles = append(user.Roles, role)
	}
	return user, rows.Err()
}

func (s *Service) InviteUser(ctx context.Context, principal *auth.Principal, request InviteUserRequest) (InviteUserResponse, error) {
	email, _ := auth.NormalizeEmail(request.Email)
	phone, _ := auth.NormalizePhone(request.Phone)
	if email == "" && phone == "" {
		return InviteUserResponse{}, fmt.Errorf("%w: email or phone required", ErrValidation)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return InviteUserResponse{}, fmt.Errorf("begin invite: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	roleIDs, err := validateRoleCodes(ctx, tx, principal, request.RoleCodes)
	if err != nil {
		return InviteUserResponse{}, err
	}

	userID := newUUID()
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, phone, password_hash, status)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), '!', 'invited')`,
		userID, principal.OrganizationID, email, phone)
	if err != nil {
		if strings.Contains(err.Error(), "uq_users_organization_email") || strings.Contains(err.Error(), "uq_users_organization_phone") {
			return InviteUserResponse{}, ErrDuplicateContact
		}
		return InviteUserResponse{}, fmt.Errorf("create invited user: %w", err)
	}

	for _, roleID := range roleIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id, assigned_by)
			VALUES ($1, $2, $3)`, userID, roleID, principal.UserID); err != nil {
			return InviteUserResponse{}, fmt.Errorf("assign invited role: %w", err)
		}
	}

	rawToken, tokenHash, err := auth.GenerateOpaqueToken()
	if err != nil {
		return InviteUserResponse{}, err
	}
	expiresAt := s.now().Add(7 * 24 * time.Hour)
	if _, err := tx.Exec(ctx, `
		INSERT INTO activation_tokens (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)`, tokenHash, userID, expiresAt); err != nil {
		return InviteUserResponse{}, fmt.Errorf("create activation token: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id, metadata)
		VALUES ($1, $2, 'user.invite', 'users', $3, jsonb_build_object('role_codes', $4::text[]))`,
		principal.OrganizationID, principal.UserID, userID, request.RoleCodes); err != nil {
		return InviteUserResponse{}, fmt.Errorf("audit invite: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return InviteUserResponse{}, fmt.Errorf("commit invite: %w", err)
	}

	activationURL := fmt.Sprintf("%s/activate?token=%s", s.appBaseURL, rawToken)
	if s.dispatcher != nil {
		s.dispatcher.DispatchNotification(
			principal.OrganizationID, userID, "account_invitation", "Undangan Akun RT Digital",
			fmt.Sprintf("Aktifkan akun RT Digital melalui tautan ini:\n%s", activationURL),
			"user", userID,
		)
	} else if email != "" {
		_ = s.mailer.SendEmail(ctx, email, "Undangan Akun RT Digital",
			fmt.Sprintf(`<p>Aktifkan akun RT Digital melalui <a href="%s">tautan ini</a>.</p>`, activationURL))
	}
	return InviteUserResponse{UserID: userID, InviteToken: rawToken, ActivationURL: activationURL, ExpiresAt: expiresAt}, nil
}

func (s *Service) AssignRole(ctx context.Context, principal *auth.Principal, userID, roleID string) error {
	return s.changeRole(ctx, principal, userID, roleID, true)
}

func (s *Service) RevokeRole(ctx context.Context, principal *auth.Principal, userID, roleID string) error {
	return s.changeRole(ctx, principal, userID, roleID, false)
}

func (s *Service) changeRole(ctx context.Context, principal *auth.Principal, userID, roleID string, assign bool) error {
	if principal.UserID == userID {
		return ErrCannotModifySelf
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin role change: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var roleCode string
	if err := tx.QueryRow(ctx, `
		SELECT code FROM roles
		WHERE id = $1 AND (organization_id IS NULL OR organization_id = $2)`,
		roleID, principal.OrganizationID).Scan(&roleCode); errors.Is(err, pgx.ErrNoRows) {
		return ErrRoleNotFound
	} else if err != nil {
		return fmt.Errorf("query role: %w", err)
	}
	if roleCode == "super_admin" && !hasRole(principal, "super_admin") {
		return ErrCannotEscalate
	}

	var exists bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND organization_id = $2)`,
		userID, principal.OrganizationID).Scan(&exists); err != nil || !exists {
		return ErrUserNotFound
	}

	action := "role.revoke"
	if assign {
		switch roleCode {
		case "super_admin", "ketua_rt", "sekretaris", "bendahara":
			var mfaEnabled bool
			if err := tx.QueryRow(ctx, `
				SELECT mfa_enabled_at IS NOT NULL
				FROM users
				WHERE id = $1 AND organization_id = $2`,
				userID, principal.OrganizationID,
			).Scan(&mfaEnabled); err != nil {
				return fmt.Errorf("check target MFA: %w", err)
			}
			if !mfaEnabled {
				return ErrMFAEnrollmentRequired
			}
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id, assigned_by)
			VALUES ($1, $2, $3)
			ON CONFLICT DO NOTHING`, userID, roleID, principal.UserID); err != nil {
			return fmt.Errorf("assign role: %w", err)
		}
		action = "role.assign"
	} else if _, err := tx.Exec(ctx, `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`, userID, roleID); err != nil {
		return fmt.Errorf("revoke role: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id, metadata)
		VALUES ($1, $2, $3, 'users', $4, jsonb_build_object('role_code', $5::text))`,
		principal.OrganizationID, principal.UserID, action, userID, roleCode); err != nil {
		return fmt.Errorf("audit role change: %w", err)
	}
	return tx.Commit(ctx)
}

func (s *Service) UpdateUser(ctx context.Context, principal *auth.Principal, userID string, emailInput, phoneInput *string, statusInput *string) error {
	if principal.UserID == userID && statusInput != nil {
		return ErrCannotModifySelf
	}
	if emailInput == nil && phoneInput == nil && statusInput == nil {
		return fmt.Errorf("%w: no changes", ErrValidation)
	}

	var email, phone *string
	if emailInput != nil {
		normalized, err := auth.NormalizeEmail(*emailInput)
		if err != nil {
			return fmt.Errorf("%w: email", ErrValidation)
		}
		email = &normalized
	}
	if phoneInput != nil {
		normalized, err := auth.NormalizePhone(*phoneInput)
		if err != nil {
			return fmt.Errorf("%w: phone", ErrValidation)
		}
		phone = &normalized
	}
	if statusInput != nil {
		switch *statusInput {
		case "active", "inactive", "locked":
		default:
			return fmt.Errorf("%w: status", ErrValidation)
		}
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin update user: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `
		UPDATE users
		SET email = COALESCE($1, email),
		    phone = COALESCE($2, phone),
		    status = COALESCE($3, status)
		WHERE id = $4 AND organization_id = $5`,
		email, phone, statusInput, userID, principal.OrganizationID)
	if err != nil {
		if strings.Contains(err.Error(), "uq_users_organization_email") || strings.Contains(err.Error(), "uq_users_organization_phone") {
			return ErrDuplicateContact
		}
		return fmt.Errorf("update user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	if statusInput != nil && *statusInput != "active" {
		if _, err := tx.Exec(ctx, `
			UPDATE sessions SET revoked_at = now(), revoke_reason = 'status_changed'
			WHERE user_id = $1 AND revoked_at IS NULL`, userID); err != nil {
			return fmt.Errorf("revoke user sessions: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id, metadata)
		VALUES ($1, $2, 'user.update', 'users', $3, jsonb_strip_nulls(jsonb_build_object('status', $4::text)))`,
		principal.OrganizationID, principal.UserID, userID, statusInput); err != nil {
		return fmt.Errorf("audit update: %w", err)
	}
	return tx.Commit(ctx)
}

func (s *Service) DeactivateUser(ctx context.Context, principal *auth.Principal, userID string) error {
	if principal.UserID == userID {
		return fmt.Errorf("%w: cannot deactivate self", ErrValidation)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin deactivate: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `
		UPDATE users SET status = 'inactive'
		WHERE id = $1 AND organization_id = $2 AND status <> 'inactive'`, userID, principal.OrganizationID)
	if err != nil {
		return fmt.Errorf("deactivate user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	if _, err := tx.Exec(ctx, `
		UPDATE sessions SET revoked_at = now(), revoke_reason = 'account_deactivated'
		WHERE user_id = $1 AND revoked_at IS NULL`, userID); err != nil {
		return fmt.Errorf("revoke user sessions: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id)
		VALUES ($1, $2, 'user.deactivate', 'users', $3)`,
		principal.OrganizationID, principal.UserID, userID); err != nil {
		return fmt.Errorf("audit deactivate: %w", err)
	}
	return tx.Commit(ctx)
}

func validateRoleCodes(ctx context.Context, tx pgx.Tx, principal *auth.Principal, codes []string) ([]string, error) {
	if len(codes) == 0 {
		return nil, fmt.Errorf("%w: role_codes", ErrValidation)
	}

	seen := make(map[string]struct{}, len(codes))
	roleIDs := make([]string, 0, len(codes))
	for _, code := range codes {
		if _, duplicate := seen[code]; duplicate || code == "" {
			return nil, fmt.Errorf("%w: role_codes", ErrValidation)
		}
		seen[code] = struct{}{}

		// Sekretaris hanya boleh mengundang warga; Ketua RT mengelola
		// penugasan melalui endpoint role terpisah.
		if !hasRole(principal, "super_admin") && (code == "super_admin" || (hasRole(principal, "sekretaris") && code != "warga")) {
			return nil, ErrCannotEscalate
		}
		var roleID string
		err := tx.QueryRow(ctx, `
			SELECT id FROM roles
			WHERE code = $1 AND (organization_id IS NULL OR organization_id = $2)`,
			code, principal.OrganizationID).Scan(&roleID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		if err != nil {
			return nil, fmt.Errorf("query requested role: %w", err)
		}
		roleIDs = append(roleIDs, roleID)
	}
	return roleIDs, nil
}

func (s *Service) audit(ctx context.Context, tx pgx.Tx, principal *auth.Principal, action, entityType, entityID string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id)
		VALUES ($1, $2, $3, $4, $5)`,
		principal.OrganizationID, principal.UserID, action, entityType, entityID,
	)
	if err != nil {
		return fmt.Errorf("audit %s: %w", action, err)
	}
	return nil
}

func hasRole(principal *auth.Principal, roleCode string) bool {
	for _, code := range principal.RoleCodes {
		if code == roleCode {
			return true
		}
	}
	return false
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
