package assets

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
	mux.Handle("GET /asset-categories", h.require("asset.read", h.listCategories))
	mux.Handle("POST /asset-categories", h.require("asset.manage", h.createCategory))
	mux.Handle("PATCH /asset-categories/{id}", h.require("asset.manage", h.updateCategory))

	mux.Handle("GET /asset-locations", h.require("asset.read", h.listLocations))
	mux.Handle("POST /asset-locations", h.require("asset.manage", h.createLocation))
	mux.Handle("PATCH /asset-locations/{id}", h.require("asset.manage", h.updateLocation))

	mux.Handle("GET /assets", h.require("asset.read", h.listAssets))
	mux.Handle("POST /assets", h.require("asset.manage", h.createAsset))
	mux.Handle("GET /assets/{id}", h.require("asset.read", h.getAsset))
	mux.Handle("PATCH /assets/{id}", h.require("asset.manage", h.updateAsset))

	mux.Handle("GET /asset-loans", h.require("asset.read", h.listLoans))
	mux.Handle("POST /asset-loans", h.require("asset.loan", h.createLoan))
	mux.Handle("POST /asset-loans/{id}/review", h.require("asset.manage", h.reviewLoan))
	mux.Handle("POST /asset-loans/{id}/return", h.require("asset.manage", h.returnLoan))

	mux.Handle("GET /asset-maintenances", h.require("asset.read", h.listMaintenances))
	mux.Handle("POST /asset-maintenances", h.require("asset.maintain", h.createMaintenance))
}

func (h *Handler) require(permission string, next http.HandlerFunc) http.Handler {
	return auth.RequireAuthenticatedPermission(h.tokens, h.authz, permission, false, next)
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListCategories(r.Context(), auth.PrincipalFromContext(r.Context()))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.CreateCategory(r.Context(), auth.PrincipalFromContext(r.Context()), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	var req UpdateCategoryRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.UpdateCategory(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) listLocations(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListLocations(r.Context(), auth.PrincipalFromContext(r.Context()))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createLocation(w http.ResponseWriter, r *http.Request) {
	var req CreateLocationRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.CreateLocation(r.Context(), auth.PrincipalFromContext(r.Context()), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) updateLocation(w http.ResponseWriter, r *http.Request) {
	var req UpdateLocationRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.UpdateLocation(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) listAssets(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	items, err := h.service.ListAssets(
		r.Context(), auth.PrincipalFromContext(r.Context()),
		q.Get("category_id"), q.Get("location_id"), q.Get("status"),
	)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) getAsset(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetAsset(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) createAsset(w http.ResponseWriter, r *http.Request) {
	var req CreateAssetRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.CreateAsset(r.Context(), auth.PrincipalFromContext(r.Context()), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) updateAsset(w http.ResponseWriter, r *http.Request) {
	var req UpdateAssetRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.UpdateAsset(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) listLoans(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	items, err := h.service.ListLoans(r.Context(), auth.PrincipalFromContext(r.Context()), q.Get("asset_id"), q.Get("status"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createLoan(w http.ResponseWriter, r *http.Request) {
	var req CreateLoanRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.CreateLoan(r.Context(), auth.PrincipalFromContext(r.Context()), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) reviewLoan(w http.ResponseWriter, r *http.Request) {
	var req ReviewLoanRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.ReviewLoan(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) returnLoan(w http.ResponseWriter, r *http.Request) {
	var req ReturnLoanRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.ReturnLoan(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) listMaintenances(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListMaintenanceLogs(r.Context(), auth.PrincipalFromContext(r.Context()), r.URL.Query().Get("asset_id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createMaintenance(w http.ResponseWriter, r *http.Request) {
	var req CreateMaintenanceLogRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	item, err := h.service.CreateMaintenanceLog(r.Context(), auth.PrincipalFromContext(r.Context()), req)
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
	case errors.Is(err, ErrCategoryNotFound):
		writeError(w, http.StatusNotFound, "ASSET_CATEGORY_NOT_FOUND", "Kategori aset tidak ditemukan.")
	case errors.Is(err, ErrLocationNotFound):
		writeError(w, http.StatusNotFound, "ASSET_LOCATION_NOT_FOUND", "Lokasi aset tidak ditemukan.")
	case errors.Is(err, ErrAssetNotFound):
		writeError(w, http.StatusNotFound, "ASSET_NOT_FOUND", "Aset tidak ditemukan.")
	case errors.Is(err, ErrLoanNotFound):
		writeError(w, http.StatusNotFound, "ASSET_LOAN_NOT_FOUND", "Peminjaman aset tidak ditemukan.")
	case errors.Is(err, ErrDuplicateCode):
		writeError(w, http.StatusConflict, "DUPLICATE_CODE", "Kode aset sudah digunakan.")
	case errors.Is(err, ErrInvalidState):
		writeError(w, http.StatusConflict, "INVALID_STATE", "Status aset tidak mengizinkan operasi ini.")
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
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
