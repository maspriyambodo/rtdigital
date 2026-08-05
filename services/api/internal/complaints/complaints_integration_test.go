package complaints_test

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
	"github.com/maspriyambodo/rtdigital/services/api/internal/complaints"
	"github.com/maspriyambodo/rtdigital/services/api/internal/httpapi"
	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
)

func TestComplaintsWorkflowIntegration(t *testing.T) {
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
	complaintsService := complaints.NewService(pool)
	server := httpapi.NewServer(
		slog.New(slog.NewTextHandler(os.Stderr, nil)),
		pool, tokens, authService, authz,
		nil, nil, nil, nil, nil, nil, false,
		complaintsService,
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
	createUser := func(roleCode, organizationID string) (string, string) {
		t.Helper()
		userID, sessionID := newID(), newID()
		hash, err := auth.HashPassword("Password123!")
		if err != nil {
			t.Fatalf("hash password: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO users (id, organization_id, email, password_hash, status)
			VALUES ($1, $2, $3, $4, 'active')`,
			userID, organizationID, fmt.Sprintf("complaint-%s-%s@example.test", roleCode, userID[:8]), hash); err != nil {
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

	_, sekretarisToken := createUser("sekretaris", orgID)
	pengurusID, pengurusToken := createUser("pengurus", orgID)
	_, warga1Token := createUser("warga", orgID)
	_, warga2Token := createUser("warga", orgID)
	_, otherOrgToken := createUser("warga", otherOrgID)

	categoryID := newID()
	otherCategoryID := newID()
	if _, err := pool.Exec(ctx, `
		INSERT INTO complaint_categories (id, organization_id, code, name, status)
		VALUES ($1, $2, 'kebersihan', 'Kebersihan', 'active'),
		       ($3, $4, 'lainnya', 'Lainnya', 'active')`,
		categoryID, orgID, otherCategoryID, otherOrgID,
	); err != nil {
		t.Fatalf("seed complaint categories: %v", err)
	}

	wargaCannotManageCategory := doRequest(http.MethodPost, "/api/v1/complaint-categories", warga1Token, map[string]any{
		"code": "warga", "name": "Kategori Warga",
	})
	if wargaCannotManageCategory.Code != http.StatusForbidden {
		t.Fatalf("resident create category = %d; expected 403", wargaCannotManageCategory.Code)
	}
	createCategory := doRequest(http.MethodPost, "/api/v1/complaint-categories", sekretarisToken, map[string]any{
		"code": "lingkungan", "name": "Lingkungan",
	})
	if createCategory.Code != http.StatusCreated {
		t.Fatalf("manager create category = %d: %s", createCategory.Code, createCategory.Body.String())
	}
	var createdCategory struct {
		Data complaints.ComplaintCategory `json:"data"`
	}
	if err := json.Unmarshal(createCategory.Body.Bytes(), &createdCategory); err != nil {
		t.Fatalf("decode created category: %v", err)
	}
	otherTenantCategory := doRequest(http.MethodGet, "/api/v1/complaint-categories/"+otherCategoryID, sekretarisToken, nil)
	if otherTenantCategory.Code != http.StatusNotFound {
		t.Fatalf("cross-tenant category = %d; expected 404", otherTenantCategory.Code)
	}

	createResponse := doRequest(http.MethodPost, "/api/v1/complaints", warga1Token, map[string]any{
		"complaint_category_id": categoryID,
		"title":                "Sampah menumpuk di blok A",
		"description":          "Sampah belum diangkut 3 hari",
		"location_description": "Depan pos kamling",
		"priority":             "high",
		"attachment_file_ids":  []string{},
	})
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create complaint = %d: %s", createResponse.Code, createResponse.Body.String())
	}
	var created struct {
		Data complaints.ComplaintItem `json:"data"`
	}
	if err := json.Unmarshal(createResponse.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode complaint: %v", err)
	}
	if created.Data.TicketNumber == "" || created.Data.Status != "new" || created.Data.ComplaintCategoryID != categoryID || created.Data.CategoryName != "Kebersihan" {
		t.Fatalf("unexpected complaint data: %+v", created.Data)
	}

	invalidCategory := doRequest(http.MethodPost, "/api/v1/complaints", warga1Token, map[string]any{
		"complaint_category_id": newID(),
		"title":                "Kategori tidak valid",
		"description":          "Harus ditolak.",
		"priority":             "normal",
		"attachment_file_ids":  []string{},
	})
	if invalidCategory.Code != http.StatusBadRequest {
		t.Fatalf("create complaint with invalid category = %d; expected 400: %s", invalidCategory.Code, invalidCategory.Body.String())
	}
	crossTenantCategory := doRequest(http.MethodPost, "/api/v1/complaints", warga1Token, map[string]any{
		"complaint_category_id": otherCategoryID,
		"title":                "Kategori lintas tenant",
		"description":          "Harus ditolak.",
		"priority":             "normal",
		"attachment_file_ids":  []string{},
	})
	if crossTenantCategory.Code != http.StatusBadRequest {
		t.Fatalf("create complaint with cross-tenant category = %d; expected 400: %s", crossTenantCategory.Code, crossTenantCategory.Body.String())
	}

	if _, err := pool.Exec(ctx, `UPDATE complaint_categories SET status = 'inactive' WHERE id = $1`, categoryID); err != nil {
		t.Fatalf("deactivate complaint category: %v", err)
	}
	inactiveCategory := doRequest(http.MethodPost, "/api/v1/complaints", warga1Token, map[string]any{
		"complaint_category_id": categoryID,
		"title":                "Kategori nonaktif",
		"description":          "Harus ditolak.",
		"priority":             "normal",
		"attachment_file_ids":  []string{},
	})
	if inactiveCategory.Code != http.StatusBadRequest {
		t.Fatalf("create complaint with inactive category = %d; expected 400: %s", inactiveCategory.Code, inactiveCategory.Body.String())
	}
	if _, err := pool.Exec(ctx, `UPDATE complaint_categories SET status = 'active' WHERE id = $1`, categoryID); err != nil {
		t.Fatalf("reactivate complaint category: %v", err)
	}

	filtered := doRequest(http.MethodGet, "/api/v1/complaints?complaint_category_id="+categoryID, warga1Token, nil)
	if filtered.Code != http.StatusOK {
		t.Fatalf("filter complaint category = %d: %s", filtered.Code, filtered.Body.String())
	}
	var filteredItems struct {
		Data []complaints.ComplaintItem `json:"data"`
	}
	if err := json.Unmarshal(filtered.Body.Bytes(), &filteredItems); err != nil {
		t.Fatalf("decode filtered complaints: %v", err)
	}
	if len(filteredItems.Data) != 1 || filteredItems.Data[0].ID != created.Data.ID {
		t.Fatalf("unexpected category filter result: %+v", filteredItems.Data)
	}

	warga2List := doRequest(http.MethodGet, "/api/v1/complaints", warga2Token, nil)
	if warga2List.Code != http.StatusOK {
		t.Fatalf("list complaint warga 2 = %d", warga2List.Code)
	}
	var warga2Items struct {
		Data []complaints.ComplaintItem `json:"data"`
	}
	if err := json.Unmarshal(warga2List.Body.Bytes(), &warga2Items); err != nil {
		t.Fatalf("decode warga2 items: %v", err)
	}
	for _, item := range warga2Items.Data {
		if item.ID == created.Data.ID {
			t.Fatal("warga 2 can see warga 1 complaint in list")
		}
	}

	warga2Detail := doRequest(http.MethodGet, "/api/v1/complaints/"+created.Data.ID, warga2Token, nil)
	if warga2Detail.Code != http.StatusForbidden {
		t.Fatalf("warga 2 get complaint = %d; expected 403", warga2Detail.Code)
	}
	otherOrgDetail := doRequest(http.MethodGet, "/api/v1/complaints/"+created.Data.ID, otherOrgToken, nil)
	if otherOrgDetail.Code != http.StatusNotFound {
		t.Fatalf("other org get complaint = %d; expected 404", otherOrgDetail.Code)
	}

	assignResponse := doRequest(http.MethodPost, "/api/v1/complaints/"+created.Data.ID+"/assign", sekretarisToken, map[string]any{
		"assigned_to": pengurusID,
	})
	if assignResponse.Code != http.StatusOK {
		t.Fatalf("assign complaint = %d: %s", assignResponse.Code, assignResponse.Body.String())
	}

	internalComment := doRequest(http.MethodPost, "/api/v1/complaints/"+created.Data.ID+"/comments", pengurusToken, map[string]any{
		"body":        "Catatan internal: koordinasi dengan dinas kebersihan",
		"is_internal": true,
	})
	if internalComment.Code != http.StatusCreated {
		t.Fatalf("internal comment = %d: %s", internalComment.Code, internalComment.Body.String())
	}
	wargaForbiddenInternal := doRequest(http.MethodPost, "/api/v1/complaints/"+created.Data.ID+"/comments", warga1Token, map[string]any{
		"body":        "Warga mencoba membuat catatan internal",
		"is_internal": true,
	})
	if wargaForbiddenInternal.Code != http.StatusForbidden {
		t.Fatalf("warga internal comment = %d; expected 403", wargaForbiddenInternal.Code)
	}
	wargaPublicComment := doRequest(http.MethodPost, "/api/v1/complaints/"+created.Data.ID+"/comments", warga1Token, map[string]any{
		"body":        "Mohon bantuannya segera ditindaklanjuti.",
		"is_internal": false,
	})
	if wargaPublicComment.Code != http.StatusCreated {
		t.Fatalf("warga public comment = %d: %s", wargaPublicComment.Code, wargaPublicComment.Body.String())
	}

	wargaView := doRequest(http.MethodGet, "/api/v1/complaints/"+created.Data.ID, warga1Token, nil)
	if wargaView.Code != http.StatusOK {
		t.Fatalf("warga view = %d", wargaView.Code)
	}
	var wargaViewData struct {
		Data complaints.ComplaintItem `json:"data"`
	}
	if err := json.Unmarshal(wargaView.Body.Bytes(), &wargaViewData); err != nil {
		t.Fatalf("decode warga view: %v", err)
	}
	for _, comment := range wargaViewData.Data.Comments {
		if comment.IsInternal {
			t.Fatal("internal comment leaked to resident view")
		}
	}

	inProgress := doRequest(http.MethodPost, "/api/v1/complaints/"+created.Data.ID+"/status", pengurusToken, map[string]any{
		"status": "in_progress",
	})
	if inProgress.Code != http.StatusOK {
		t.Fatalf("in_progress = %d: %s", inProgress.Code, inProgress.Body.String())
	}
	resolvedWithoutNote := doRequest(http.MethodPost, "/api/v1/complaints/"+created.Data.ID+"/status", pengurusToken, map[string]any{
		"status": "resolved",
	})
	if resolvedWithoutNote.Code != http.StatusBadRequest {
		t.Fatalf("resolved without note = %d; expected 400", resolvedWithoutNote.Code)
	}
	resolved := doRequest(http.MethodPost, "/api/v1/complaints/"+created.Data.ID+"/status", pengurusToken, map[string]any{
		"status":          "resolved",
		"resolution_note": "Sampah telah diangkut oleh truk pengangkut RT.",
	})
	if resolved.Code != http.StatusOK {
		t.Fatalf("resolved = %d: %s", resolved.Code, resolved.Body.String())
	}
	pengurusCannotClose := doRequest(http.MethodPost, "/api/v1/complaints/"+created.Data.ID+"/status", pengurusToken, map[string]any{
		"status":          "closed",
		"resolution_note": "Pengurus mencoba menutup",
	})
	if pengurusCannotClose.Code != http.StatusForbidden {
		t.Fatalf("pengurus close = %d; expected 403", pengurusCannotClose.Code)
	}
	closed := doRequest(http.MethodPost, "/api/v1/complaints/"+created.Data.ID+"/status", warga1Token, map[string]any{
		"status":          "closed",
		"resolution_note": "Terima kasih, lingkungan sudah bersih kembali.",
	})
	if closed.Code != http.StatusOK {
		t.Fatalf("warga close = %d: %s", closed.Code, closed.Body.String())
	}

	finalDetail := doRequest(http.MethodGet, "/api/v1/complaints/"+created.Data.ID, warga1Token, nil)
	if finalDetail.Code != http.StatusOK {
		t.Fatalf("final get = %d", finalDetail.Code)
	}
	var finalData struct {
		Data complaints.ComplaintItem `json:"data"`
	}
	if err := json.Unmarshal(finalDetail.Body.Bytes(), &finalData); err != nil {
		t.Fatalf("decode final data: %v", err)
	}
	if finalData.Data.Status != "closed" || finalData.Data.ClosedAt == nil || finalData.Data.ResolutionNote == nil {
		t.Fatalf("unexpected final state: %+v", finalData.Data)
	}
}