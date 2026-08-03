package audit

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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
	mux.Handle("GET /audit-logs", h.require("audit.read", h.list))
	mux.Handle("GET /audit-logs/{id}", h.require("audit.read", h.get))
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

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	filter := Filter{
		Action:      r.URL.Query().Get("action"),
		ActorUserID: r.URL.Query().Get("actor_user_id"),
		EntityType:  r.URL.Query().Get("entity_type"),
		EntityID:    r.URL.Query().Get("entity_id"),
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 {
			h.writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Parameter limit harus bernilai 1 sampai 100.")
			return
		}
		filter.Limit = limit
	}
	if raw := r.URL.Query().Get("cursor"); raw != "" {
		cursor, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || cursor < 1 {
			h.writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Parameter cursor harus angka positif.")
			return
		}
		filter.Cursor = cursor
	}

	items, err := h.service.List(r.Context(), auth.PrincipalFromContext(r.Context()).OrganizationID, filter)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context(), auth.PrincipalFromContext(r.Context()).OrganizationID, r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		h.writeError(w, http.StatusNotFound, "AUDIT_LOG_NOT_FOUND", "Audit log tidak ditemukan.")
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