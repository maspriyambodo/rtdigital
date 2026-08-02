package files

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

type Handler struct {
	service *Service
	tokens  *auth.TokenManager
	authz   *auth.AuthorizationService
}

func NewHandler(service *Service, tokens *auth.TokenManager, authz *auth.AuthorizationService) *Handler {
	return &Handler{service: service, tokens: tokens, authz: authz}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /files/presign-upload", h.requireAuthenticated(h.presignUpload))
	mux.Handle("POST /files/confirm-upload", h.requireAuthenticated(h.confirmUpload))
	mux.Handle("GET /files/{id}/download", h.requireAuthenticated(h.presignDownload))
}

func (h *Handler) requireAuthenticated(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix = "Bearer "
		header := r.Header.Get("Authorization")
		if len(header) < len(prefix) || header[:len(prefix)] != prefix {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Access token wajib diisi.")
			return
		}
		claims, err := h.tokens.VerifyAccessToken(header[len(prefix):])
		if err != nil {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Access token tidak valid atau kedaluwarsa.")
			return
		}
		principal, err := h.authz.BuildPrincipal(r.Context(), claims)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "SESSION_INVALID", "Sesi tidak valid atau telah berakhir.")
			return
		}
		next.ServeHTTP(w, r.WithContext(auth.WithPrincipal(r.Context(), principal)))
	})
}

func (h *Handler) presignUpload(w http.ResponseWriter, r *http.Request) {
	var request PresignUploadRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	principal := auth.PrincipalFromContext(r.Context())
	if !canUploadFor(principal, request.EntityType) {
		writeError(w, http.StatusForbidden, "PERMISSION_DENIED", "Akses tidak diizinkan.")
		return
	}

	response, err := h.service.PresignUpload(r.Context(), principal, request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": response})
}

func canUploadFor(principal *auth.Principal, entityType string) bool {
	if principal == nil {
		return false
	}
	switch entityType {
	case "payment":
		return principal.HasPermission("payment.submit")
	case "cash_transaction":
		return principal.HasPermission("cash.create")
	case "announcement":
		return principal.HasPermission("announcement.create")
	case "event":
		return principal.HasPermission("event.create")
	default:
		return false
	}
}

func (h *Handler) confirmUpload(w http.ResponseWriter, r *http.Request) {
	var request ConfirmUploadRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	response, err := h.service.ConfirmUpload(r.Context(), auth.PrincipalFromContext(r.Context()), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": response})
}

func (h *Handler) presignDownload(w http.ResponseWriter, r *http.Request) {
	response, err := h.service.PresignDownload(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": response})
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Data permintaan tidak valid.")
	case errors.Is(err, ErrFileNotFound):
		writeError(w, http.StatusNotFound, "FILE_NOT_FOUND", "File tidak ditemukan atau akses ditolak.")
	case errors.Is(err, ErrFileUnavailable):
		writeError(w, http.StatusConflict, "FILE_UNAVAILABLE", "File belum tersedia atau metadata tidak sesuai.")
	case errors.Is(err, ErrStorage):
		writeError(w, http.StatusServiceUnavailable, "STORAGE_ERROR", "Layanan penyimpanan tidak tersedia.")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan sistem.")
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Format JSON tidak valid.")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{
		"code":    code,
		"message": message,
	}})
}
