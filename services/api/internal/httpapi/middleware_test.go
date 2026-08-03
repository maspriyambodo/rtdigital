package httpapi

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestSecurityHeaders(t *testing.T) {
	handler := withSecurityHeaders(true, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	for header, want := range map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Permissions-Policy":        "camera=(), microphone=(), geolocation=()",
		"Content-Security-Policy":   "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
	} {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("%s = %q; want %q", header, got, want)
		}
	}
}

func TestCORSAndCSRF(t *testing.T) {
	origins := map[string]struct{}{"https://allowed.test": {}}
	handler := withCORS(origins, withCSRF(origins, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	t.Run("allowed preflight", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", "https://allowed.test")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d; want 204", rec.Code)
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://allowed.test" {
			t.Errorf("Access-Control-Allow-Origin = %q", got)
		}
	})

	t.Run("denied preflight", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", "https://denied.test")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d; want 403", rec.Code)
		}
	})

	t.Run("cookie mutation requires allowed origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "token"})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("missing origin status = %d; want 403", rec.Code)
		}

		req.Header.Set("Origin", "https://denied.test")
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("denied origin status = %d; want 403", rec.Code)
		}

		req.Header.Set("Origin", "https://allowed.test")
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("allowed origin status = %d; want 200", rec.Code)
		}
	})
}

func TestRateLimit(t *testing.T) {
	limiter := newRateLimiter(2, time.Minute)
	handler := withRateLimit(limiter, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := range 3 {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
		req.Header.Set("CF-Connecting-IP", "192.0.2.1")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if i < 2 && rec.Code != http.StatusOK {
			t.Errorf("request %d status = %d; want 200", i+1, rec.Code)
		}
		if i == 2 && rec.Code != http.StatusTooManyRequests {
			t.Errorf("request %d status = %d; want 429", i+1, rec.Code)
		}
	}
}

func TestSanitizedPath(t *testing.T) {
	requestURL, err := url.Parse("/api/v1/auth/login?email=test%40example.test&password=secret&token=123&safe=yes")
	if err != nil {
		t.Fatal(err)
	}

	got := sanitizedPath(requestURL)
	if strings.Contains(got, "secret") || strings.Contains(got, "123") {
		t.Errorf("sensitive values leaked: %s", got)
	}
	if !strings.Contains(got, "email=test%40example.test") || !strings.Contains(got, "safe=yes") {
		t.Errorf("safe values missing: %s", got)
	}
	if !strings.Contains(got, "%5BREDACTED%5D") {
		t.Errorf("redaction marker missing: %s", got)
	}
}