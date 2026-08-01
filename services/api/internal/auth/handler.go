package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"
)

type claimsContextKey struct{}

type Handler struct {
	service       *Service
	tokens        *TokenManager
	secureCookies bool
}

func NewHandler(service *Service, tokens *TokenManager, secureCookies bool) *Handler {
	return &Handler{service: service, tokens: tokens, secureCookies: secureCookies}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/login", h.login)
	mux.HandleFunc("POST /auth/refresh", h.refresh)
	mux.Handle("POST /auth/logout", h.requireAccess(http.HandlerFunc(h.logout), false))
	mux.Handle("POST /auth/logout-all", h.requireAccess(http.HandlerFunc(h.logoutAll), false))
	mux.HandleFunc("POST /auth/activate", h.activate)
	mux.HandleFunc("POST /auth/forgot-password", h.forgotPassword)
	mux.HandleFunc("POST /auth/reset-password", h.resetPassword)
	mux.Handle("POST /auth/mfa/verify", h.requireAccess(http.HandlerFunc(h.verifyMFA), false))
	mux.Handle("PATCH /me/password", h.requireAccess(http.HandlerFunc(h.changePassword), true))
	mux.Handle("POST /me/mfa/generate", h.requireAccess(http.HandlerFunc(h.generateMFA), true))
	mux.Handle("POST /me/mfa/enable", h.requireAccess(http.HandlerFunc(h.enableMFA), true))
	mux.Handle("POST /me/mfa/disable", h.requireAccess(http.HandlerFunc(h.disableMFA), true))
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if strings.TrimSpace(body.Identifier) == "" || body.Password == "" {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_FAILED", "Identitas dan kata sandi wajib diisi.")
		return
	}

	result, err := h.service.Login(r.Context(), body.Identifier, body.Password, r.UserAgent(), clientIP(r))
	if err != nil {
		h.writeAuthError(w, r, err)
		return
	}
	h.setRefreshCookie(w, result.RefreshToken, result.RefreshExpires)
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Sesi tidak ditemukan.")
		return
	}

	result, err := h.service.Refresh(r.Context(), cookie.Value, r.UserAgent(), clientIP(r))
	if err != nil {
		h.clearRefreshCookie(w)
		h.writeAuthError(w, r, err)
		return
	}
	h.setRefreshCookie(w, result.RefreshToken, result.RefreshExpires)
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Logout(r.Context(), ClaimsFromContext(r.Context()).SessionID); err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal keluar.")
		return
	}
	h.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) logoutAll(w http.ResponseWriter, r *http.Request) {
	if err := h.service.LogoutAll(r.Context(), ClaimsFromContext(r.Context()).UserID); err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal keluar dari semua perangkat.")
		return
	}
	h.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) activate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if strings.TrimSpace(body.Token) == "" {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_FAILED", "Token aktivasi wajib diisi.")
		return
	}
	if err := h.service.ActivateAccount(r.Context(), body.Token, body.Password); err != nil {
		h.writeAuthError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if err := h.service.RequestPasswordReset(r.Context(), body.Email); err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Permintaan reset gagal diproses.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if strings.TrimSpace(body.Token) == "" {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_FAILED", "Token reset wajib diisi.")
		return
	}
	if err := h.service.ResetPassword(r.Context(), body.Token, body.Password); err != nil {
		h.writeAuthError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) verifyMFA(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code string `json:"code"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	result, err := h.service.VerifyMFA(r.Context(), ClaimsFromContext(r.Context()), body.Code)
	if err != nil {
		h.writeAuthError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if err := h.service.ChangePassword(r.Context(), ClaimsFromContext(r.Context()).UserID, body.OldPassword, body.NewPassword); err != nil {
		h.writeAuthError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) generateMFA(w http.ResponseWriter, r *http.Request) {
	enrollment, err := h.service.GenerateMFASecret(r.Context(), ClaimsFromContext(r.Context()).UserID)
	if err != nil {
		h.writeAuthError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": enrollment})
}

func (h *Handler) enableMFA(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code string `json:"code"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if err := h.service.EnableMFA(r.Context(), ClaimsFromContext(r.Context()).UserID, body.Code); err != nil {
		h.writeAuthError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) disableMFA(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code string `json:"code"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if err := h.service.DisableMFA(r.Context(), ClaimsFromContext(r.Context()).UserID, body.Code); err != nil {
		h.writeAuthError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) requireAccess(next http.Handler, requireMFA bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix = "Bearer "
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, prefix) {
			writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Access token wajib diisi.")
			return
		}
		claims, err := h.tokens.VerifyAccessToken(strings.TrimPrefix(header, prefix))
		if err != nil {
			writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Access token tidak valid atau kedaluwarsa.")
			return
		}
		if requireMFA && !claims.MFA {
			writeError(w, r, http.StatusForbidden, "MFA_REQUIRED", "Verifikasi MFA diperlukan.")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsContextKey{}, claims)))
	})
}

func ClaimsFromContext(ctx context.Context) *TokenClaims {
	claims, _ := ctx.Value(claimsContextKey{}).(*TokenClaims)
	return claims
}

func (h *Handler) setRefreshCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name: "refresh_token", Value: token, Path: "/auth", Expires: expires,
		HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: "refresh_token", Value: "", Path: "/auth", MaxAge: -1,
		HttpOnly: true, Secure: h.secureCookies, SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) writeAuthError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrInvalidCredentials), errors.Is(err, ErrInvalidTOTPCode):
		writeError(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Identitas, kata sandi, atau kode MFA tidak valid.")
	case errors.Is(err, ErrAccountInactive):
		writeError(w, r, http.StatusForbidden, "ACCOUNT_INACTIVE", "Akun tidak aktif.")
	case errors.Is(err, ErrAccountInvited):
		writeError(w, r, http.StatusForbidden, "ACTIVATION_REQUIRED", "Akun belum diaktivasi.")
	case errors.Is(err, ErrAccountLocked):
		writeError(w, r, http.StatusTooManyRequests, "ACCOUNT_LOCKED", "Akun terkunci sementara.")
	case errors.Is(err, ErrSessionExpired), errors.Is(err, ErrInvalidToken):
		writeError(w, r, http.StatusUnauthorized, "SESSION_EXPIRED", "Sesi telah berakhir.")
	case errors.Is(err, ErrTokenNotFound):
		writeError(w, r, http.StatusNotFound, "TOKEN_NOT_FOUND", "Token tidak ditemukan atau kedaluwarsa.")
	case errors.Is(err, ErrWeakPassword):
		writeError(w, r, http.StatusBadRequest, "WEAK_PASSWORD", "Kata sandi minimal 8 karakter.")
	default:
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan internal.")
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Permintaan tidak valid.")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]any{
		"code": code, "message": message, "details": []any{}, "request_id": r.Header.Get("X-Request-ID"),
	}})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}