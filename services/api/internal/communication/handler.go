package communication

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
	mux.Handle("GET /announcements", h.require("announcement.read", h.listAnnouncements))
	mux.Handle("POST /announcements", h.require("announcement.create", h.createAnnouncement))
	mux.Handle("GET /announcements/{id}", h.require("announcement.read", h.getAnnouncement))
	mux.Handle("PATCH /announcements/{id}", h.require("announcement.update", h.updateAnnouncement))
	mux.Handle("POST /announcements/{id}/publish", h.require("announcement.update", h.publishAnnouncement))
	mux.Handle("POST /announcements/{id}/archive", h.require("announcement.archive", h.archiveAnnouncement))
	mux.Handle("GET /announcements/{id}/read-stats", h.require("announcement.create", h.getReadStats))

	mux.Handle("GET /events", h.require("event.read", h.listEvents))
	mux.Handle("POST /events", h.require("event.create", h.createEvent))
	mux.Handle("GET /events/{id}", h.require("event.read", h.getEvent))
	mux.Handle("PATCH /events/{id}", h.require("event.update", h.updateEvent))
	mux.Handle("POST /events/{id}/cancel", h.require("event.cancel", h.cancelEvent))
}

func (h *Handler) require(permission string, next http.HandlerFunc) http.Handler {
	return auth.RequireAuthenticatedPermission(h.tokens, h.authz, permission, false, next)
}

func (h *Handler) listAnnouncements(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	items, err := h.service.ListAnnouncements(r.Context(), auth.PrincipalFromContext(r.Context()), AnnouncementFilter{
		Status:   query.Get("status"),
		Category: query.Get("category"),
		Priority: query.Get("priority"),
		Search:   query.Get("search"),
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createAnnouncement(w http.ResponseWriter, r *http.Request) {
	var request CreateAnnouncementRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.CreateAnnouncement(r.Context(), auth.PrincipalFromContext(r.Context()), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) getAnnouncement(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetAnnouncement(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) updateAnnouncement(w http.ResponseWriter, r *http.Request) {
	var request UpdateAnnouncementRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.UpdateAnnouncement(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) publishAnnouncement(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.PublishAnnouncement(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) archiveAnnouncement(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.ArchiveAnnouncement(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) getReadStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetAnnouncementReadStats(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": stats})
}

func (h *Handler) listEvents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	items, err := h.service.ListEvents(r.Context(), auth.PrincipalFromContext(r.Context()), EventFilter{
		Status:   query.Get("status"),
		Upcoming: query.Get("upcoming") == "true",
		Search:   query.Get("search"),
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createEvent(w http.ResponseWriter, r *http.Request) {
	var request CreateEventRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.CreateEvent(r.Context(), auth.PrincipalFromContext(r.Context()), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) getEvent(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetEvent(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) updateEvent(w http.ResponseWriter, r *http.Request) {
	var request UpdateEventRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.UpdateEvent(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) cancelEvent(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.CancelEvent(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Data permintaan tidak valid.")
	case errors.Is(err, ErrAnnouncementNotFound):
		writeError(w, http.StatusNotFound, "ANNOUNCEMENT_NOT_FOUND", "Pengumuman tidak ditemukan.")
	case errors.Is(err, ErrEventNotFound):
		writeError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Agenda tidak ditemukan.")
	case errors.Is(err, ErrInvalidState):
		writeError(w, http.StatusConflict, "INVALID_STATE", "Status data tidak mengizinkan operasi ini.")
	case errors.Is(err, ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Akses tidak diizinkan.")
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
