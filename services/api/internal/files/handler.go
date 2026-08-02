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
	mux.Handle("POST /files/presign-upload", h.require("payment.submit", h.presignUpload))
	mux.Handle("POST /files/confirm-upload", h.require("payment.submit", h.confirmUpload))
	mux.Handle("GET /files/{id}/download", h.require("payment.read", h.presignDownload))
}

func (h *Handler) require(permission string, next http.HandlerFunc) http.Handler {
	return auth.RequireAuthenticatedPermission(h.tokens, h.authz, permission, false, next)
}

func (h *Handler) presignUpload(w http.ResponseWriter, r *http.Request) {
	var request PresignUploadRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	response, err := h.service.PresignUpload(r.Context(), auth.PrincipalFromContext(r.Context()), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": response})
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
