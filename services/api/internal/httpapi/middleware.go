package httpapi

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const maxRequestBodyBytes = 50 << 20

type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*rateBucket
	limit   int
	window  time.Duration
}

type rateBucket struct {
	count    int
	windowAt time.Time
	lastSeen time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		buckets: make(map[string]*rateBucket),
		limit:   limit,
		window:  window,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket := rl.buckets[key]
	if bucket == nil || now.Sub(bucket.windowAt) >= rl.window {
		rl.buckets[key] = &rateBucket{count: 1, windowAt: now, lastSeen: now}
		return true
	}
	bucket.lastSeen = now
	if bucket.count >= rl.limit {
		return false
	}
	bucket.count++
	return true
}

func (rl *rateLimiter) prune() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-2 * rl.window)
	for key, bucket := range rl.buckets {
		if bucket.lastSeen.Before(cutoff) {
			delete(rl.buckets, key)
		}
	}
}

func withSecurityHeaders(production bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		if production {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

func withCORS(allowedOrigins map[string]struct{}, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		_, allowed := allowedOrigins[origin]
		if origin != "" && allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID, Idempotency-Key, X-CSRF-Token")
			w.Header().Set("Access-Control-Max-Age", "600")
		}

		if r.Method == http.MethodOptions {
			if origin != "" && !allowed {
				writeMiddlewareError(w, http.StatusForbidden, "CORS_ORIGIN_DENIED", "Origin tidak diizinkan.")
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withCSRF(allowedOrigins map[string]struct{}, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !unsafeMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}
		if _, err := r.Cookie("refresh_token"); err != nil {
			next.ServeHTTP(w, r)
			return
		}

		origin := r.Header.Get("Origin")
		if origin == "" {
			writeMiddlewareError(w, http.StatusForbidden, "CSRF_DETECTED", "Origin wajib untuk permintaan berbasis cookie.")
			return
		}
		if _, allowed := allowedOrigins[origin]; !allowed {
			writeMiddlewareError(w, http.StatusForbidden, "CSRF_DETECTED", "Origin tidak diizinkan.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withBodyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
		next.ServeHTTP(w, r)
	})
}

func withRateLimit(limiter *rateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			next.ServeHTTP(w, r)
			return
		}
		if !limiter.allow(clientAddress(r)) {
			w.Header().Set("Retry-After", "60")
			writeMiddlewareError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Terlalu banyak permintaan. Coba lagi sesaat.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withSanitizedLogging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.InfoContext(r.Context(), "request completed",
			"method", r.Method,
			"path", sanitizedPath(r.URL),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", RequestID(r.Context()),
		)
	})
}

func allowedOrigins() map[string]struct{} {
	raw := os.Getenv("APP_URL")
	if raw == "" {
		raw = "http://localhost:3000,http://127.0.0.1:3000"
	}

	origins := make(map[string]struct{})
	for _, value := range strings.Split(raw, ",") {
		value = strings.TrimRight(strings.TrimSpace(value), "/")
		if parsed, err := url.Parse(value); err == nil && parsed.Scheme != "" && parsed.Host != "" {
			origins[value] = struct{}{}
		}
	}
	return origins
}

func unsafeMethod(method string) bool {
	return method == http.MethodPost || method == http.MethodPut ||
		method == http.MethodPatch || method == http.MethodDelete
}

func clientAddress(r *http.Request) string {
	if ip := net.ParseIP(strings.TrimSpace(r.Header.Get("CF-Connecting-IP"))); ip != nil {
		return ip.String()
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func sanitizedPath(requestURL *url.URL) string {
	query := requestURL.Query()
	for key := range query {
		if sensitiveParameter(key) {
			query.Set(key, "[REDACTED]")
		}
	}
	if encoded := query.Encode(); encoded != "" {
		return requestURL.Path + "?" + encoded
	}
	return requestURL.Path
}

func sensitiveParameter(key string) bool {
	key = strings.ToLower(key)
	for _, marker := range []string{"token", "secret", "password", "code", "key", "authorization"} {
		if strings.Contains(key, marker) {
			return true
		}
	}
	return false
}

func writeMiddlewareError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code": code, "message": message, "details": []any{},
		},
	})
}