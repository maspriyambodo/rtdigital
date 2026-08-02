package invoices_test

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

func TestInvoicesIntegration(t *testing.T) {
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

	authService := auth.NewService(pool, tokens, crypter, auth.NoopMailer{}, "http://localhost:3000")
	authz := auth.NewAuthorizationService(pool)
	usersService := users.NewService(pool, auth.NoopMailer{}, "http://localhost:3000")
	residentsService := residents.NewService(pool, crypter, "12345678901234567890123456789012")
	server := httpapi.NewServer(
		slog.New(slog.NewTextHandler(os.Stderr, nil)),
		pool,
		tokens,
		authService,
		authz,
		usersService,
		residentsService,
		invoices.NewService(pool),
		nil,
		nil,
		false,
	)

	suffix := time.Now().UnixNano() & 0xffffffff
	nextID := func(segment string) string {
		id := fmt.Sprintf("%08x-%s-4000-8000-%012x", suffix, segment, suffix)
		suffix++
		return id
	}
	orgID := nextID("5555")
	otherOrgID := nextID("6666")
	if _, err := pool.Exec(ctx, `
		INSERT INTO organizations (id, name, rt_number, rw_number, status)
		VALUES ($1, 'RT Tagihan', '01', '01', 'active'),
		       ($2, 'RT Lain', '02', '01', 'active')`,
		orgID, otherOrgID,
	); err != nil {
		t.Fatalf("create organizations: %v", err)
	}

	createUser := func(role, organizationID string) (string, string) {
		t.Helper()
		userID := nextID("0000")
		sessionID := nextID("1111")
		hash, err := auth.HashPassword("Password123!")
		if err != nil {
			t.Fatalf("hash password: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO users (id, organization_id, email, password_hash, status)
			VALUES ($1, $2, $3, $4, 'active')`,
			userID, organizationID, fmt.Sprintf("%s-%s@example.test", role, userID), hash,
		); err != nil {
			t.Fatalf("create user: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id)
			SELECT $1, id FROM roles
			WHERE code = $2 AND organization_id IS NULL`,
			userID, role,
		); err != nil {
			t.Fatalf("assign role: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO sessions (id, user_id, organization_id, refresh_token_hash, expires_at)
			VALUES ($1, $2, $3, $4, now() + interval '1 hour')`,
			sessionID, userID, organizationID, fmt.Sprintf("refresh-hash-%s", sessionID),
		); err != nil {
			t.Fatalf("create session: %v", err)
		}
		token, err := tokens.IssueAccessToken(auth.TokenClaims{
			UserID: userID, OrganizationID: organizationID, SessionID: sessionID, MFA: true,
		})
		if err != nil {
			t.Fatalf("issue token: %v", err)
		}
		return userID, token
	}

	bendaharaID, bendaharaToken := createUser("bendahara", orgID)
	wargaOneID, wargaOneToken := createUser("warga", orgID)
	_, wargaTwoToken := createUser("warga", orgID)
	_, otherBendaharaToken := createUser("bendahara", otherOrgID)

	createHousehold := func(code, number, linkedUserID string) string {
		t.Helper()
		unitID, householdID, residentID, memberID := nextID("2222"), nextID("3333"), nextID("4444"), nextID("7777")
		if _, err := pool.Exec(ctx, `
			INSERT INTO house_units (id, organization_id, code, occupancy_status, status)
			VALUES ($1, $2, $3, 'owned', 'active')`,
			unitID, orgID, code,
		); err != nil {
			t.Fatalf("create house unit: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO residents (id, organization_id, full_name, resident_status, verification_status)
			VALUES ($1, $2, $3, 'active', 'verified')`,
			residentID, orgID, number,
		); err != nil {
			t.Fatalf("create resident: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO households (id, organization_id, house_unit_id, internal_number, domicile_status, verification_status)
			VALUES ($1, $2, $3, $4, 'permanent', 'verified')`,
			householdID, orgID, unitID, number,
		); err != nil {
			t.Fatalf("create household: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO household_members (id, organization_id, household_id, resident_id, relationship, is_active)
			VALUES ($1, $2, $3, $4, 'head', true)`,
			memberID, orgID, householdID, residentID,
		); err != nil {
			t.Fatalf("create household member: %v", err)
		}
		if _, err := pool.Exec(ctx, `UPDATE households SET head_resident_id = $1 WHERE id = $2`, residentID, householdID); err != nil {
			t.Fatalf("set household head: %v", err)
		}
		if linkedUserID != "" {
			if _, err := pool.Exec(ctx, `UPDATE users SET resident_id = $1 WHERE id = $2`, residentID, linkedUserID); err != nil {
				t.Fatalf("link resident account: %v", err)
			}
		}
		return householdID
	}

	householdOneID := createHousehold("A-01", "KK-A01", wargaOneID)
	_ = createHousehold("A-02", "KK-A02", "")

	request := func(method, path, token string, body string, headers map[string]string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		for key, value := range headers {
			req.Header.Set(key, value)
		}
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)
		return res
	}

	var dueTypeID string
	t.Run("due type CRUD and tenant isolation", func(t *testing.T) {
		res := request(http.MethodPost, "/api/v1/due-types", bendaharaToken, `{"name":"Iuran Kebersihan","amount":50000,"frequency":"monthly","due_day":10}`, nil)
		if res.Code != http.StatusCreated {
			t.Fatalf("create due type = %d: %s", res.Code, res.Body.String())
		}
		var result struct {
			Data invoices.DueType `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &result); err != nil {
			t.Fatalf("decode due type: %v", err)
		}
		dueTypeID = result.Data.ID

		res = request(http.MethodPatch, "/api/v1/due-types/"+dueTypeID, bendaharaToken, `{"description":"Iuran lingkungan"}`, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("update due type = %d: %s", res.Code, res.Body.String())
		}
		res = request(http.MethodGet, "/api/v1/due-types/"+dueTypeID, otherBendaharaToken, "", nil)
		if res.Code != http.StatusNotFound {
			t.Fatalf("cross-tenant due type = %d, want 404", res.Code)
		}
	})

	var invoiceID string
	t.Run("individual invoice adjustment cancellation and audit", func(t *testing.T) {
		body := fmt.Sprintf(`{"household_id":%q,"due_type_id":%q,"period_start":"2026-08-01","period_end":"2026-08-31","due_date":"2026-08-10","adjustment_amount":5000,"adjustment_reason":"Diskon awal"}`, householdOneID, dueTypeID)
		res := request(http.MethodPost, "/api/v1/invoices", bendaharaToken, body, nil)
		if res.Code != http.StatusCreated {
			t.Fatalf("create invoice = %d: %s", res.Code, res.Body.String())
		}
		var result struct {
			Data invoices.Invoice `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &result); err != nil {
			t.Fatalf("decode invoice: %v", err)
		}
		invoiceID = result.Data.ID
		if result.Data.InvoiceNumber == "" || result.Data.AdjustmentAmount != 5000 || result.Data.Status != "unpaid" {
			t.Fatalf("unexpected invoice: %+v", result.Data)
		}

		res = request(http.MethodPost, "/api/v1/invoices/"+invoiceID+"/cancel", bendaharaToken, `{"reason":"Salah periode"}`, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("cancel invoice = %d: %s", res.Code, res.Body.String())
		}
		var auditExists bool
		if err := pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM audit_logs
				WHERE organization_id = $1 AND actor_user_id = $2
				  AND entity_id = $3 AND action = 'invoice.cancel'
			)`,
			orgID, bendaharaID, invoiceID,
		).Scan(&auditExists); err != nil || !auditExists {
			t.Fatalf("cancel audit = %v, err = %v", auditExists, err)
		}
	})

	t.Run("bulk generation is idempotent", func(t *testing.T) {
		body := fmt.Sprintf(`{"due_type_id":%q,"period_start":"2026-09-01","period_end":"2026-09-30","due_date":"2026-09-10"}`, dueTypeID)
		headers := map[string]string{"Idempotency-Key": "epic4-bulk-2026-09"}
		first := request(http.MethodPost, "/api/v1/invoices/generate", bendaharaToken, body, headers)
		if first.Code != http.StatusCreated {
			t.Fatalf("generate bulk = %d: %s", first.Code, first.Body.String())
		}
		var firstResult struct {
			Data invoices.GenerateInvoicesResult `json:"data"`
		}
		if err := json.Unmarshal(first.Body.Bytes(), &firstResult); err != nil {
			t.Fatalf("decode bulk result: %v", err)
		}
		if firstResult.Data.TotalCreated != 2 {
			t.Fatalf("created = %d, want 2", firstResult.Data.TotalCreated)
		}

		second := request(http.MethodPost, "/api/v1/invoices/generate", bendaharaToken, body, headers)
		if second.Code != http.StatusCreated {
			t.Fatalf("repeat bulk = %d: %s", second.Code, second.Body.String())
		}
		var secondResult struct {
			Data invoices.GenerateInvoicesResult `json:"data"`
		}
		if err := json.Unmarshal(second.Body.Bytes(), &secondResult); err != nil {
			t.Fatalf("decode replay result: %v", err)
		}
		if secondResult.Data.TotalCreated != firstResult.Data.TotalCreated {
			t.Fatalf("replay created = %d, want %d", secondResult.Data.TotalCreated, firstResult.Data.TotalCreated)
		}
	})

	t.Run("resident scope hides other household invoices", func(t *testing.T) {
		res := request(http.MethodGet, "/api/v1/invoices", wargaOneToken, "", nil)
		if res.Code != http.StatusOK {
			t.Fatalf("list own invoices = %d: %s", res.Code, res.Body.String())
		}
		var result struct {
			Data []invoices.Invoice `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &result); err != nil {
			t.Fatalf("decode resident invoices: %v", err)
		}
		if len(result.Data) == 0 {
			t.Fatal("resident received no own invoices")
		}
		for _, item := range result.Data {
			if item.HouseholdID != householdOneID {
				t.Fatalf("resident received household %s", item.HouseholdID)
			}
		}

		res = request(http.MethodGet, "/api/v1/invoices/"+invoiceID, wargaTwoToken, "", nil)
		if res.Code != http.StatusNotFound {
			t.Fatalf("other resident invoice access = %d, want 404", res.Code)
		}
	})
}
