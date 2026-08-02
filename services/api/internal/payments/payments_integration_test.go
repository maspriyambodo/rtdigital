package payments_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
	"github.com/maspriyambodo/rtdigital/services/api/internal/files"
	"github.com/maspriyambodo/rtdigital/services/api/internal/httpapi"
	"github.com/maspriyambodo/rtdigital/services/api/internal/invoices"
	"github.com/maspriyambodo/rtdigital/services/api/internal/payments"
	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
	"github.com/maspriyambodo/rtdigital/services/api/internal/residents"
	"github.com/maspriyambodo/rtdigital/services/api/internal/users"
)

type memoryStorage struct {
	objects map[string]platform.ObjectMetadata
}

func (m *memoryStorage) PresignUpload(_ context.Context, key, contentType string, _ time.Duration) (platform.PresignedURL, error) {
	m.objects[key] = platform.ObjectMetadata{SizeBytes: 1234, ContentType: contentType}
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

func TestPaymentsIntegration(t *testing.T) {
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
		payments.NewService(pool),
		false,
	)

	var sequence uint64 = uint64(time.Now().UnixNano())
	nextID := func(segment string) string {
		sequence++
		head := sequence & 0xffffffff
		tail := sequence & 0xffffffffffff
		return fmt.Sprintf("%08x-%s-4000-8000-%012x", head, segment, tail)
	}

	orgID := nextID("9001")
	otherOrgID := nextID("9002")
	if _, err := pool.Exec(ctx, `
		INSERT INTO organizations (id, name, rt_number, rw_number, status)
		VALUES ($1, 'RT Test Pembayaran', '01', '01', 'active'),
		       ($2, 'RT Tenant Lain', '02', '01', 'active')`,
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
			sessionID, userID, organizationID, fmt.Sprintf("refresh-hash-%s", sessionID),
		); err != nil {
			t.Fatalf("create session: %v", err)
		}
		token, err := tokens.IssueAccessToken(auth.TokenClaims{
			UserID: userID, OrganizationID: organizationID, SessionID: sessionID, MFA: true,
		})
		if err != nil {
			t.Fatalf("issue access token: %v", err)
		}
		return userID, token
	}

	wargaID, wargaToken := createUser("warga", orgID)
	bendaharaID, bendaharaToken := createUser("bendahara", orgID)
	_, otherBendaharaToken := createUser("bendahara", otherOrgID)

	unitID := nextID("3000")
	residentID := nextID("4000")
	householdID := nextID("5000")
	dueTypeID := nextID("6000")
	fixtureTx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin payment fixture: %v", err)
	}
	defer func() { _ = fixtureTx.Rollback(ctx) }()

	memberID := nextID("7000")
	if _, err := fixtureTx.Exec(ctx, `
		INSERT INTO house_units (id, organization_id, code, occupancy_status, status)
		VALUES ($1, $2, 'PAY-01', 'owned', 'active')`,
		unitID, orgID,
	); err != nil {
		t.Fatalf("create house unit fixture: %v", err)
	}
	if _, err := fixtureTx.Exec(ctx, `
		INSERT INTO residents (id, organization_id, full_name, resident_status, verification_status)
		VALUES ($1, $2, 'Warga Pembayar', 'active', 'verified')`,
		residentID, orgID,
	); err != nil {
		t.Fatalf("create resident fixture: %v", err)
	}
	if _, err := fixtureTx.Exec(ctx, `
		INSERT INTO households (id, organization_id, house_unit_id, internal_number, domicile_status, verification_status)
		VALUES ($1, $2, $3, 'KK-PAY-01', 'permanent', 'verified')`,
		householdID, orgID, unitID,
	); err != nil {
		t.Fatalf("create household fixture: %v", err)
	}
	if _, err := fixtureTx.Exec(ctx, `
		INSERT INTO household_members (id, organization_id, household_id, resident_id, relationship, is_active)
		VALUES ($1, $2, $3, $4, 'head', true)`,
		memberID, orgID, householdID, residentID,
	); err != nil {
		t.Fatalf("create household member fixture: %v", err)
	}
	if _, err := fixtureTx.Exec(ctx, `
		UPDATE households SET head_resident_id = $1 WHERE id = $2`,
		residentID, householdID,
	); err != nil {
		t.Fatalf("set household head fixture: %v", err)
	}
	if _, err := fixtureTx.Exec(ctx, `
		UPDATE users SET resident_id = $1 WHERE id = $2`,
		residentID, wargaID,
	); err != nil {
		t.Fatalf("link resident fixture: %v", err)
	}
	if _, err := fixtureTx.Exec(ctx, `
		INSERT INTO due_types (id, organization_id, name, amount, frequency, status)
		VALUES ($1, $2, 'Iuran Keamanan', 100000, 'monthly', 'active')`,
		dueTypeID, orgID,
	); err != nil {
		t.Fatalf("create due type fixture: %v", err)
	}
	if err := fixtureTx.Commit(ctx); err != nil {
		t.Fatalf("commit payment fixture: %v", err)
	}

	createInvoice := func(amount float64) string {
		t.Helper()
		invoiceID := nextID("8000")
		if _, err := pool.Exec(ctx, `
			INSERT INTO invoices (
				id, organization_id, household_id, due_type_id, invoice_number,
				period_start, period_end, due_date, amount, status
			) VALUES (
				$1, $2, $3, $4, $5,
				'2026-08-01', '2026-08-31', '2026-08-10', $6, 'unpaid'
			)`,
			invoiceID, orgID, householdID, dueTypeID, "INV-"+invoiceID[:8], amount,
		); err != nil {
			t.Fatalf("create invoice: %v", err)
		}
		return invoiceID
	}

	request := func(method, path, token, body string, headers map[string]string) *httptest.ResponseRecorder {
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

	invoiceID := createInvoice(100000)
	var proofFileID string
	t.Run("presign, confirm, submit idempotently", func(t *testing.T) {
		res := request(
			http.MethodPost,
			"/api/v1/files/presign-upload",
			wargaToken,
			fmt.Sprintf(`{"entity_type":"payment","entity_id":%q,"purpose":"payment_proof","original_name":"bukti.jpg","mime_type":"image/jpeg","size_bytes":1234}`, invoiceID),
			nil,
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
		proofFileID = upload.Data.FileID

		res = request(http.MethodPost, "/api/v1/files/confirm-upload", wargaToken, fmt.Sprintf(`{"file_id":%q}`, proofFileID), nil)
		if res.Code != http.StatusOK {
			t.Fatalf("confirm = %d: %s", res.Code, res.Body.String())
		}

		body := fmt.Sprintf(`{"invoice_id":%q,"method":"transfer","amount":100000,"paid_at":"2026-08-01T10:00:00Z","proof_file_id":%q}`, invoiceID, proofFileID)
		headers := map[string]string{"Idempotency-Key": "payment-idempotency-test"}
		first := request(http.MethodPost, "/api/v1/payments", wargaToken, body, headers)
		second := request(http.MethodPost, "/api/v1/payments", wargaToken, body, headers)
		if first.Code != http.StatusCreated || second.Code != http.StatusCreated {
			t.Fatalf("idempotent submit = %d/%d: %s", first.Code, second.Code, second.Body.String())
		}

		var firstData, secondData struct {
			Data struct {
				ID            string `json:"id"`
				InvoiceStatus string `json:"invoice_status"`
			} `json:"data"`
		}
		if err := json.Unmarshal(first.Body.Bytes(), &firstData); err != nil {
			t.Fatalf("decode first submit: %v", err)
		}
		if err := json.Unmarshal(second.Body.Bytes(), &secondData); err != nil {
			t.Fatalf("decode second submit: %v", err)
		}
		if firstData.Data.ID == "" || firstData.Data.ID != secondData.Data.ID || firstData.Data.InvoiceStatus != "pending_verification" {
			t.Fatalf("unexpected idempotency response: %#v / %#v", firstData, secondData)
		}
	})

	t.Run("signed download and tenant isolation", func(t *testing.T) {
		if res := request(http.MethodGet, "/api/v1/files/"+proofFileID+"/download", wargaToken, "", nil); res.Code != http.StatusOK {
			t.Fatalf("owner download = %d: %s", res.Code, res.Body.String())
		}
		if res := request(http.MethodGet, "/api/v1/payments", otherBendaharaToken, "", nil); res.Code != http.StatusOK {
			t.Fatalf("other tenant list = %d: %s", res.Code, res.Body.String())
		}
		if res := request(http.MethodPost, "/api/v1/payments/"+proofFileID+"/verify", otherBendaharaToken, `{}`, nil); res.Code != http.StatusNotFound {
			t.Fatalf("cross-tenant verify = %d, want 404", res.Code)
		}
	})

	t.Run("verify, audit, and cancellation reverse invoice", func(t *testing.T) {
		var paymentID string
		if err := pool.QueryRow(ctx, `
			SELECT id FROM payments
			WHERE organization_id = $1 AND idempotency_key = 'payment-idempotency-test'`,
			orgID,
		).Scan(&paymentID); err != nil {
			t.Fatalf("load payment: %v", err)
		}

		if _, err := pool.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id)
			SELECT $1, id FROM roles
			WHERE code = 'bendahara' AND organization_id IS NULL`,
			wargaID,
		); err != nil {
			t.Fatalf("grant test verifier role: %v", err)
		}
		if res := request(http.MethodPost, "/api/v1/payments/"+paymentID+"/verify", wargaToken, `{}`, nil); res.Code != http.StatusForbidden {
			t.Fatalf("self verify = %d, want %d: %s", res.Code, http.StatusForbidden, res.Body.String())
		}

		res := request(http.MethodPost, "/api/v1/payments/"+paymentID+"/verify", bendaharaToken, `{}`, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("verify = %d: %s", res.Code, res.Body.String())
		}

		var status string
		var paidAmount float64
		if err := pool.QueryRow(ctx, `SELECT status, paid_amount FROM invoices WHERE id = $1`, invoiceID).Scan(&status, &paidAmount); err != nil {
			t.Fatalf("load verified invoice: %v", err)
		}
		if status != "paid" || paidAmount != 100000 {
			t.Fatalf("verified invoice = %q / %v", status, paidAmount)
		}

		var audited bool
		if err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM audit_logs
				WHERE organization_id = $1
				  AND actor_user_id = $2
				  AND action = 'payment.verify'
				  AND entity_id = $3
			)`,
			orgID, bendaharaID, paymentID,
		).Scan(&audited); err != nil || !audited {
			t.Fatalf("payment verification audit missing: %v", err)
		}

		res = request(http.MethodPost, "/api/v1/payments/"+paymentID+"/cancel", bendaharaToken, `{"reason":"Salah mutasi"}`, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("cancel = %d: %s", res.Code, res.Body.String())
		}
		if err := pool.QueryRow(ctx, `SELECT status, paid_amount FROM invoices WHERE id = $1`, invoiceID).Scan(&status, &paidAmount); err != nil {
			t.Fatalf("load reversed invoice: %v", err)
		}
		if status != "unpaid" || paidAmount != 0 {
			t.Fatalf("reversed invoice = %q / %v", status, paidAmount)
		}
	})

	t.Run("parallel submissions cannot exceed invoice balance", func(t *testing.T) {
		parallelInvoiceID := createInvoice(100000)
		const workers = 5
		results := make(chan int, workers)
		var group sync.WaitGroup

		for worker := 0; worker < workers; worker++ {
			group.Add(1)
			go func(index int) {
				defer group.Done()
				res := request(
					http.MethodPost,
					"/api/v1/payments",
					wargaToken,
					fmt.Sprintf(`{"invoice_id":%q,"method":"cash","amount":100000,"paid_at":"2026-08-01T10:00:00Z"}`, parallelInvoiceID),
					map[string]string{"Idempotency-Key": fmt.Sprintf("parallel-payment-%d", index)},
				)
				results <- res.Code
			}(worker)
		}
		group.Wait()
		close(results)

		successes := 0
		for code := range results {
			if code == http.StatusCreated {
				successes++
				continue
			}
			if code != http.StatusConflict {
				t.Fatalf("parallel submit = %d, want 201 or 409", code)
			}
		}
		if successes != 1 {
			t.Fatalf("parallel successful submissions = %d, want 1", successes)
		}
	})

	t.Run("transfer requires proof", func(t *testing.T) {
		res := request(
			http.MethodPost,
			"/api/v1/payments",
			wargaToken,
			fmt.Sprintf(`{"invoice_id":%q,"method":"transfer","amount":1,"paid_at":"2026-08-01T10:00:00Z"}`, createInvoice(100000)),
			map[string]string{"Idempotency-Key": "missing-proof"},
		)
		if res.Code != http.StatusBadRequest {
			t.Fatalf("transfer without proof = %d, want 400", res.Code)
		}
	})
}
