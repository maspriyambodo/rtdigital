package securityops_test

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
	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
	"github.com/maspriyambodo/rtdigital/services/api/internal/securityops"
)

func TestSecurityOpsIntegration(t *testing.T) {
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

	secService := securityops.NewService(pool)
	server := httpapi.NewServer(
		slog.New(slog.NewTextHandler(os.Stderr, nil)),
		pool,
		tokens,
		auth.NewService(pool, tokens, crypter, auth.NoopMailer{}, "http://localhost:3000"),
		auth.NewAuthorizationService(pool),
		nil, nil, nil, nil, nil, nil,
		false,
		secService,
	)

	sequence := uint64(time.Now().UnixNano())
	nextID := func(segment string) string {
		sequence++
		return fmt.Sprintf("%08x-%s-4000-8000-%012x", sequence&0xffffffff, segment, sequence&0xffffffffffff)
	}

	orgID := nextID("9701")
	if _, err := pool.Exec(ctx, `
		INSERT INTO organizations (id, name, rt_number, rw_number, status)
		VALUES ($1, 'RT Test Security', '01', '01', 'active')`,
		orgID,
	); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	createUser := func(role string) (string, string) {
		t.Helper()
		userID, sessionID := nextID("1700"), nextID("2700")
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
	_, wargaToken := createUser("warga")

	request := func(method, path, token, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)
		return res
	}

	t.Run("patrol post and schedule creation", func(t *testing.T) {
		res := request(http.MethodPost, "/api/v1/patrol-posts", adminToken, `{"code":"POS-01","name":"Pos Utama Utama RT 01"}`)
		if res.Code != http.StatusCreated {
			t.Fatalf("create patrol post = %d: %s", res.Code, res.Body.String())
		}
		var postRes struct {
			Data securityops.PatrolPost `json:"data"`
		}
		_ = json.Unmarshal(res.Body.Bytes(), &postRes)

		schedBody := fmt.Sprintf(`{
			"post_id": "%s",
			"shift_date": "2026-08-10",
			"shift_start_time": "22:00",
			"shift_end_time": "04:00"
		}`, postRes.Data.ID)
		res = request(http.MethodPost, "/api/v1/patrol-schedules", adminToken, schedBody)
		if res.Code != http.StatusCreated {
			t.Fatalf("create patrol schedule = %d: %s", res.Code, res.Body.String())
		}

		res = request(http.MethodPost, "/api/v1/emergency-alerts", wargaToken, `{"category":"fire","location_details":"Rumah Blok A1 No 5"}`)
		if res.Code != http.StatusCreated {
			t.Fatalf("create emergency alert = %d: %s", res.Code, res.Body.String())
		}
	})
}
