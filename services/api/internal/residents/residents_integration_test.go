package residents_test

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

func TestResidentsIntegration(t *testing.T) {
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
	server := httpapi.NewServer(logger, pool, tokens, authService, authz, usersService, residentsService, invoicesService, false)

	suffix := time.Now().UnixNano() & 0xffffffff
	orgID := fmt.Sprintf("%08x-3333-3333-3333-333333333333", suffix)
	otherOrgID := fmt.Sprintf("%08x-4444-4444-4444-444444444444", suffix)
	if _, err := pool.Exec(ctx, `
		INSERT INTO organizations (id, name, rt_number, rw_number, status)
		VALUES ($1, 'RT Residents Test', '01', '01', 'active'),
		       ($2, 'RT Residents Other', '02', '01', 'active')`, orgID, otherOrgID); err != nil {
		t.Fatalf("create organizations: %v", err)
	}
	createTestUser := func(roleCode, organizationID string) (string, string) {
		t.Helper()

		id := fmt.Sprintf("%08x-0000-4000-8000-%012x", suffix, suffix)
		suffix++
		email := fmt.Sprintf("residents-%s-%d@example.test", roleCode, suffix)
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

	_, adminToken := createTestUser("sekretaris", orgID)
	_, otherAdminToken := createTestUser("sekretaris", otherOrgID)

	doRequest := func(method, path, token string, body []byte) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest(method, path, bytes.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		return response
	}

	t.Run("create and isolate house unit", func(t *testing.T) {
		res := doRequest(http.MethodPost, "/api/v1/house-units", adminToken, []byte(`{"code":"A1","occupancy_status":"owned"}`))
		if res.Code != http.StatusCreated {
			t.Fatalf("create house unit status = %d, want %d: %s", res.Code, http.StatusCreated, res.Body.String())
		}

		var payload struct {
			Data residents.HouseUnit `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
			t.Fatalf("parse response: %v", err)
		}

		res = doRequest(http.MethodGet, "/api/v1/house-units/"+payload.Data.ID, otherAdminToken, nil)
		if res.Code != http.StatusNotFound {
			t.Fatalf("cross-tenant status = %d, want %d", res.Code, http.StatusNotFound)
		}
	})

	t.Run("create household and assign active head", func(t *testing.T) {
		res := doRequest(http.MethodPost, "/api/v1/house-units", adminToken, []byte(`{"code":"A2","occupancy_status":"rented"}`))
		if res.Code != http.StatusCreated {
			t.Fatalf("create unit status = %d: %s", res.Code, res.Body.String())
		}
		var unit struct {
			Data residents.HouseUnit `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &unit); err != nil {
			t.Fatalf("parse unit: %v", err)
		}

		body := []byte(fmt.Sprintf(`{"house_unit_id":"%s","internal_number":"KK-01","domicile_status":"permanent"}`, unit.Data.ID))
		res = doRequest(http.MethodPost, "/api/v1/households", adminToken, body)
		if res.Code != http.StatusCreated {
			t.Fatalf("create household status = %d: %s", res.Code, res.Body.String())
		}
		var household struct {
			Data residents.Household `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &household); err != nil {
			t.Fatalf("parse household: %v", err)
		}

		res = doRequest(http.MethodPost, "/api/v1/residents", adminToken, []byte(`{"full_name":"Kepala Keluarga","resident_status":"active"}`))
		if res.Code != http.StatusCreated {
			t.Fatalf("create household resident status = %d: %s", res.Code, res.Body.String())
		}
		var resident struct {
			Data residents.Resident `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &resident); err != nil {
			t.Fatalf("parse household resident: %v", err)
		}

		body = []byte(fmt.Sprintf(`{"resident_id":"%s","relationship":"head"}`, resident.Data.ID))
		res = doRequest(http.MethodPost, "/api/v1/households/"+household.Data.ID+"/members", adminToken, body)
		if res.Code != http.StatusNoContent {
			t.Fatalf("assign head status = %d: %s", res.Code, res.Body.String())
		}
	})

	t.Run("encrypt, mask, and audit national ID", func(t *testing.T) {
		res := doRequest(http.MethodPost, "/api/v1/residents", adminToken, []byte(`{"full_name":"Budi Santoso","national_id":"1234567890123456","resident_status":"active"}`))
		if res.Code != http.StatusCreated {
			t.Fatalf("create resident status = %d: %s", res.Code, res.Body.String())
		}

		var payload struct {
			Data residents.Resident `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
			t.Fatalf("parse response: %v", err)
		}
		residentID := payload.Data.ID

		res = doRequest(http.MethodGet, "/api/v1/residents/"+residentID, adminToken, nil)
		if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
			t.Fatalf("parse masked response: %v", err)
		}
		if payload.Data.NationalID == nil || *payload.Data.NationalID != "••••" {
			t.Fatalf("masked national ID = %v, want ••••", payload.Data.NationalID)
		}

		res = doRequest(http.MethodGet, "/api/v1/residents/"+residentID+"?reason=Verifikasi%20KTP", adminToken, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("sensitive read status = %d: %s", res.Code, res.Body.String())
		}
		if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
			t.Fatalf("parse sensitive response: %v", err)
		}
		if payload.Data.NationalID == nil || *payload.Data.NationalID != "1234567890123456" {
			t.Fatalf("decrypted national ID = %v", payload.Data.NationalID)
		}

		var auditExists bool
		if err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM audit_logs
				WHERE organization_id = $1 AND entity_id = $2 AND action = 'resident.read_sensitive'
			)`, orgID, residentID).Scan(&auditExists); err != nil {
			t.Fatalf("check audit: %v", err)
		}
		if !auditExists {
			t.Fatal("sensitive access audit not found")
		}
	})

	t.Run("resident corrections workflow", func(t *testing.T) {
		wargaUserID, wargaToken := createTestUser("warga", orgID)

		res := doRequest(http.MethodPost, "/api/v1/residents", adminToken, []byte(`{"full_name":"Warga Koreksi","resident_status":"active"}`))
		if res.Code != http.StatusCreated {
			t.Fatalf("create resident status = %d: %s", res.Code, res.Body.String())
		}
		var resident struct {
			Data residents.Resident `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &resident); err != nil {
			t.Fatalf("parse resident: %v", err)
		}
		if _, err := pool.Exec(ctx, `UPDATE users SET resident_id = $1 WHERE id = $2`, resident.Data.ID, wargaUserID); err != nil {
			t.Fatalf("link user resident: %v", err)
		}

		res = doRequest(http.MethodPost, "/api/v1/residents/"+resident.Data.ID+"/corrections", wargaToken, []byte(`{"requested_changes":{"phone":"081234567890"},"reason":"Ganti nomor"}`))
		if res.Code != http.StatusCreated {
			t.Fatalf("submit correction status = %d: %s", res.Code, res.Body.String())
		}
		var correction struct {
			Data residents.ResidentCorrection `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &correction); err != nil {
			t.Fatalf("parse correction: %v", err)
		}

		res = doRequest(http.MethodPost, "/api/v1/resident-corrections/"+correction.Data.ID+"/approve", adminToken, []byte(`{"note":"Disetujui"}`))
		if res.Code != http.StatusNoContent {
			t.Fatalf("approve correction status = %d: %s", res.Code, res.Body.String())
		}

		res = doRequest(http.MethodGet, "/api/v1/residents/"+resident.Data.ID, adminToken, nil)
		if err := json.Unmarshal(res.Body.Bytes(), &resident); err != nil {
			t.Fatalf("parse updated resident: %v", err)
		}
		if resident.Data.Phone == nil || *resident.Data.Phone != "081234567890" {
			t.Fatalf("phone = %v, want 081234567890", resident.Data.Phone)
		}
	})

	t.Run("csv dry run and import", func(t *testing.T) {
		csvContent := "full_name,resident_status,national_id,phone\nWarga Impor Satu,active,9876543210987654,0811111111\nWarga Impor Dua,active,9876543210987655,0822222222\n"

		doImport := func(path string) *httptest.ResponseRecorder {
			request := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(csvContent))
			request.Header.Set("Authorization", "Bearer "+adminToken)
			request.Header.Set("Content-Type", "text/csv")
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			return response
		}

		res := doImport("/api/v1/residents/import?dry_run=true")
		if res.Code != http.StatusOK {
			t.Fatalf("dry run status = %d: %s", res.Code, res.Body.String())
		}
		var dryRun struct {
			Data residents.ImportResult `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &dryRun); err != nil {
			t.Fatalf("parse dry run: %v", err)
		}
		if dryRun.Data.ValidRows != 2 || dryRun.Data.ImportedRows != 0 {
			t.Fatalf("unexpected dry run: %+v", dryRun.Data)
		}

		res = doImport("/api/v1/residents/import")
		if res.Code != http.StatusOK {
			t.Fatalf("import status = %d: %s", res.Code, res.Body.String())
		}
		var imported struct {
			Data residents.ImportResult `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &imported); err != nil {
			t.Fatalf("parse import: %v", err)
		}
		if imported.Data.ImportedRows != 2 {
			t.Fatalf("imported rows = %d, want 2", imported.Data.ImportedRows)
		}
	})
}
