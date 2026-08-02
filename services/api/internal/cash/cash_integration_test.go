package cash_test

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
	"github.com/maspriyambodo/rtdigital/services/api/internal/cash"
	"github.com/maspriyambodo/rtdigital/services/api/internal/httpapi"
	"github.com/maspriyambodo/rtdigital/services/api/internal/invoices"
	"github.com/maspriyambodo/rtdigital/services/api/internal/payments"
	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
	"github.com/maspriyambodo/rtdigital/services/api/internal/residents"
	"github.com/maspriyambodo/rtdigital/services/api/internal/users"
)

func TestCashBookIntegration(t *testing.T) {
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
	)

	sequence := uint64(time.Now().UnixNano())
	nextID := func(segment string) string {
		sequence++
		return fmt.Sprintf("%08x-%s-4000-8000-%012x", sequence&0xffffffff, segment, sequence&0xffffffffffff)
	}

	orgID := nextID("9001")
	if _, err := pool.Exec(ctx, `
		INSERT INTO organizations (id, name, rt_number, rw_number, status)
		VALUES ($1, 'RT Test Kas', '01', '01', 'active')`,
		orgID,
	); err != nil {
		t.Fatalf("create organization: %v", err)
	}

	createUser := func(role string) (string, string) {
		t.Helper()
		userID, sessionID := nextID("1000"), nextID("2000")
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
			t.Fatalf("issue token: %v", err)
		}
		return userID, token
	}

	wargaID, _ := createUser("warga")
	_, bendaharaToken := createUser("bendahara")

	request := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+bendaharaToken)
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)
		return res
	}

	createCategory := func(name, categoryType string) string {
		t.Helper()
		res := request(http.MethodPost, "/api/v1/cash/categories", fmt.Sprintf(`{"name":%q,"type":%q}`, name, categoryType))
		if res.Code != http.StatusCreated {
			t.Fatalf("create category = %d: %s", res.Code, res.Body.String())
		}
		var response struct {
			Data cash.Category `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode category: %v", err)
		}
		return response.Data.ID
	}

	t.Run("manual transaction, reversal, balance, and append-only delete", func(t *testing.T) {
		categoryID := createCategory("Operasional", "expense")
		res := request(
			http.MethodPost,
			"/api/v1/cash/transactions",
			fmt.Sprintf(`{"type":"expense","category_id":%q,"amount":25000,"transaction_date":"2026-08-02","description":"Beli kertas"}`, categoryID),
		)
		if res.Code != http.StatusCreated {
			t.Fatalf("record transaction = %d: %s", res.Code, res.Body.String())
		}

		var recorded struct {
			Data cash.Transaction `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &recorded); err != nil {
			t.Fatalf("decode transaction: %v", err)
		}

		res = request(http.MethodPost, "/api/v1/cash/transactions/"+recorded.Data.ID+"/reverse", `{"reason":"Salah nominal"}`)
		if res.Code != http.StatusCreated {
			t.Fatalf("reverse transaction = %d: %s", res.Code, res.Body.String())
		}

		res = request(http.MethodGet, "/api/v1/cash/book", "")
		if res.Code != http.StatusOK {
			t.Fatalf("get book = %d: %s", res.Code, res.Body.String())
		}
		var bookResponse struct {
			Data cash.CashBook `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &bookResponse); err != nil {
			t.Fatalf("decode book: %v", err)
		}
		book := bookResponse.Data
		if book.TotalExpense != 25000 || book.TotalIncome != 25000 || book.Balance != 0 {
			t.Fatalf("unexpected balance: %#v", book)
		}
		if len(book.Transactions) != 2 || book.Transactions[0].Status != "reversed" || book.Transactions[1].ReversalOfID == nil {
			t.Fatalf("unexpected reversal history: %#v", book.Transactions)
		}

		if _, err := pool.Exec(ctx, `DELETE FROM cash_transactions WHERE id = $1`, recorded.Data.ID); err == nil {
			t.Fatal("cash transaction delete succeeded; expected append-only trigger rejection")
		}
	})

	t.Run("verified payment creates exactly one cash income", func(t *testing.T) {
		unitID, residentID, householdID, dueTypeID := nextID("3000"), nextID("4000"), nextID("5000"), nextID("6000")
		if _, err := pool.Exec(ctx, `
			INSERT INTO house_units (id, organization_id, code, occupancy_status, status)
			VALUES ($1, $2, 'KAS-01', 'owned', 'active')`,
			unitID, orgID,
		); err != nil {
			t.Fatalf("create unit: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO residents (id, organization_id, full_name, resident_status, verification_status)
			VALUES ($1, $2, 'Warga Kas', 'active', 'verified')`,
			residentID, orgID,
		); err != nil {
			t.Fatalf("create resident: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO households (id, organization_id, house_unit_id, internal_number, domicile_status, verification_status)
			VALUES ($1, $2, $3, 'KK-KAS-01', 'permanent', 'verified')`,
			householdID, orgID, unitID,
		); err != nil {
			t.Fatalf("create household: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO due_types (id, organization_id, name, amount, frequency, status)
			VALUES ($1, $2, 'Iuran Kas', 50000, 'monthly', 'active')`,
			dueTypeID, orgID,
		); err != nil {
			t.Fatalf("create due type: %v", err)
		}

		invoiceID, paymentID := nextID("8000"), nextID("9500")
		if _, err := pool.Exec(ctx, `
			INSERT INTO invoices (
				id, organization_id, household_id, due_type_id, invoice_number,
				period_start, period_end, due_date, amount, status
			) VALUES ($1, $2, $3, $4, $5, '2026-08-01', '2026-08-31', '2026-08-10', 50000, 'unpaid')`,
			invoiceID, orgID, householdID, dueTypeID, "INV-"+invoiceID[:8],
		); err != nil {
			t.Fatalf("create invoice: %v", err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO payments (
				id, organization_id, invoice_id, payment_number, method, amount, paid_at,
				verification_status, created_by
			) VALUES ($1, $2, $3, $4, 'cash', 50000, '2026-08-02T12:00:00Z', 'pending', $5)`,
			paymentID, orgID, invoiceID, "PAY-"+paymentID[:8], wargaID,
		); err != nil {
			t.Fatalf("create payment: %v", err)
		}

		res := request(http.MethodPost, "/api/v1/payments/"+paymentID+"/verify", `{}`)
		if res.Code != http.StatusOK {
			t.Fatalf("verify payment = %d: %s", res.Code, res.Body.String())
		}

		var count int
		var amount float64
		if err := pool.QueryRow(ctx, `
			SELECT COUNT(*), COALESCE(SUM(amount), 0)
			FROM cash_transactions
			WHERE organization_id = $1
			  AND reference_type = 'payment'
			  AND reference_id = $2`,
			orgID, paymentID,
		).Scan(&count, &amount); err != nil {
			t.Fatalf("query payment cash transaction: %v", err)
		}
		if count != 1 || amount != 50000 {
			t.Fatalf("payment cash relation = count:%d amount:%v", count, amount)
		}
	})
}
