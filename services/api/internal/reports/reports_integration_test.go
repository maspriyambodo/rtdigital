package reports_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
	"github.com/maspriyambodo/rtdigital/services/api/internal/cash"
	"github.com/maspriyambodo/rtdigital/services/api/internal/dashboard"
	"github.com/maspriyambodo/rtdigital/services/api/internal/httpapi"
	"github.com/maspriyambodo/rtdigital/services/api/internal/invoices"
	"github.com/maspriyambodo/rtdigital/services/api/internal/payments"
	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
	"github.com/maspriyambodo/rtdigital/services/api/internal/reports"
	"github.com/maspriyambodo/rtdigital/services/api/internal/residents"
	"github.com/maspriyambodo/rtdigital/services/api/internal/users"
)

func TestReportsIntegration(t *testing.T) {
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

	cashService := cash.NewService(pool)
	reportsService := reports.NewService(pool)
	server := httpapi.NewServer(
		slog.New(slog.NewTextHandler(os.Stderr, nil)),
		pool,
		tokens,
		auth.NewService(pool, tokens, crypter, auth.NoopMailer{}, "http://localhost:3000"),
		auth.NewAuthorizationService(pool),
		users.NewService(pool, auth.NoopMailer{}, "http://localhost:3000"),
		residents.NewService(pool, crypter, "12345678901234567890123456789012"),
		invoices.NewService(pool),
		nil,
		payments.NewService(pool, cashService),
		cashService,
		false,
		nil, nil, nil, nil,
		dashboard.NewService(pool),
		reportsService,
		nil,
	)

	sequence := uint64(time.Now().UnixNano())
	nextID := func(segment string) string {
		sequence++
		return fmt.Sprintf("%08x-%s-4000-8000-%012x", sequence&0xffffffff, segment, sequence&0xffffffffffff)
	}

	createOrganization := func(name string) string {
		t.Helper()
		id := nextID("9001")
		if _, err := pool.Exec(ctx, `
			INSERT INTO organizations (id, name, rt_number, rw_number, status)
			VALUES ($1, $2, '01', '01', 'active')`,
			id, name,
		); err != nil {
			t.Fatalf("create organization: %v", err)
		}
		return id
	}

	createBendahara := func(orgID string) (string, string) {
		t.Helper()
		userID, sessionID := nextID("1000"), nextID("2000")
		if _, err := pool.Exec(ctx, `
			INSERT INTO users (id, organization_id, email, password_hash, status)
			VALUES ($1, $2, $3, '!', 'active')`,
			userID, orgID, fmt.Sprintf("bendahara-%s@example.test", userID),
		); err != nil {
			t.Fatalf("create user: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id)
			SELECT $1, id FROM roles WHERE code = 'bendahara' AND organization_id IS NULL`,
			userID,
		); err != nil {
			t.Fatalf("assign bendahara role: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO sessions (id, user_id, organization_id, refresh_token_hash, expires_at)
			VALUES ($1, $2, $3, $4, now() + interval '1 hour')`,
			sessionID, userID, orgID, "refresh-"+sessionID,
		); err != nil {
			t.Fatalf("create session: %v", err)
		}
		token, err := tokens.IssueAccessToken(auth.TokenClaims{
			UserID: userID, OrganizationID: orgID, SessionID: sessionID, MFA: true,
		})
		if err != nil {
			t.Fatalf("issue access token: %v", err)
		}
		return userID, token
	}

	request := func(token, path string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, req)
		return recorder
	}

	orgID := createOrganization("RT Laporan Satu")
	otherOrgID := createOrganization("RT Laporan Dua")
	userID, token := createBendahara(orgID)

	residentID := nextID("3001")
	otherResidentID := nextID("3002")
	if _, err := pool.Exec(ctx, `
		INSERT INTO residents (id, organization_id, full_name, resident_status, verification_status)
		VALUES
			($1, $2, 'Warga Aktif Laporan', 'active', 'verified'),
			($3, $4, 'Warga Tenant Lain', 'active', 'verified')`,
		residentID, orgID, otherResidentID, otherOrgID,
	); err != nil {
		t.Fatalf("create residents: %v", err)
	}

	if _, err := pool.Exec(ctx, `
		INSERT INTO cash_transactions (
			id, organization_id, transaction_number, type, amount, transaction_date,
			description, status, created_by
		) VALUES
			($1, $2, 'KAS-REPORT-ONE', 'income', 150000, current_date, 'Kas organisasi satu', 'active', $3),
			($4, $5, 'KAS-REPORT-TWO', 'income', 900000, current_date, 'Kas organisasi dua', 'active', $3)`,
		nextID("4001"), orgID, userID, nextID("4002"), otherOrgID,
	); err != nil {
		t.Fatalf("create cash transactions: %v", err)
	}

	t.Run("dashboard aggregate excludes another tenant", func(t *testing.T) {
		response := request(token, "/api/v1/dashboard/admin")
		if response.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
		}

		var payload struct {
			Data struct {
				ActiveResidents int     `json:"active_residents"`
				CashBalance     float64 `json:"cash_balance"`
			} `json:"data"`
		}
		if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
			t.Fatalf("decode dashboard response: %v", err)
		}
		if payload.Data.ActiveResidents != 1 {
			t.Errorf("expected 1 resident in tenant, got %d", payload.Data.ActiveResidents)
		}
		if payload.Data.CashBalance != 150000 {
			t.Errorf("expected tenant cash balance 150000, got %.2f", payload.Data.CashBalance)
		}
	})

	t.Run("CSV export filters tenant, records audit", func(t *testing.T) {
		response := request(token, "/api/v1/reports/residents?format=csv&status=active")
		if response.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
		}
		if contentType := response.Header().Get("Content-Type"); contentType != "text/csv; charset=utf-8" {
			t.Errorf("expected CSV content type, got %q", contentType)
		}
		if !strings.Contains(response.Body.String(), "Warga Aktif Laporan") {
			t.Error("CSV missing current tenant resident")
		}
		if strings.Contains(response.Body.String(), "Warga Tenant Lain") {
			t.Error("CSV leaked another tenant resident")
		}

		var count int
		if err := pool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM audit_logs
			WHERE organization_id = $1
			  AND actor_user_id = $2
			  AND action = 'report.export'
			  AND entity_type = 'residents'
			  AND metadata->>'record_count' = '1'`,
			orgID, userID,
		).Scan(&count); err != nil {
			t.Fatalf("query export audit log: %v", err)
		}
		if count != 1 {
			t.Errorf("expected one scoped export audit log, got %d", count)
		}
	})

	t.Run("invalid date rejected", func(t *testing.T) {
		response := request(token, "/api/v1/reports/residents?start_date=not-a-date")
		if response.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d: %s", response.Code, response.Body.String())
		}
	})
}