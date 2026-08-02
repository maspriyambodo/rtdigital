package residents

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
	mux.Handle("GET /house-units", h.require("house_unit.read", h.listHouseUnits))
	mux.Handle("POST /house-units", h.require("house_unit.create", h.createHouseUnit))
	mux.Handle("GET /house-units/{id}", h.require("house_unit.read", h.getHouseUnit))
	mux.Handle("PATCH /house-units/{id}", h.require("house_unit.update", h.updateHouseUnit))
	mux.Handle("POST /house-units/{id}/deactivate", h.require("house_unit.deactivate", h.deactivateHouseUnit))

	mux.Handle("GET /households", h.require("household.read", h.listHouseholds))
	mux.Handle("POST /households", h.require("household.create", h.createHousehold))
	mux.Handle("GET /households/{id}", h.require("household.read", h.getHousehold))
	mux.Handle("POST /households/{id}/members", h.require("household.update", h.addHouseholdMember))

	mux.Handle("GET /residents", h.require("resident.read", h.listResidents))
	mux.Handle("POST /residents", h.require("resident.create", h.createResident))
	mux.Handle("POST /residents/import", h.require("resident.create", h.importResidents))
	mux.Handle("GET /residents/{id}", h.require("resident.read", h.getResident))
	mux.Handle("POST /residents/{id}/verify", h.require("resident.verify", h.verifyResident))
	mux.Handle("POST /residents/{id}/corrections", h.require("resident.correction.submit", h.submitCorrection))
	mux.Handle("GET /resident-corrections", h.require("resident.correction.review", h.listCorrections))
	mux.Handle("POST /resident-corrections/{id}/{action}", h.require("resident.correction.review", h.reviewCorrection))
}

func (h *Handler) require(permission string, next http.HandlerFunc) http.Handler {
	return auth.RequireAuthenticatedPermission(h.tokens, h.authz, permission, false, next)
}

func (h *Handler) listHouseUnits(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListHouseUnits(r.Context(), auth.PrincipalFromContext(r.Context()))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createHouseUnit(w http.ResponseWriter, r *http.Request) {
	var request CreateHouseUnitRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.CreateHouseUnit(r.Context(), auth.PrincipalFromContext(r.Context()), request)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) getHouseUnit(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetHouseUnit(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) updateHouseUnit(w http.ResponseWriter, r *http.Request) {
	var request UpdateHouseUnitRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.UpdateHouseUnit(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) deactivateHouseUnit(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeactivateHouseUnit(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id")); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listHouseholds(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListHouseholds(r.Context(), auth.PrincipalFromContext(r.Context()))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createHousehold(w http.ResponseWriter, r *http.Request) {
	var request CreateHouseholdRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.CreateHousehold(r.Context(), auth.PrincipalFromContext(r.Context()), request)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) getHousehold(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetHousehold(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) addHouseholdMember(w http.ResponseWriter, r *http.Request) {
	var request HouseholdMemberRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	if err := h.service.AddHouseholdMember(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listResidents(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListResidents(r.Context(), auth.PrincipalFromContext(r.Context()), r.URL.Query().Get("q"), r.URL.Query().Get("status"))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createResident(w http.ResponseWriter, r *http.Request) {
	var request CreateResidentRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.CreateResident(r.Context(), auth.PrincipalFromContext(r.Context()), request)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) getResident(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetResident(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), r.URL.Query().Get("reason"))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) verifyResident(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.VerifyResident(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) importResidents(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/csv" {
		writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Content-Type harus text/csv.")
		return
	}
	result, err := h.service.ImportCSV(
		r.Context(),
		auth.PrincipalFromContext(r.Context()),
		http.MaxBytesReader(w, r.Body, 10<<20),
		r.URL.Query().Get("dry_run") == "true",
	)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func (h *Handler) submitCorrection(w http.ResponseWriter, r *http.Request) {
	var request CreateResidentCorrectionRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.SubmitCorrection(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) listCorrections(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListCorrections(r.Context(), auth.PrincipalFromContext(r.Context()), r.URL.Query().Get("status"))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) reviewCorrection(w http.ResponseWriter, r *http.Request) {
	var request ReviewResidentCorrectionRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	if err := h.service.ReviewCorrection(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), r.PathValue("action"), request); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Data permintaan tidak valid.")
	case errors.Is(err, ErrHouseUnitNotFound):
		writeError(w, http.StatusNotFound, "HOUSE_UNIT_NOT_FOUND", "Rumah/unit tidak ditemukan.")
	case errors.Is(err, ErrResidentNotFound):
		writeError(w, http.StatusNotFound, "RESIDENT_NOT_FOUND", "Warga tidak ditemukan.")
	case errors.Is(err, ErrHouseholdNotFound):
		writeError(w, http.StatusNotFound, "HOUSEHOLD_NOT_FOUND", "Keluarga tidak ditemukan.")
	case errors.Is(err, ErrCorrectionNotFound):
		writeError(w, http.StatusNotFound, "CORRECTION_NOT_FOUND", "Koreksi tidak ditemukan.")
	case errors.Is(err, ErrDuplicateData):
		writeError(w, http.StatusConflict, "DUPLICATE_DATA", "Data sudah digunakan.")
	case errors.Is(err, ErrConstraint):
		writeError(w, http.StatusConflict, "CONSTRAINT_VIOLATION", "Tindakan melanggar aturan bisnis.")
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{
		"code": code, "message": message,
	}})
}
