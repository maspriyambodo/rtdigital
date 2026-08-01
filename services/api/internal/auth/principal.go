package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrPermissionDenied = errors.New("permission denied")

type Principal struct {
	UserID         string
	OrganizationID string
	SessionID      string
	MFA            bool
	RoleCodes      []string
	Permissions    map[string]struct{}
}

func (p *Principal) HasPermission(code string) bool {
	_, ok := p.Permissions[code]
	return ok
}

type principalContextKey struct{}

func PrincipalFromContext(ctx context.Context) *Principal {
	principal, _ := ctx.Value(principalContextKey{}).(*Principal)
	return principal
}

type AuthorizationService struct {
	db *pgxpool.Pool
}

func NewAuthorizationService(db *pgxpool.Pool) *AuthorizationService {
	return &AuthorizationService{db: db}
}

func (s *AuthorizationService) BuildPrincipal(ctx context.Context, claims *TokenClaims) (*Principal, error) {
	if claims == nil || claims.UserID == "" || claims.OrganizationID == "" || claims.SessionID == "" {
		return nil, ErrInvalidToken
	}

	principal := &Principal{
		UserID:         claims.UserID,
		OrganizationID: claims.OrganizationID,
		SessionID:      claims.SessionID,
		MFA:            claims.MFA,
		Permissions:    make(map[string]struct{}),
	}

	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT r.code, p.code
		FROM users u
		JOIN sessions s ON s.id = $3 AND s.user_id = u.id
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles r ON r.id = ur.role_id
		JOIN role_permissions rp ON rp.role_id = r.id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE u.id = $1
		  AND u.organization_id = $2
		  AND u.status = 'active'
		  AND s.organization_id = $2
		  AND s.revoked_at IS NULL
		  AND s.expires_at > now()
		  AND (r.organization_id IS NULL OR r.organization_id = $2)`,
		claims.UserID, claims.OrganizationID, claims.SessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("query effective permissions: %w", err)
	}
	defer rows.Close()

	roleCodes := make(map[string]struct{})
	for rows.Next() {
		var roleCode, permissionCode string
		if err := rows.Scan(&roleCode, &permissionCode); err != nil {
			return nil, fmt.Errorf("scan effective permission: %w", err)
		}
		roleCodes[roleCode] = struct{}{}
		principal.Permissions[permissionCode] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate effective permissions: %w", err)
	}

	for code := range roleCodes {
		principal.RoleCodes = append(principal.RoleCodes, code)
	}
	return principal, nil
}

func RequirePermission(authz *AuthorizationService, code string, requireMFA bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Access token wajib diisi.")
			return
		}
		if requireMFA && !claims.MFA {
			writeError(w, r, http.StatusForbidden, "MFA_REQUIRED", "Verifikasi MFA diperlukan.")
			return
		}

		principal, err := authz.BuildPrincipal(r.Context(), claims)
		if err != nil {
			writeError(w, r, http.StatusUnauthorized, "SESSION_INVALID", "Sesi tidak valid atau telah berakhir.")
			return
		}
		if !principal.HasPermission(code) {
			writeError(w, r, http.StatusForbidden, "PERMISSION_DENIED", "Akses tidak diizinkan untuk tindakan ini.")
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalContextKey{}, principal)))
	})
}

func RequireAuthenticatedPermission(tokens *TokenManager, authz *AuthorizationService, code string, requireMFA bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix = "Bearer "
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, prefix) {
			writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Access token wajib diisi.")
			return
		}
		claims, err := tokens.VerifyAccessToken(strings.TrimPrefix(header, prefix))
		if err != nil {
			writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Access token tidak valid atau kedaluwarsa.")
			return
		}
		RequirePermission(authz, code, requireMFA, next).ServeHTTP(
			w,
			r.WithContext(context.WithValue(r.Context(), claimsContextKey{}, claims)),
		)
	})
}
