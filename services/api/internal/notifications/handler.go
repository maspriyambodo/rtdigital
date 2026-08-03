package notifications

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
	mux.Handle("GET /notifications", h.require(h.list))
	mux.Handle("PATCH /notifications/{id}/read", h.require(h.markAsRead))
	mux.Handle("POST /notifications/read-all", h.require(h.markAllAsRead))
}

func (h *Handler) require(next http.HandlerFunc) http.Handler {
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

		next.ServeHTTP(w, r.WithContext(auth.WithPrincipal(r.Context(), principal)))
	})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	filter := Filter{UnreadOnly: r.URL.Query().Get("unread_only") == "true"}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 {
			h.writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Parameter limit harus bernilai 1 sampai 100.")
			return
		}
		filter.Limit = limit
	}

	items, err := h.service.List(r.Context(), auth.PrincipalFromContext(r.Context()), filter)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) markAsRead(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.MarkAsRead(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) markAllAsRead(w http.ResponseWriter, r *http.Request) {
	count, err := h.service.MarkAllAsRead(r.Context(), auth.PrincipalFromContext(r.Context()))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]any{"data": map[string]int64{"updated_count": count}})
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		h.writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Data permintaan tidak valid.")
	case errors.Is(err, ErrNotificationNotFound):
		h.writeError(w, http.StatusNotFound, "NOTIFICATION_NOT_FOUND", "Notifikasi tidak ditemukan.")
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
	h.writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}