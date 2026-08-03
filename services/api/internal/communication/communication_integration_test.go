package communication_test

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
	"github.com/maspriyambodo/rtdigital/services/api/internal/communication"
	"github.com/maspriyambodo/rtdigital/services/api/internal/files"
	"github.com/maspriyambodo/rtdigital/services/api/internal/httpapi"
	"github.com/maspriyambodo/rtdigital/services/api/internal/invoices"
	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
	"github.com/maspriyambodo/rtdigital/services/api/internal/residents"
	"github.com/maspriyambodo/rtdigital/services/api/internal/users"
)

type memoryStorage struct {
	objects map[string]platform.ObjectMetadata
}

func (m *memoryStorage) PresignUpload(_ context.Context, key, contentType string, _ time.Duration) (platform.PresignedURL, error) {
	m.objects[key] = platform.ObjectMetadata{SizeBytes: 123, ContentType: contentType}
	return platform.PresignedURL{
		URL:     "https://storage.test/upload/" + key,
		Headers: map[string]string{"Content-Type": contentType},
	}, nil
}

func (m *memoryStorage) HeadObject(_ context.Context, key string) (platform.ObjectMetadata, error) {
	item, ok := m.objects[key]
	if !ok {
		return platform.ObjectMetadata{}, fmt.Errorf("object %q not found", key)
	}
	return item, nil
}

func (m *memoryStorage) PresignDownload(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://storage.test/download/" + key, nil
}

