package users

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
	mux.Handle("GET /roles", auth.RequireAuthenticatedPermission(h.tokens, h.authz, "role.assign", true, http.HandlerFunc(h.listRoles)))
	mux.Handle("GET /users", auth.RequireAuthenticatedPermission(h.tokens, h.authz, "user.read", true, http.HandlerFunc(h.listUsers)))
	mux.Handle("GET /users/{id}", auth.RequireAuthenticatedPermission(h.tokens, h.authz, "user.read", true, http.HandlerFunc(h.getUser)))
	mux.Handle("PATCH /users/{id}", auth.RequireAuthenticatedPermission(h.tokens, h.authz, "user.update", true, http.HandlerFunc(h.updateUser)))
	mux.Handle("POST /users/invite", auth.RequireAuthenticatedPermission(h.tokens, h.authz, "user.invite", true, http.HandlerFunc(h.inviteUser)))
	mux.Handle("POST /users/{id}/deactivate", auth.RequireAuthenticatedPermission(h.tokens, h.authz, "user.deactivate", true, http.HandlerFunc(h.deactivateUser)))
	mux.Handle("POST /users/{id}/roles", auth.RequireAuthenticatedPermission(h.tokens, h.authz, "role.assign", true, http.HandlerFunc(h.assignRole)))
	mux.Handle("DELETE /users/{id}/roles/{role_id}", auth.RequireAuthenticatedPermission(h.tokens, h.authz, "role.revoke", true, http.HandlerFunc(h.revokeRole)))
}

func (h *Handler) listRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.service.ListRoles(r.Context(), auth.PrincipalFromContext(r.Context()))
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat daftar peran.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": roles})
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers(r.Context(), auth.PrincipalFromContext(r.Context()))
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat daftar pengguna.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": users})
}

func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.service.GetUser(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": user})
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email  *string `json:"email"`
		Phone  *string `json:"phone"`
		Status *string `json:"status"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	if err := h.service.UpdateUser(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request.Email, request.Phone, request.Status); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) inviteUser(w http.ResponseWriter, r *http.Request) {
	var request InviteUserRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	response, err := h.service.InviteUser(r.Context(), auth.PrincipalFromContext(r.Context()), request)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": response})
}

func (h *Handler) deactivateUser(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeactivateUser(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id")); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) assignRole(w http.ResponseWriter, r *http.Request) {
	var request struct {
		RoleID string `json:"role_id"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	if request.RoleID == "" {
		writeError(w, r, http.StatusBadRequest, "VALIDATION_FAILED", "Role ID wajib diisi.")
		return
	}
	if err := h.service.AssignRole(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request.RoleID); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) revokeRole(w http.ResponseWriter, r *http.Request) {
	if err := h.service.RevokeRole(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), r.PathValue("role_id")); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		writeError(w, r, http.StatusBadRequest, "VALIDATION_FAILED", "Data permintaan tidak valid.")
	case errors.Is(err, ErrUserNotFound):
		writeError(w, r, http.StatusNotFound, "USER_NOT_FOUND", "Pengguna tidak ditemukan.")
	case errors.Is(err, ErrRoleNotFound):
		writeError(w, r, http.StatusNotFound, "ROLE_NOT_FOUND", "Peran tidak ditemukan.")
	case errors.Is(err, ErrDuplicateContact):
		writeError(w, r, http.StatusConflict, "DUPLICATE_CONTACT", "Email atau nomor telepon sudah digunakan.")
	case errors.Is(err, ErrCannotEscalate):
		writeError(w, r, http.StatusForbidden, "ESCALATION_DENIED", "Tidak diizinkan memberikan peran ini.")
	case errors.Is(err, ErrMFAEnrollmentRequired):
		writeError(w, r, http.StatusForbidden, "MFA_ENROLLMENT_REQUIRED", "Akun target harus mengaktifkan MFA untuk peran ini.")
	case errors.Is(err, ErrCannotModifySelf):
		writeError(w, r, http.StatusForbidden, "SELF_MODIFICATION_DENIED", "Tidak dapat mengubah peran sendiri.")
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