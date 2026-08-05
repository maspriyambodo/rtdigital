package letters

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
	"github.com/maspriyambodo/rtdigital/services/api/internal/files"
)

type Handler struct {
	service *Service
	tokens  *auth.TokenManager
	authz   *auth.AuthorizationService
	storage StorageClient
	files   *files.Service
}

func NewHandler(service *Service, tokens *auth.TokenManager, authz *auth.AuthorizationService, storage StorageClient, filesService *files.Service) *Handler {
	return &Handler{service: service, tokens: tokens, authz: authz, storage: storage, files: filesService}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /letter-types", h.require("letter_type.read", h.listLetterTypes))
	mux.Handle("POST /letter-types", h.require("letter_type.create", h.createLetterType))
	mux.Handle("GET /letter-types/{id}", h.require("letter_type.read", h.getLetterType))
	mux.Handle("PATCH /letter-types/{id}", h.require("letter_type.update", h.updateLetterType))
	mux.Handle("POST /letter-types/{id}/deactivate", h.require("letter_type.deactivate", h.deactivateLetterType))

	mux.Handle("GET /letter-requests", h.require("letter_request.read", h.listLetterRequests))
	mux.Handle("POST /letter-requests", h.require("letter_request.submit", h.submitLetterRequest))
	mux.Handle("GET /letter-requests/{id}", h.require("letter_request.read", h.getLetterRequest))
	mux.Handle("PATCH /letter-requests/{id}", h.require("letter_request.submit", h.updateLetterRequest))
	mux.Handle("POST /letter-requests/{id}/process", h.require("letter_request.process", h.processLetterRequest))
	mux.Handle("POST /letter-requests/{id}/request-revision", h.require("letter_request.request_revision", h.requestRevision))
	mux.Handle("POST /letter-requests/{id}/approve", h.require("letter_request.approve", h.approveLetterRequest))
	mux.Handle("POST /letter-requests/{id}/reject", h.require("letter_request.process", h.rejectLetterRequest))
	mux.Handle("POST /letter-requests/{id}/issue", h.require("letter_request.issue", h.issueLetter))
	mux.Handle("POST /letter-requests/{id}/cancel", h.require("letter_request.issue", h.cancelLetterRequest))
	mux.Handle("GET /letter-requests/{id}/download", h.require("letter_request.download", h.downloadLetter))
	mux.Handle("GET /letters/verify/{code}", http.HandlerFunc(h.verifyPublic))
}

func (h *Handler) require(permission string, next http.HandlerFunc) http.Handler {
	return auth.RequireAuthenticatedPermission(h.tokens, h.authz, permission, false, next)
}

func (h *Handler) listLetterTypes(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListLetterTypes(r.Context(), auth.PrincipalFromContext(r.Context()), r.URL.Query().Get("include_inactive") == "true")
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) createLetterType(w http.ResponseWriter, r *http.Request) {
	var request CreateLetterTypeRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.CreateLetterType(r.Context(), auth.PrincipalFromContext(r.Context()), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) getLetterType(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetLetterType(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) updateLetterType(w http.ResponseWriter, r *http.Request) {
	var request UpdateLetterTypeRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.UpdateLetterType(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) deactivateLetterType(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.DeactivateLetterType(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) listLetterRequests(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	items, err := h.service.ListLetterRequests(r.Context(), auth.PrincipalFromContext(r.Context()), LetterRequestFilter{
		Status: query.Get("status"), LetterTypeID: query.Get("letter_type_id"), Search: query.Get("search"),
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) submitLetterRequest(w http.ResponseWriter, r *http.Request) {
	var request SubmitLetterRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.SubmitLetterRequest(r.Context(), auth.PrincipalFromContext(r.Context()), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h *Handler) getLetterRequest(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetLetterRequest(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) updateLetterRequest(w http.ResponseWriter, r *http.Request) {
	var request UpdateLetterRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.UpdateLetterRequest(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) processLetterRequest(w http.ResponseWriter, r *http.Request) {
	h.review(w, r, func(request ReviewLetterRequest) (LetterRequestItem, error) {
		return h.service.ProcessLetterRequest(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	})
}

func (h *Handler) requestRevision(w http.ResponseWriter, r *http.Request) {
	h.review(w, r, func(request ReviewLetterRequest) (LetterRequestItem, error) {
		return h.service.RequestRevision(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	})
}

func (h *Handler) approveLetterRequest(w http.ResponseWriter, r *http.Request) {
	h.review(w, r, func(request ReviewLetterRequest) (LetterRequestItem, error) {
		return h.service.ApproveLetterRequest(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	})
}

func (h *Handler) rejectLetterRequest(w http.ResponseWriter, r *http.Request) {
	h.review(w, r, func(request ReviewLetterRequest) (LetterRequestItem, error) {
		return h.service.RejectLetterRequest(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), request)
	})
}

func (h *Handler) review(w http.ResponseWriter, r *http.Request, action func(ReviewLetterRequest) (LetterRequestItem, error)) {
	var request ReviewLetterRequest
	if r.ContentLength > 0 && !decodeJSON(w, r, &request) {
		return
	}
	item, err := action(request)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) issueLetter(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.IssueLetter(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), h.storage)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) cancelLetterRequest(w http.ResponseWriter, r *http.Request) {
	var request CancelLetterRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.service.CancelLetterRequest(
		r.Context(),
		auth.PrincipalFromContext(r.Context()),
		r.PathValue("id"),
		request,
	)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) downloadLetter(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetLetterRequest(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	if item.Status != "issued" || item.IssuedFileID == nil || h.files == nil {
		h.writeServiceError(w, ErrInvalidState)
		return
	}
	response, err := h.files.PresignDownload(r.Context(), auth.PrincipalFromContext(r.Context()), *item.IssuedFileID)
	if err != nil {
		h.writeServiceError(w, ErrForbidden)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": response})
}

func (h *Handler) verifyPublic(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.VerifyPublicLetter(r.Context(), r.PathValue("code"))
	if err != nil {
		writeError(w, http.StatusNotFound, "LETTER_NOT_FOUND", "Surat tidak ditemukan atau kode verifikasi tidak valid.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Data permintaan tidak valid.")
	case errors.Is(err, ErrLetterTypeNotFound):
		writeError(w, http.StatusNotFound, "LETTER_TYPE_NOT_FOUND", "Jenis surat tidak ditemukan.")
	case errors.Is(err, ErrLetterRequestNotFound):
		writeError(w, http.StatusNotFound, "LETTER_REQUEST_NOT_FOUND", "Pengajuan surat tidak ditemukan.")
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