func TestCommunicationIntegration(t *testing.T) {
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

	storage := &memoryStorage{objects: make(map[string]platform.ObjectMetadata)}
	communicationService := communication.NewService(pool)
	server := httpapi.NewServer(
		slog.New(slog.NewTextHandler(os.Stderr, nil)),
		pool,
		tokens,
		auth.NewService(pool, tokens, crypter, auth.NoopMailer{}, "http://localhost:3000"),
		auth.NewAuthorizationService(pool),
		users.NewService(pool, auth.NoopMailer{}, "http://localhost:3000"),
		residents.NewService(pool, crypter, "12345678901234567890123456789012"),
		invoices.NewService(pool),
		files.NewService(pool, storage),
		nil,
		nil,
		false,
		communicationService,
	)

	suffix := time.Now().UnixNano() & 0xffffffff
	nextID := func(segment string) string {
		id := fmt.Sprintf("%08x-%s-4000-8000-%012x", suffix, segment, suffix)
		suffix++
		return id
	}

	orgID := nextID("7701")
	otherOrgID := nextID("7702")
	if _, err := pool.Exec(ctx, `
		INSERT INTO organizations (id, name, rt_number, rw_number, status)
		VALUES ($1, 'RT Komunikasi', '01', '01', 'active'),
		       ($2, 'RT Lain Komunikasi', '02', '01', 'active')`,
		orgID, otherOrgID,
	); err != nil {
		t.Fatalf("create organizations: %v", err)
	}

	createUser := func(role, organizationID string) (string, string) {
		t.Helper()
		userID := nextID("1000")
		sessionID := nextID("2000")
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
			sessionID, userID, organizationID, "refresh-"+sessionID,
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

	wargaID, wargaToken := createUser("warga", orgID)
	_, otherWargaToken := createUser("warga", orgID)
	_, sekretarisToken := createUser("sekretaris", orgID)
	_, otherOrgToken := createUser("warga", otherOrgID)

	unitID, residentID, householdID := nextID("3000"), nextID("4000"), nextID("5000")
	if _, err := pool.Exec(ctx, `
		INSERT INTO house_units (id, organization_id, code, occupancy_status, status)
		VALUES ($1, $2, 'KOM-01', 'owned', 'active')`,
		unitID, orgID,
	); err != nil {
		t.Fatalf("create unit: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO residents (id, organization_id, full_name, resident_status, verification_status)
		VALUES ($1, $2, 'Warga Pembaca', 'active', 'verified')`,
		residentID, orgID,
	); err != nil {
		t.Fatalf("create resident: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO households (id, organization_id, house_unit_id, internal_number, domicile_status, verification_status)
		VALUES ($1, $2, $3, 'KK-KOM-01', 'permanent', 'verified')`,
		householdID, orgID, unitID,
	); err != nil {
		t.Fatalf("create household: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO household_members (id, organization_id, household_id, resident_id, relationship, is_active)
		VALUES ($1, $2, $3, $4, 'head', true)`,
		nextID("7000"), orgID, householdID, residentID,
	); err != nil {
		t.Fatalf("create household member: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE households SET head_resident_id = $1 WHERE id = $2`, residentID, householdID); err != nil {
		t.Fatalf("set head: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE users SET resident_id = $1 WHERE id = $2`, residentID, wargaID); err != nil {
		t.Fatalf("link user: %v", err)
	}

	request := func(method, path, token, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)
		return res
	}

	var announcementID string
	t.Run("draft, target visibility, attachment, publish, read statistics", func(t *testing.T) {
		res := request(
			http.MethodPost,
			"/api/v1/files/presign-upload",
			sekretarisToken,
			`{"entity_type":"announcement","entity_id":"temp-id","purpose":"attachment","original_name":"doc.pdf","mime_type":"application/pdf","size_bytes":123}`,
		)
		if res.Code != http.StatusOK {
			t.Fatalf("presign = %d: %s", res.Code, res.Body.String())
		}
		var upload struct {
			Data struct {
				FileID string `json:"file_id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &upload); err != nil {
			t.Fatalf("decode presign: %v", err)
		}

		res = request(http.MethodPost, "/api/v1/files/confirm-upload", sekretarisToken, fmt.Sprintf(`{"file_id":%q}`, upload.Data.FileID))
		if res.Code != http.StatusOK {
			t.Fatalf("confirm = %d: %s", res.Code, res.Body.String())
		}

		body := fmt.Sprintf(`{"title":"Gotong Royong","content":"Mari bersihkan selokan.","category":"event","priority":"normal","status":"draft","targets":[{"target_type":"all"}],"attachment_file_ids":[%q]}`, upload.Data.FileID)
		res = request(http.MethodPost, "/api/v1/announcements", sekretarisToken, body)
		if res.Code != http.StatusCreated {
			t.Fatalf("create announcement = %d: %s", res.Code, res.Body.String())
		}
		var created struct {
			Data communication.AnnouncementItem `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &created); err != nil {
			t.Fatalf("decode announcement: %v", err)
		}
		announcementID = created.Data.ID
		if len(created.Data.Attachments) != 1 {
			t.Fatalf("attachments = %d, want 1", len(created.Data.Attachments))
		}

		res = request(http.MethodGet, "/api/v1/announcements", wargaToken, "")
		if res.Code != http.StatusOK {
			t.Fatalf("list draft = %d: %s", res.Code, res.Body.String())
		}
		var list struct {
			Data []communication.AnnouncementItem `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &list); err != nil {
			t.Fatalf("decode draft list: %v", err)
		}
		for _, item := range list.Data {
			if item.ID == announcementID {
				t.Fatal("draft visible to resident")
			}
		}

		res = request(http.MethodPost, "/api/v1/announcements/"+announcementID+"/publish", otherOrgToken, "")
		if res.Code != http.StatusForbidden {
			t.Fatalf("cross-tenant publish by non-manager = %d, want 403", res.Code)
		}

		res = request(http.MethodPost, "/api/v1/announcements/"+announcementID+"/publish", sekretarisToken, "")
		if res.Code != http.StatusOK {
			t.Fatalf("publish = %d: %s", res.Code, res.Body.String())
		}

		res = request(http.MethodGet, "/api/v1/announcements", wargaToken, "")
		if res.Code != http.StatusOK {
			t.Fatalf("list published = %d: %s", res.Code, res.Body.String())
		}
		if err := json.Unmarshal(res.Body.Bytes(), &list); err != nil {
			t.Fatalf("decode published list: %v", err)
		}
		found := false
		for _, item := range list.Data {
			if item.ID == announcementID {
				found = true
				if item.IsRead {
					t.Fatal("announcement already read")
				}
			}
		}
		if !found {
			t.Fatal("published announcement not visible to target")
		}

		res = request(http.MethodGet, "/api/v1/announcements/"+announcementID, wargaToken, "")
		if res.Code != http.StatusOK {
			t.Fatalf("read announcement = %d: %s", res.Code, res.Body.String())
		}
		res = request(http.MethodGet, "/api/v1/files/"+upload.Data.FileID+"/download", wargaToken, "")
		if res.Code != http.StatusOK {
			t.Fatalf("target download attachment = %d: %s", res.Code, res.Body.String())
		}

		res = request(http.MethodGet, "/api/v1/announcements/"+announcementID+"/read-stats", sekretarisToken, "")
		if res.Code != http.StatusOK {
			t.Fatalf("get statistics = %d: %s", res.Code, res.Body.String())
		}
		var stats struct {
			Data communication.ReadStats `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &stats); err != nil {
			t.Fatalf("decode statistics: %v", err)
		}
		if stats.Data.ReadCount != 1 || stats.Data.TotalAudience < 1 {
			t.Fatalf("unexpected statistics: %+v", stats.Data)
		}

		res = request(http.MethodGet, "/api/v1/announcements/"+announcementID, otherWargaToken, "")
		if res.Code != http.StatusOK {
			t.Fatalf("same tenant target access = %d: %s", res.Code, res.Body.String())
		}
	})

	t.Run("agenda create, update, upcoming visibility, cancel", func(t *testing.T) {
		startsAt := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
		res := request(
			http.MethodPost,
			"/api/v1/events",
			sekretarisToken,
			fmt.Sprintf(`{"title":"Rapat RT","starts_at":%q,"location":"Balai RT","status":"planned","attachment_file_ids":[]}`, startsAt),
		)
		if res.Code != http.StatusCreated {
			t.Fatalf("create agenda = %d: %s", res.Code, res.Body.String())
		}
		var created struct {
			Data communication.EventItem `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &created); err != nil {
			t.Fatalf("decode agenda: %v", err)
		}

		res = request(
			http.MethodPatch,
			"/api/v1/events/"+created.Data.ID,
			sekretarisToken,
			fmt.Sprintf(`{"title":"Rapat RT Bulanan","starts_at":%q,"location":"Balai RT","status":"planned","attachment_file_ids":[]}`, startsAt),
		)
		if res.Code != http.StatusOK {
			t.Fatalf("update agenda = %d: %s", res.Code, res.Body.String())
		}

		res = request(http.MethodGet, "/api/v1/events?upcoming=true", wargaToken, "")
		if res.Code != http.StatusOK {
			t.Fatalf("list agenda = %d: %s", res.Code, res.Body.String())
		}
		var events struct {
			Data []communication.EventItem `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &events); err != nil {
			t.Fatalf("decode agenda list: %v", err)
		}
		found := false
		for _, item := range events.Data {
			if item.ID == created.Data.ID {
				found = true
				if item.Title != "Rapat RT Bulanan" {
					t.Fatalf("agenda title = %q", item.Title)
				}
			}
		}
		if !found {
			t.Fatal("upcoming agenda missing")
		}

		res = request(http.MethodPost, "/api/v1/events/"+created.Data.ID+"/cancel", sekretarisToken, "")
		if res.Code != http.StatusOK {
			t.Fatalf("cancel agenda = %d: %s", res.Code, res.Body.String())
		}
	})
}