package users_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
	"github.com/maspriyambodo/rtdigital/services/api/internal/httpapi"
	"github.com/maspriyambodo/rtdigital/services/api/internal/invoices"
	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
	"github.com/maspriyambodo/rtdigital/services/api/internal/residents"
	"github.com/maspriyambodo/rtdigital/services/api/internal/users"
)

func TestAuthorizationAndRBAC(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set; skipping database integration test")
	}

	ctx := context.Background()
	pool, err := platform.NewDatabase(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	tokens, err := auth.NewTokenManager("test-secret-must-be-at-least-32-chars-long")
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	crypter, err := auth.NewAESCrypter([]byte("12345678901234567890123456789012"))
	if err != nil {
		t.Fatalf("create crypter: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	authService := auth.NewService(pool, tokens, crypter, auth.NoopMailer{}, "http://localhost:3000")
	authz := auth.NewAuthorizationService(pool)
	usersService := users.NewService(pool, auth.NoopMailer{}, "http://localhost:3000")
	residentsService := residents.NewService(pool, crypter, "12345678901234567890123456789012")
	invoicesService := invoices.NewService(pool)
	server := httpapi.NewServer(logger, pool, tokens, authService, authz, usersService, residentsService, invoicesService, nil, nil, nil, false)

	const orgID = "11111111-1111-1111-1111-111111111111"
	const otherOrgID = "22222222-2222-2222-2222-222222222222"
	if _, err := pool.Exec(ctx, `
		INSERT INTO organizations (id, name, rt_number, rw_number, status)
		VALUES ($1, 'RT Test', '01', '01', 'active'),
		       ($2, 'RT Test Lain', '02', '01', 'active')
		ON CONFLICT (id) DO NOTHING`, orgID, otherOrgID); err != nil {
		t.Fatalf("create organizations: %v", err)
	}

	suffix := time.Now().UnixNano() & 0xffffffff
	createTestUser := func(roleCode, organizationID string) (string, string) {
		t.Helper()

		id := fmt.Sprintf("%08x-0000-4000-8000-%012x", suffix, suffix)
		suffix++
		email := fmt.Sprintf("rbac-%s-%d@example.test", roleCode, suffix)
		hash, err := auth.HashPassword("Password123!")
		if err != nil {
			t.Fatalf("hash password: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO users (id, organization_id, email, password_hash, status)
			VALUES ($1, $2, $3, $4, 'active')`, id, organizationID, email, hash); err != nil {
			t.Fatalf("create %s user: %v", roleCode, err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id)
			SELECT $1, id FROM roles
			WHERE code = $2 AND (organization_id IS NULL OR organization_id = $3)`,
			id, roleCode, organizationID); err != nil {
			t.Fatalf("assign %s role: %v", roleCode, err)
		}

		sessionID := fmt.Sprintf("%08x-0000-4000-8000-%012x", suffix, suffix)
		suffix++
		if _, err := pool.Exec(ctx, `
			INSERT INTO sessions (id, user_id, organization_id, refresh_token_hash, expires_at)
			VALUES ($1, $2, $3, $4, now() + interval '1 hour')`,
			sessionID, id, organizationID, fmt.Sprintf("hash-%s", sessionID)); err != nil {
			t.Fatalf("create session: %v", err)
		}
		token, err := tokens.IssueAccessToken(auth.TokenClaims{
			UserID: id, OrganizationID: organizationID, SessionID: sessionID, MFA: true,
		})
		if err != nil {
			t.Fatalf("issue access token: %v", err)
		}
		return id, token
	}

	ketuaID, ketuaToken := createTestUser("ketua_rt", orgID)
	_, sekretarisToken := createTestUser("sekretaris", orgID)
	wargaID, wargaToken := createTestUser("warga", orgID)
	otherOrgUserID, _ := createTestUser("warga", otherOrgID)

	doRequest := func(method, path, token string, body []byte) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest(method, path, bytes.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		return response
	}

	t.Run("warga denied user list", func(t *testing.T) {
		if response := doRequest(http.MethodGet, "/api/v1/users", wargaToken, nil); response.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusForbidden, response.Body.String())
		}
	})

	t.Run("sekretaris cannot assign role", func(t *testing.T) {
		var roleID string
		if err := pool.QueryRow(ctx, `SELECT id FROM roles WHERE code = 'bendahara' LIMIT 1`).Scan(&roleID); err != nil {
			t.Fatalf("find role: %v", err)
		}
		body, _ := json.Marshal(map[string]string{"role_id": roleID})
		if response := doRequest(http.MethodPost, "/api/v1/users/"+wargaID+"/roles", sekretarisToken, body); response.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusForbidden, response.Body.String())
		}
	})

	t.Run("ketua can assign non-super-admin role", func(t *testing.T) {
		var roleID string
		if err := pool.QueryRow(ctx, `SELECT id FROM roles WHERE code = 'pengurus' LIMIT 1`).Scan(&roleID); err != nil {
			t.Fatalf("find role: %v", err)
		}
		body, _ := json.Marshal(map[string]string{"role_id": roleID})
		if response := doRequest(http.MethodPost, "/api/v1/users/"+wargaID+"/roles", ketuaToken, body); response.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusNoContent, response.Body.String())
		}
	})

	t.Run("mandatory MFA role assignment requires target MFA enrollment", func(t *testing.T) {
		targetUserID, _ := createTestUser("warga", orgID)

		var roleID string
		if err := pool.QueryRow(ctx, `SELECT id FROM roles WHERE code = 'bendahara' LIMIT 1`).Scan(&roleID); err != nil {
			t.Fatalf("find role: %v", err)
		}
		body, _ := json.Marshal(map[string]string{"role_id": roleID})

		response := doRequest(http.MethodPost, "/api/v1/users/"+targetUserID+"/roles", ketuaToken, body)
		if response.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusForbidden, response.Body.String())
		}
		if !bytes.Contains(response.Body.Bytes(), []byte("MFA_ENROLLMENT_REQUIRED")) {
			t.Fatalf("body = %s, want MFA_ENROLLMENT_REQUIRED", response.Body.String())
		}
	})

	t.Run("self role modification denied", func(t *testing.T) {
		var roleID string
		if err := pool.QueryRow(ctx, `SELECT id FROM roles WHERE code = 'warga' LIMIT 1`).Scan(&roleID); err != nil {
			t.Fatalf("find role: %v", err)
		}
		body, _ := json.Marshal(map[string]string{"role_id": roleID})
		if response := doRequest(http.MethodPost, "/api/v1/users/"+ketuaID+"/roles", ketuaToken, body); response.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusForbidden, response.Body.String())
		}
	})

	t.Run("cross-tenant user hidden", func(t *testing.T) {
		if response := doRequest(http.MethodGet, "/api/v1/users/"+otherOrgUserID, ketuaToken, nil); response.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusNotFound, response.Body.String())
		}
	})

	t.Run("sensitive endpoint requires MFA", func(t *testing.T) {
		claims, err := tokens.VerifyAccessToken(ketuaToken)
		if err != nil {
			t.Fatalf("verify test token: %v", err)
		}
		claims.MFA = false
		nonMFAToken, err := tokens.IssueAccessToken(*claims)
		if err != nil {
			t.Fatalf("issue non-MFA token: %v", err)
		}
		if response := doRequest(http.MethodGet, "/api/v1/users", nonMFAToken, nil); response.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusForbidden, response.Body.String())
		}
	})

	t.Run("office handover enforces MFA and revokes outgoing access", func(t *testing.T) {
		outgoingUserID, _ := createTestUser("bendahara", orgID)
		incomingUserID, _ := createTestUser("warga", orgID)
		createBody, err := json.Marshal(map[string]any{
			"outgoing_user_id": outgoingUserID,
			"checklist": map[string]bool{
				"akses": true,
				"rekening": true,
				"kas": true,
			},
		})
		if err != nil {
			t.Fatalf("marshal create handover: %v", err)
		}
		res := doRequest(http.MethodPost, "/api/v1/office-handovers", ketuaToken, createBody)
		if res.Code != http.StatusCreated {
			t.Fatalf("create handover = %d, want 201: %s", res.Code, res.Body.String())
		}
		var created struct {
			Data users.OfficeHandover `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &created); err != nil {
			t.Fatalf("decode created handover: %v", err)
		}

		completeBody, err := json.Marshal(map[string]any{
			"incoming_user_id": incomingUserID,
			"checklist": map[string]bool{
				"akses": true,
				"rekening": true,
				"kas": true,
			},
		})
		if err != nil {
			t.Fatalf("marshal complete handover: %v", err)
		}
		res = doRequest(http.MethodPost, "/api/v1/office-handovers/"+created.Data.ID+"/complete", ketuaToken, completeBody)
		if res.Code != http.StatusForbidden {
			t.Fatalf("complete without incoming MFA = %d, want 403: %s", res.Code, res.Body.String())
		}
		if !bytes.Contains(res.Body.Bytes(), []byte("MFA_ENROLLMENT_REQUIRED")) {
			t.Fatalf("missing MFA guard response: %s", res.Body.String())
		}

		if _, err := pool.Exec(ctx, `UPDATE users SET mfa_enabled_at = now() WHERE id = $1`, incomingUserID); err != nil {
			t.Fatalf("enable incoming MFA: %v", err)
		}
		res = doRequest(http.MethodPost, "/api/v1/office-handovers/"+created.Data.ID+"/complete", ketuaToken, completeBody)
		if res.Code != http.StatusOK {
			t.Fatalf("complete handover = %d, want 200: %s", res.Code, res.Body.String())
		}

		var incomingHasRole, outgoingHasRole, outgoingSessionsRevoked, auditExists bool
		if err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM user_roles ur
				JOIN roles r ON r.id = ur.role_id
				WHERE ur.user_id = $1 AND r.code = 'bendahara'
			)`, incomingUserID).Scan(&incomingHasRole); err != nil {
			t.Fatalf("check incoming role: %v", err)
		}
		if err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM user_roles ur
				JOIN roles r ON r.id = ur.role_id
				WHERE ur.user_id = $1 AND r.code = 'bendahara'
			)`, outgoingUserID).Scan(&outgoingHasRole); err != nil {
			t.Fatalf("check outgoing role: %v", err)
		}
		if err := pool.QueryRow(ctx, `
			SELECT NOT EXISTS (
				SELECT 1 FROM sessions WHERE user_id = $1 AND revoked_at IS NULL
			)`, outgoingUserID).Scan(&outgoingSessionsRevoked); err != nil {
			t.Fatalf("check outgoing sessions: %v", err)
		}
		if err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM audit_logs
				WHERE organization_id = $1
				  AND entity_id = $2
				  AND action = 'office_handover.complete'
			)`, orgID, created.Data.ID).Scan(&auditExists); err != nil {
			t.Fatalf("check handover audit: %v", err)
		}
		if !incomingHasRole || outgoingHasRole || !outgoingSessionsRevoked || !auditExists {
			t.Fatalf(
				"handover invariant failed: incoming_role=%t outgoing_role=%t sessions_revoked=%t audit=%t",
				incomingHasRole, outgoingHasRole, outgoingSessionsRevoked, auditExists,
			)
		}
	})
}
