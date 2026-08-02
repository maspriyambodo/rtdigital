package letters_test

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
	"github.com/maspriyambodo/rtdigital/services/api/internal/letters"
	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
)

type mockStorage struct {
	objects map[string][]byte
}

func (m *mockStorage) PutObject(_ context.Context, key string, data []byte, _ string) error {
	if m.objects == nil {
		m.objects = make(map[string][]byte)
	}
	m.objects[key] = data
	return nil
}

func TestLettersWorkflowIntegration(t *testing.T) {
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
	lettersService := letters.NewService(pool)
	server := httpapi.NewServer(
		slog.New(slog.NewTextHandler(os.Stderr, nil)),
		pool, tokens, authService, authz,
		nil, nil, nil, nil, nil, nil, false,
		lettersService, &mockStorage{},
	)

	const orgID = "11111111-1111-1111-1111-111111111111"
	const otherOrgID = "22222222-2222-2222-2222-222222222222"
	if _, err := pool.Exec(ctx, `
		INSERT INTO organizations (id, name, rt_number, rw_number, status)
		VALUES ($1, 'RT Test', '01', '01', 'active'),
		       ($2, 'RT Lain', '02', '01', 'active')
		ON CONFLICT (id) DO NOTHING`, orgID, otherOrgID); err != nil {
		t.Fatalf("seed organizations: %v", err)
	}

	suffix := time.Now().UnixNano() & 0xffffffff
	newID := func() string {
		id := fmt.Sprintf("%08x-0000-4000-8000-%012x", suffix, suffix)
		suffix++
		return id
	}
	createResident := func(organizationID, name string) string {
		t.Helper()
		id := newID()
		if _, err := pool.Exec(ctx, `
			INSERT INTO residents (
				id, organization_id, full_name, encrypted_nik, nik_hash, birth_date,
				gender, religion, marital_status, occupation, status
			) VALUES ($1, $2, $3, 'enc', $4, '1990-01-01', 'male', 'islam', 'married', 'swasta', 'verified')`,
			id, organizationID, name, "nik-"+id); err != nil {
			t.Fatalf("create resident: %v", err)
		}
		return id
	}
	createUser := func(roleCode, organizationID string, residentID *string) (string, string) {
		t.Helper()
		userID := newID()
		sessionID := newID()
		hash, err := auth.HashPassword("Password123!")
		if err != nil {
			t.Fatalf("hash password: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO users (id, organization_id, email, password_hash, status, resident_id)
			VALUES ($1, $2, $3, $4, 'active', $5)`,
			userID, organizationID, fmt.Sprintf("letters-%s-%s@example.test", roleCode, userID[:8]), hash, residentID); err != nil {
			t.Fatalf("create user: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id)
			SELECT $1, id FROM roles WHERE code = $2 AND (organization_id IS NULL OR organization_id = $3)`,
			userID, roleCode, organizationID); err != nil {
			t.Fatalf("assign role: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO sessions (id, user_id, organization_id, refresh_token_hash, expires_at)
			VALUES ($1, $2, $3, $4, now() + interval '1 hour')`,
			sessionID, userID, organizationID, "hash-"+sessionID); err != nil {
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
	doRequest := func(method, path, token string, body any) *httptest.ResponseRecorder {
		t.Helper()
		var data []byte
		if body != nil {
			var err error
			data, err = json.Marshal(body)
			if err != nil {
				t.Fatalf("marshal request: %v", err)
			}
		}
		request := httptest.NewRequest(method, path, bytes.NewReader(data))
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		return response
	}

	residentID := createResident(orgID, "Warga Surat")
	_, ketuaToken := createUser("ketua_rt", orgID, nil)
	_, sekretarisToken := createUser("sekretaris", orgID, nil)
	_, wargaToken := createUser("warga", orgID, &residentID)
	_, otherOrgToken := createUser("warga", otherOrgID, nil)

	typeResponse := doRequest(http.MethodPost, "/api/v1/letter-types", sekretarisToken, map[string]any{
		"name":           "Surat Domisili " + fmt.Sprint(suffix),
		"requirements":   []any{},
		"form_schema":    map[string]any{"fields": []any{map[string]any{"name": "keperluan", "label": "Keperluan", "type": "text", "required": true}}},
		"template":       "Surat {LETTER_NUMBER} untuk {RESIDENT_NAME}: {KEPERLUAN}.",
		"number_pattern": "{NUM}/RT01/{YEAR}",
		"status":         "active",
	})
	if typeResponse.Code != http.StatusCreated {
		t.Fatalf("create type = %d: %s", typeResponse.Code, typeResponse.Body.String())
	}
	var typePayload struct {
		Data letters.LetterTypeItem `json:"data"`
	}
	if err := json.Unmarshal(typeResponse.Body.Bytes(), &typePayload); err != nil {
		t.Fatalf("decode type: %v", err)
	}

	submitResponse := doRequest(http.MethodPost, "/api/v1/letter-requests", wargaToken, map[string]any{
		"letter_type_id":       typePayload.Data.ID,
		"resident_id":          residentID,
		"form_data":            map[string]string{"keperluan": "KPR"},
		"attachment_file_ids":  []string{},
	})
	if submitResponse.Code != http.StatusCreated {
		t.Fatalf("submit = %d: %s", submitResponse.Code, submitResponse.Body.String())
	}
	var requestPayload struct {
		Data letters.LetterRequestItem `json:"data"`
	}
	if err := json.Unmarshal(submitResponse.Body.Bytes(), &requestPayload); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	requestID := requestPayload.Data.ID

	if response := doRequest(http.MethodGet, "/api/v1/letter-requests/"+requestID, otherOrgToken, nil); response.Code != http.StatusNotFound {
		t.Fatalf("cross-tenant read = %d, want 404: %s", response.Code, response.Body.String())
	}
	if response := doRequest(http.MethodPost, "/api/v1/letter-requests/"+requestID+"/process", sekretarisToken, nil); response.Code != http.StatusOK {
		t.Fatalf("process = %d: %s", response.Code, response.Body.String())
	}
	if response := doRequest(http.MethodPost, "/api/v1/letter-requests/"+requestID+"/request-revision", sekretarisToken, map[string]string{"resident_note": "Lengkapi alamat"}); response.Code != http.StatusOK {
		t.Fatalf("revision = %d: %s", response.Code, response.Body.String())
	}
	if response := doRequest(http.MethodPatch, "/api/v1/letter-requests/"+requestID, wargaToken, map[string]any{
		"letter_type_id": typePayload.Data.ID,
		"resident_id":    residentID,
		"form_data":      map[string]string{"keperluan": "KPR Bank"},
	}); response.Code != http.StatusOK {
		t.Fatalf("resubmit = %d: %s", response.Code, response.Body.String())
	}
	if response := doRequest(http.MethodPost, "/api/v1/letter-requests/"+requestID+"/process", sekretarisToken, nil); response.Code != http.StatusOK {
		t.Fatalf("reprocess = %d: %s", response.Code, response.Body.String())
	}
	if response := doRequest(http.MethodPost, "/api/v1/letter-requests/"+requestID+"/approve", ketuaToken, nil); response.Code != http.StatusOK {
		t.Fatalf("approve = %d: %s", response.Code, response.Body.String())
	}
	issued := doRequest(http.MethodPost, "/api/v1/letter-requests/"+requestID+"/issue", sekretarisToken, nil)
	if issued.Code != http.StatusOK {
		t.Fatalf("issue = %d: %s", issued.Code, issued.Body.String())
	}
	if err := json.Unmarshal(issued.Body.Bytes(), &requestPayload); err != nil {
		t.Fatalf("decode issued request: %v", err)
	}
	if requestPayload.Data.Status != "issued" || requestPayload.Data.LetterNumber == nil || *requestPayload.Data.LetterNumber == "" {
		t.Fatalf("unexpected issued request: %#v", requestPayload.Data)
	}
}