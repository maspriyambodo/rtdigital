package cash

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
	mux.Handle("GET /cash/book", h.require("cash.read", h.getBook))
	mux.Handle("GET /cash/categories", h.require("cash.read", h.listCategories))
	mux.Handle("POST /cash/categories", h.require("cash.create", h.createCategory))
	mux.Handle("PATCH /cash/categories/{id}", h.require("cash.update", h.updateCategory))
	mux.Handle("POST /cash/transactions", h.require("cash.create", h.recordManual))
	mux.Handle("POST /cash/transactions/{id}/reverse", h.require("cash.reverse", h.reverse))
}

func (h *Handler) require(permission string, next http.HandlerFunc) http.Handler {
	return auth.RequireAuthenticatedPermission(h.tokens, h.authz, permission, false, next)
}

func (h *Handler) getBook(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	book, err := h.service.GetBook(r.Context(), auth.PrincipalFromContext(r.Context()), TransactionFilter{
		StartDate:  query.Get("start_date"),
		EndDate:    query.Get("end_date"),
		Type:       query.Get("type"),
		CategoryID: query.Get("category_id"),
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": book})
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.ListCategories(r.Context(), auth.PrincipalFromContext(r.Context()))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": categories})
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var request CreateCategoryRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	category, err := h.service.CreateCategory(r.Context(), auth.PrincipalFromContext(r.Context()), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": category})
}

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	var request UpdateCategoryRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	category, err := h.service.UpdateCategory(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": category})
}

func (h *Handler) recordManual(w http.ResponseWriter, r *http.Request) {
	var request RecordTransactionRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	transaction, err := h.service.RecordManual(r.Context(), auth.PrincipalFromContext(r.Context()), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": transaction})
}

func (h *Handler) reverse(w http.ResponseWriter, r *http.Request) {
	var request ReverseTransactionRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	transaction, err := h.service.Reverse(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": transaction})
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Data permintaan tidak valid.")
	case errors.Is(err, ErrCategoryNotFound):
		writeError(w, http.StatusNotFound, "CATEGORY_NOT_FOUND", "Kategori kas tidak ditemukan.")
	case errors.Is(err, ErrTransactionNotFound):
		writeError(w, http.StatusNotFound, "TRANSACTION_NOT_FOUND", "Transaksi kas tidak ditemukan.")
	case errors.Is(err, ErrDuplicateData):
		writeError(w, http.StatusConflict, "DUPLICATE_DATA", "Data sudah digunakan.")
	case errors.Is(err, ErrConstraint):
		writeError(w, http.StatusConflict, "CONSTRAINT_VIOLATION", "Tindakan melanggar aturan bisnis keuangan.")
	case errors.Is(err, ErrInvalidState):
		writeError(w, http.StatusConflict, "INVALID_TRANSACTION_STATE", "Status transaksi kas tidak mengizinkan tindakan ini.")
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
