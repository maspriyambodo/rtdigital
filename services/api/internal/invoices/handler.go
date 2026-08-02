package invoices

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
	mux.Handle("GET /due-types", h.require("due_type.read", h.listDueTypes))
	mux.Handle("POST /due-types", h.require("due_type.create", h.createDueType))
	mux.Handle("GET /due-types/{id}", h.require("due_type.read", h.getDueType))
	mux.Handle("PATCH /due-types/{id}", h.require("due_type.update", h.updateDueType))
	mux.Handle("POST /due-types/{id}/deactivate", h.require("due_type.deactivate", h.deactivateDueType))

	mux.Handle("GET /invoices", h.require("invoice.read", h.listInvoices))
	mux.Handle("POST /invoices", h.require("invoice.create", h.createInvoice))
	mux.Handle("POST /invoices/generate", h.require("invoice.create", h.generateInvoices))
	mux.Handle("GET /invoices/{id}", h.require("invoice.read", h.getInvoice))
	mux.Handle("PATCH /invoices/{id}", h.require("invoice.update", h.updateInvoice))
	mux.Handle("POST /invoices/{id}/cancel", h.require("invoice.cancel", h.cancelInvoice))
}

func (h *Handler) require(permission string, next http.HandlerFunc) http.Handler {
	return auth.RequireAuthenticatedPermission(h.tokens, h.authz, permission, false, next)
}

func (h *Handler) listDueTypes(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListDueTypes(r.Context(), auth.PrincipalFromContext(r.Context()), r.URL.Query().Get("status"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createDueType(w http.ResponseWriter, r *http.Request) {
	var request CreateDueTypeRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.CreateDueType(r.Context(), auth.PrincipalFromContext(r.Context()), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) getDueType(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetDueType(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) updateDueType(w http.ResponseWriter, r *http.Request) {
	var request UpdateDueTypeRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.UpdateDueType(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) deactivateDueType(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeactivateDueType(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id")); err != nil {
		h.writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listInvoices(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	items, err := h.service.ListInvoices(r.Context(), auth.PrincipalFromContext(r.Context()), InvoiceFilter{
		Status:      query.Get("status"),
		DueTypeID:   query.Get("due_type_id"),
		HouseholdID: query.Get("household_id"),
		PeriodStart: query.Get("period_start"),
		PeriodEnd:   query.Get("period_end"),
		OnlyArrears: query.Get("arrears") == "true",
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createInvoice(w http.ResponseWriter, r *http.Request) {
	var request CreateInvoiceRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.CreateInvoice(r.Context(), auth.PrincipalFromContext(r.Context()), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) generateInvoices(w http.ResponseWriter, r *http.Request) {
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		writeError(w, http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED", "Header Idempotency-Key wajib disertakan.")
		return
	}
	var request GenerateInvoicesRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	result, err := h.service.GenerateInvoices(r.Context(), auth.PrincipalFromContext(r.Context()), idempotencyKey, request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": result})
}

func (h *Handler) getInvoice(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetInvoice(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) updateInvoice(w http.ResponseWriter, r *http.Request) {
	var request UpdateInvoiceRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.UpdateInvoice(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) cancelInvoice(w http.ResponseWriter, r *http.Request) {
	var request CancelInvoiceRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.CancelInvoice(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
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
	case errors.Is(err, ErrDueTypeNotFound):
		writeError(w, http.StatusNotFound, "DUE_TYPE_NOT_FOUND", "Jenis iuran tidak ditemukan.")
	case errors.Is(err, ErrInvoiceNotFound):
		writeError(w, http.StatusNotFound, "INVOICE_NOT_FOUND", "Tagihan tidak ditemukan.")
	case errors.Is(err, ErrHouseholdNotFound):
		writeError(w, http.StatusNotFound, "HOUSEHOLD_NOT_FOUND", "Keluarga tidak ditemukan.")
	case errors.Is(err, ErrDuplicateData):
		writeError(w, http.StatusConflict, "DUPLICATE_DATA", "Data sudah digunakan.")
	case errors.Is(err, ErrConstraint), errors.Is(err, ErrInvalidInvoiceStatus):
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
		"code":    code,
		"message": message,
	}})
}
