package complaints

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
	mux.Handle("GET /complaints", h.require("complaint.read", h.listComplaints))
	mux.Handle("POST /complaints", h.require("complaint.submit", h.createComplaint))
	mux.Handle("GET /complaints/{id}", h.require("complaint.read", h.getComplaint))
	mux.Handle("PATCH /complaints/{id}", h.require("complaint.submit", h.updateComplaint))
	mux.Handle("POST /complaints/{id}/assign", h.require("complaint.assign", h.assignComplaint))
	mux.Handle("POST /complaints/{id}/status", h.require("complaint.read", h.updateStatus))
	mux.Handle("POST /complaints/{id}/comments", h.require("complaint.comment", h.addComment))
}

func (h *Handler) require(permission string, next http.HandlerFunc) http.Handler {
	return auth.RequireAuthenticatedPermission(h.tokens, h.authz, permission, false, next)
}

func (h *Handler) listComplaints(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	items, err := h.service.ListComplaints(r.Context(), auth.PrincipalFromContext(r.Context()), ComplaintFilter{
		Status:     query.Get("status"),
		Category:   query.Get("category"),
		AssignedTo: query.Get("assigned_to"),
		Search:     query.Get("search"),
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createComplaint(w http.ResponseWriter, r *http.Request) {
	var request CreateComplaintRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.CreateComplaint(r.Context(), auth.PrincipalFromContext(r.Context()), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) getComplaint(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetComplaint(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) updateComplaint(w http.ResponseWriter, r *http.Request) {
	var request UpdateComplaintRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.UpdateComplaint(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) assignComplaint(w http.ResponseWriter, r *http.Request) {
	var request AssignComplaintRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.AssignComplaint(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	var request UpdateStatusRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.UpdateStatus(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) addComment(w http.ResponseWriter, r *http.Request) {
	var request AddCommentRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.AddComment(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Data permintaan tidak valid.")
	case errors.Is(err, ErrComplaintNotFound):
		writeError(w, http.StatusNotFound, "COMPLAINT_NOT_FOUND", "Aduan tidak ditemukan.")
	case errors.Is(err, ErrInvalidState):
		writeError(w, http.StatusConflict, "INVALID_STATE", "Status data tidak mengizinkan operasi ini.")
	case errors.Is(err, ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Akses tidak diizinkan.")
	case errors.Is(err, ErrConflict):
		writeError(w, http.StatusConflict, "CONFLICT", "Data sudah ada.")
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
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
