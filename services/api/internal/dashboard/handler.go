package dashboard

import (
	"encoding/json"
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
	mux.Handle(
		"GET /dashboard/resident",
		auth.RequireAuthenticatedPermission(h.tokens, h.authz, "organization.read", false, http.HandlerFunc(h.getResident)),
	)
	mux.Handle(
		"GET /dashboard/admin",
		auth.RequireAuthenticatedPermission(h.tokens, h.authz, "organization.read", true, http.HandlerFunc(h.getAdmin)),
	)
}

func (h *Handler) getResident(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.GetResidentDashboard(r.Context(), auth.PrincipalFromContext(r.Context()))
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat dashboard warga.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func (h *Handler) getAdmin(w http.ResponseWriter, r *http.Request) {
	principal := auth.PrincipalFromContext(r.Context())
	if !canReadAdminDashboard(principal) {
		writeError(w, r, http.StatusForbidden, "PERMISSION_DENIED", "Akses dashboard pengurus tidak diizinkan.")
		return
	}

	data, err := h.service.GetAdminDashboard(r.Context(), principal)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat dashboard pengurus.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func canReadAdminDashboard(principal *auth.Principal) bool {
	if principal == nil {
		return false
	}
	for _, role := range principal.RoleCodes {
		switch role {
		case "ketua_rt", "sekretaris", "bendahara":
			return true
		}
	}
	return false
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]any{
		"code":       code,
		"message":    message,
		"details":    []any{},
		"request_id": r.Header.Get("X-Request-ID"),
	}})
}
