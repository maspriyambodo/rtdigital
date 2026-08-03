package settings

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

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
	mux.Handle("GET /organizations/current", h.require("organization.read", h.get))
	mux.Handle("PATCH /organizations/current", h.require("organization.update", h.update))
}

func (h *Handler) require(permission string, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix = "Bearer "
		token := r.Header.Get("Authorization")
		if !strings.HasPrefix(token, prefix) {
			h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Access token wajib diisi.")
			return
		}

		claims, err := h.tokens.VerifyAccessToken(strings.TrimPrefix(token, prefix))
		if err != nil {
			h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Access token tidak valid atau kedaluwarsa.")
			return
		}
		principal, err := h.authz.BuildPrincipal(r.Context(), claims)
		if err != nil {
			h.writeError(w, http.StatusUnauthorized, "SESSION_INVALID", "Sesi tidak valid atau telah berakhir.")
			return
		}
		if !principal.HasPermission(permission) {
			h.writeError(w, http.StatusForbidden, "FORBIDDEN", "Anda tidak memiliki izin untuk melakukan tindakan ini.")
			return
		}

		next.ServeHTTP(w, r.WithContext(auth.WithPrincipal(r.Context(), principal)))
	})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context(), auth.PrincipalFromContext(r.Context()).OrganizationID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var req UpdateOrganizationSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "MALFORMED_JSON", "Format JSON tidak valid.")
		return
	}

	item, err := h.service.Update(
		r.Context(),
		auth.PrincipalFromContext(r.Context()),
		req,
		r.Header.Get("X-Request-ID"),
	)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		h.writeError(w, http.StatusNotFound, "ORGANIZATION_NOT_FOUND", "Pengaturan organisasi tidak ditemukan.")
	case strings.Contains(err.Error(), "required"),
		strings.Contains(err.Error(), "invalid"),
		strings.Contains(err.Error(), "between"),
		strings.Contains(err.Error(), "valid JSON"):
		h.writeError(w, http.StatusUnprocessableEntity, "VALIDATION_FAILED", err.Error())
	default:
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan sistem.")
	}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code, message string) {
	h.writeJSON(w, status, map[string]any{"error": map[string]string{
		"code":       code,
		"message":    message,
	}})
}