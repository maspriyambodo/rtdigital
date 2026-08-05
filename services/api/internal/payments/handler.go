package payments

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
	mux.Handle("GET /payments", h.require("payment.read", h.list))
	mux.Handle("POST /payments", h.require("payment.submit", h.submit))
	mux.Handle("GET /payments/{id}", h.require("payment.read", h.get))
	mux.Handle("GET /payments/queue", h.require("payment.verify", h.queue))
	mux.Handle("POST /payments/{id}/verify", h.require("payment.verify", h.verify))
	mux.Handle("POST /payments/{id}/reject", h.require("payment.reject", h.reject))
	mux.Handle("POST /payments/{id}/cancel", h.require("payment.cancel", h.cancel))
}

func (h *Handler) require(permission string, next http.HandlerFunc) http.Handler {
	return auth.RequireAuthenticatedPermission(h.tokens, h.authz, permission, false, next)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	items, err := h.service.List(r.Context(), auth.PrincipalFromContext(r.Context()), PaymentFilter{
		InvoiceID:          query.Get("invoice_id"),
		VerificationStatus: query.Get("status"),
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) queue(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.Queue(r.Context(), auth.PrincipalFromContext(r.Context()))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		writeError(w, http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED", "Header Idempotency-Key wajib disertakan.")
		return
	}

	var request SubmitPaymentRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	response, err := h.service.Submit(r.Context(), auth.PrincipalFromContext(r.Context()), idempotencyKey, request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": response})
}

func (h *Handler) verify(w http.ResponseWriter, r *http.Request) {
	var request VerifyPaymentRequest
	if r.ContentLength > 0 && !decodeJSON(w, r, &request) {
		return
	}

	response, err := h.service.Verify(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": response})
}

func (h *Handler) reject(w http.ResponseWriter, r *http.Request) {
	var request RejectPaymentRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	response, err := h.service.Reject(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": response})
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	var request CancelPaymentRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	response, err := h.service.Cancel(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": response})
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Data permintaan tidak valid.")
	case errors.Is(err, ErrPaymentNotFound):
		writeError(w, http.StatusNotFound, "PAYMENT_NOT_FOUND", "Pembayaran tidak ditemukan.")
	case errors.Is(err, ErrInvoiceNotFound):
		writeError(w, http.StatusNotFound, "INVOICE_NOT_FOUND", "Tagihan tidak ditemukan.")
	case errors.Is(err, ErrFileNotFound):
		writeError(w, http.StatusNotFound, "FILE_NOT_FOUND", "Bukti pembayaran tidak ditemukan.")
	case errors.Is(err, ErrFileNotConfirmed):
		writeError(w, http.StatusConflict, "FILE_NOT_CONFIRMED", "Bukti pembayaran belum terkonfirmasi.")
	case errors.Is(err, ErrDuplicateData):
		writeError(w, http.StatusConflict, "DUPLICATE_DATA", "Data pembayaran sudah digunakan.")
	case errors.Is(err, ErrConstraint):
		writeError(w, http.StatusConflict, "CONSTRAINT_VIOLATION", "Tindakan melanggar aturan bisnis keuangan.")
	case errors.Is(err, ErrInvalidPaymentState):
		writeError(w, http.StatusConflict, "INVALID_PAYMENT_STATE", "Status pembayaran tidak mengizinkan tindakan ini.")
	case errors.Is(err, ErrInvalidInvoiceState):
		writeError(w, http.StatusConflict, "INVALID_INVOICE_STATE", "Status tagihan tidak mengizinkan pembayaran baru.")
	case errors.Is(err, ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Akses atau pemisahan tugas tidak mengizinkan tindakan ini.")
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
