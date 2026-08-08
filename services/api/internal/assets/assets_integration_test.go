package assets_test

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

	"github.com/maspriyambodo/rtdigital/services/api/internal/assets"
	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
	"github.com/maspriyambodo/rtdigital/services/api/internal/httpapi"
	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
)

func TestAssetsIntegration(t *testing.T) {
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

	assetsService := assets.NewService(pool)
	server := httpapi.NewServer(
		slog.New(slog.NewTextHandler(os.Stderr, nil)),
		pool,
		tokens,
		auth.NewService(pool, tokens, crypter, auth.NoopMailer{}, "http://localhost:3000"),
		auth.NewAuthorizationService(pool),
		nil, nil, nil, nil, nil, nil,
		false,
		assetsService,
	)

	sequence := uint64(time.Now().UnixNano())
	nextID := func(segment string) string {
		sequence++
		return fmt.Sprintf("%08x-%s-4000-8000-%012x", sequence&0xffffffff, segment, sequence&0xffffffffffff)
	}

	orgID := nextID("9101")
	if _, err := pool.Exec(ctx, `
		INSERT INTO organizations (id, name, rt_number, rw_number, status)
		VALUES ($1, 'RT Test Aset', '01', '01', 'active')`,
		orgID,
	); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	createUser := func(role string) (string, string) {
		t.Helper()
		userID, sessionID := nextID("1100"), nextID("2100")
		hash, err := auth.HashPassword("Password123!")
		if err != nil {
			t.Fatalf("hash password: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO users (id, organization_id, email, password_hash, status)
			VALUES ($1, $2, $3, $4, 'active')`,
			userID, orgID, fmt.Sprintf("%s-%s@example.test", role, userID), hash,
		); err != nil {
			t.Fatalf("create user: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id)
			SELECT $1, id FROM roles WHERE code = $2 AND organization_id IS NULL`,
			userID, role,
		); err != nil {
			t.Fatalf("assign role: %v", err)
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

	_, adminToken := createUser("ketua_rt")
	wargaID, wargaToken := createUser("warga")

	request := func(method, path, token, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)
		return res
	}

	res := request(http.MethodPost, "/api/v1/asset-categories", adminToken, `{"code":"CAT-01","name":"Tenda & Meja"}`)
	if res.Code != http.StatusCreated {
		t.Fatalf("create category = %d: %s", res.Code, res.Body.String())
	}
	var catRes struct {
		Data assets.AssetCategory `json:"data"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &catRes)

	res = request(http.MethodPost, "/api/v1/asset-locations", adminToken, `{"code":"LOC-01","name":"Gudang Balai RT"}`)
	if res.Code != http.StatusCreated {
		t.Fatalf("create location = %d: %s", res.Code, res.Body.String())
	}
	var locRes struct {
		Data assets.AssetLocation `json:"data"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &locRes)

	createAssetBody := fmt.Sprintf(`{
		"category_id": "%s",
		"location_id": "%s",
		"code": "AST-001",
		"name": "Tenda Lipat 3x3m",
		"condition": "good"
	}`, catRes.Data.ID, locRes.Data.ID)

	res = request(http.MethodPost, "/api/v1/assets", adminToken, createAssetBody)
	if res.Code != http.StatusCreated {
		t.Fatalf("create asset = %d: %s", res.Code, res.Body.String())
	}
	var assetRes struct {
		Data assets.Asset `json:"data"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &assetRes)

	loanBody := fmt.Sprintf(`{
		"asset_id": "%s",
		"loan_date": "2026-08-10",
		"due_date": "2026-08-12",
		"condition_out": "good",
		"notes": "Acara syukuran warga"
	}`, assetRes.Data.ID)

	res = request(http.MethodPost, "/api/v1/asset-loans", wargaToken, loanBody)
	if res.Code != http.StatusCreated {
		t.Fatalf("create loan = %d: %s", res.Code, res.Body.String())
	}
	var loanRes struct {
		Data assets.AssetLoan `json:"data"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &loanRes)
	if loanRes.Data.BorrowerID != wargaID {
		t.Fatalf("borrower_id mismatch: got %s, want %s", loanRes.Data.BorrowerID, wargaID)
	}

	res = request(http.MethodPost, "/api/v1/asset-loans/"+loanRes.Data.ID+"/review", adminToken, `{"action":"approve"}`)
	if res.Code != http.StatusOK {
		t.Fatalf("approve loan = %d: %s", res.Code, res.Body.String())
	}

	res = request(http.MethodPost, "/api/v1/asset-loans/"+loanRes.Data.ID+"/return", adminToken, `{"condition_in":"good","notes":"Dikembalikan utuh"}`)
	if res.Code != http.StatusOK {
		t.Fatalf("return loan = %d: %s", res.Code, res.Body.String())
	}
}
