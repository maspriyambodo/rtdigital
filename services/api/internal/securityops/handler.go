package securityops

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
	mux.Handle("GET /patrol-posts", h.require("security.patrol.manage", h.listPosts))
	mux.Handle("POST /patrol-posts", h.require("security.patrol.manage", h.createPost))
	mux.Handle("GET /patrol-schedules", h.require("security.patrol.manage", h.listSchedules))
	mux.Handle("POST /patrol-schedules", h.require("security.patrol.manage", h.createSchedule))
	mux.Handle("POST /patrol-assignments", h.require("security.patrol.manage", h.assignPatrol))
	mux.Handle("POST /patrol-assignments/{id}/swap", h.require("security.patrol.attend", h.swapPatrol))
	mux.Handle("POST /patrol-attendances", h.require("security.patrol.attend", h.checkInPatrol))
	mux.Handle("POST /patrol-incidents", h.require("security.incident.report", h.reportIncident))
	mux.Handle("GET /patrol-incidents", h.require("security.patrol.manage", h.listIncidents))

	mux.Handle("GET /community-activities", h.require("activity.read", h.listActivities))
	mux.Handle("POST /community-activities", h.require("activity.manage", h.createActivity))

	mux.Handle("POST /visitor-invites", h.require("visitor.invite", h.createVisitorInvite))
	mux.Handle("POST /visitor-logs", h.require("visitor.manage", h.checkInVisitor))
	mux.Handle("GET /visitor-logs", h.require("visitor.manage", h.listVisitorLogs))

	mux.Handle("POST /emergency-alerts", h.require("emergency.alert", h.createEmergencyAlert))
	mux.Handle("GET /emergency-alerts", h.require("emergency.manage", h.listEmergencyAlerts))
	mux.Handle("POST /emergency-alerts/{id}/acknowledge", h.require("emergency.manage", h.acknowledgeEmergencyAlert))
}

func (h *Handler) require(permission string, next http.HandlerFunc) http.Handler {
	return auth.RequireAuthenticatedPermission(h.tokens, h.authz, permission, false, next)
}

func (h *Handler) listPosts(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListPatrolPosts(r.Context(), auth.PrincipalFromContext(r.Context()))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createPost(w http.ResponseWriter, r *http.Request) {
	var req CreatePatrolPostRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.CreatePatrolPost(r.Context(), auth.PrincipalFromContext(r.Context()), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) listSchedules(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	items, err := h.service.ListPatrolSchedules(r.Context(), auth.PrincipalFromContext(r.Context()), q.Get("post_id"), q.Get("shift_date"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createSchedule(w http.ResponseWriter, r *http.Request) {
	var req CreatePatrolScheduleRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.CreatePatrolSchedule(r.Context(), auth.PrincipalFromContext(r.Context()), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) assignPatrol(w http.ResponseWriter, r *http.Request) {
	var req AssignPatrolRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.AssignPatrol(r.Context(), auth.PrincipalFromContext(r.Context()), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) swapPatrol(w http.ResponseWriter, r *http.Request) {
	var req SwapPatrolRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.SwapPatrol(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) checkInPatrol(w http.ResponseWriter, r *http.Request) {
	var req CheckInPatrolRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.CheckInPatrol(r.Context(), auth.PrincipalFromContext(r.Context()), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) reportIncident(w http.ResponseWriter, r *http.Request) {
	var req ReportIncidentRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.ReportIncident(r.Context(), auth.PrincipalFromContext(r.Context()), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) listIncidents(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListIncidents(r.Context(), auth.PrincipalFromContext(r.Context()))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) listActivities(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListActivities(r.Context(), auth.PrincipalFromContext(r.Context()))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createActivity(w http.ResponseWriter, r *http.Request) {
	var req CreateActivityRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.CreateActivity(r.Context(), auth.PrincipalFromContext(r.Context()), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) createVisitorInvite(w http.ResponseWriter, r *http.Request) {
	var req CreateVisitorInviteRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	p := auth.PrincipalFromContext(r.Context())
	item, err := h.service.CreateVisitorInvite(r.Context(), p, p.UserID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) checkInVisitor(w http.ResponseWriter, r *http.Request) {
	var req CheckInVisitorRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.CheckInVisitor(r.Context(), auth.PrincipalFromContext(r.Context()), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) listVisitorLogs(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListVisitorLogs(r.Context(), auth.PrincipalFromContext(r.Context()))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createEmergencyAlert(w http.ResponseWriter, r *http.Request) {
	var req CreateEmergencyAlertRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	p := auth.PrincipalFromContext(r.Context())
	item, err := h.service.CreateEmergencyAlert(r.Context(), p, p.UserID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) listEmergencyAlerts(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListEmergencyAlerts(r.Context(), auth.PrincipalFromContext(r.Context()), r.URL.Query().Get("status"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) acknowledgeEmergencyAlert(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.AcknowledgeEmergencyAlert(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
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
	case errors.Is(err, ErrPostNotFound):
		writeError(w, http.StatusNotFound, "POST_NOT_FOUND", "Pos ronda tidak ditemukan.")
	case errors.Is(err, ErrScheduleNotFound):
		writeError(w, http.StatusNotFound, "SCHEDULE_NOT_FOUND", "Jadwal ronda tidak ditemukan.")
	case errors.Is(err, ErrAssignmentNotFound):
		writeError(w, http.StatusNotFound, "ASSIGNMENT_NOT_FOUND", "Penugasan ronda tidak ditemukan.")
	case errors.Is(err, ErrActivityNotFound):
		writeError(w, http.StatusNotFound, "ACTIVITY_NOT_FOUND", "Kegiatan warga tidak ditemukan.")
	case errors.Is(err, ErrInviteNotFound):
		writeError(w, http.StatusNotFound, "INVITE_NOT_FOUND", "Undangan tamu tidak ditemukan.")
	case errors.Is(err, ErrAlertNotFound):
		writeError(w, http.StatusNotFound, "ALERT_NOT_FOUND", "Panggilan darurat tidak ditemukan.")
	case errors.Is(err, ErrDuplicateCode):
		writeError(w, http.StatusConflict, "DUPLICATE_CODE", "Kode atau data sudah terdaftar.")
	case errors.Is(err, ErrInvalidState):
		writeError(w, http.StatusConflict, "INVALID_STATE", "Status tidak mengizinkan operasi ini.")
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
