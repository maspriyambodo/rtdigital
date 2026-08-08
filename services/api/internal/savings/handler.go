package savings

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
	// Products
	mux.Handle("GET /savings/products", h.require("savings.read", h.listProducts))
	mux.Handle("POST /savings/products", h.require("savings.manage", h.createProduct))
	mux.Handle("GET /savings/products/{id}", h.require("savings.read", h.getProduct))
	mux.Handle("PATCH /savings/products/{id}", h.require("savings.manage", h.updateProduct))

	// Accounts
	mux.Handle("GET /savings/accounts", h.require("savings.read", h.listAccounts))
	mux.Handle("POST /savings/accounts", h.require("savings.write", h.createAccount))
	mux.Handle("GET /savings/accounts/{id}", h.require("savings.read", h.getAccount))
	mux.Handle("POST /savings/accounts/{id}/close", h.require("savings.manage", h.closeAccount))

	// Transactions
	mux.Handle("GET /savings/transactions", h.require("savings.read", h.listTransactions))
	mux.Handle("POST /savings/deposit", h.require("savings.write", h.deposit))
	mux.Handle("POST /savings/withdraw", h.require("savings.write", h.withdraw))
	mux.Handle("POST /savings/transactions/{id}/verify", h.require("savings.verify", h.verifyTransaction))
	mux.Handle("POST /savings/transactions/{id}/reverse", h.require("savings.manage", h.reverseTransaction))

	// Reports
	mux.Handle("GET /savings/reports/reconciliation", h.require("savings.read", h.getReconciliation))
}

func (h *Handler) require(permission string, next http.HandlerFunc) http.Handler {
	return auth.RequireAuthenticatedPermission(h.tokens, h.authz, permission, false, next)
}


func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	prods, err := h.service.ListProducts(r.Context(), auth.PrincipalFromContext(r.Context()), status)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": prods})
}

func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductReq
	if !decodeJSON(w, r, &req) {
		return
	}
	prod, err := h.service.CreateProduct(r.Context(), auth.PrincipalFromContext(r.Context()), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": prod})
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	prod, err := h.service.GetProduct(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": prod})
}

func (h *Handler) updateProduct(w http.ResponseWriter, r *http.Request) {
	var req UpdateProductReq
	if !decodeJSON(w, r, &req) {
		return
	}
	prod, err := h.service.UpdateProduct(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": prod})
}

func (h *Handler) listAccounts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	accounts, err := h.service.ListAccounts(r.Context(), auth.PrincipalFromContext(r.Context()), AccountFilter{
		ProductID:   q.Get("product_id"),
		HouseholdID: q.Get("household_id"),
		Status:      q.Get("status"),
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": accounts})
}

func (h *Handler) createAccount(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountReq
	if !decodeJSON(w, r, &req) {
		return
	}
	acc, err := h.service.CreateAccount(r.Context(), auth.PrincipalFromContext(r.Context()), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": acc})
}

func (h *Handler) getAccount(w http.ResponseWriter, r *http.Request) {
	acc, err := h.service.GetAccount(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": acc})
}

func (h *Handler) closeAccount(w http.ResponseWriter, r *http.Request) {
	acc, err := h.service.CloseAccount(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": acc})
}



func (h *Handler) listTransactions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	txs, err := h.service.ListTransactions(r.Context(), auth.PrincipalFromContext(r.Context()), TransactionFilter{
		AccountID:          q.Get("account_id"),
		ProductID:          q.Get("product_id"),
		HouseholdID:        q.Get("household_id"),
		Type:               q.Get("type"),
		VerificationStatus: q.Get("verification_status"),
		StartDate:          q.Get("start_date"),
		EndDate:            q.Get("end_date"),
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": txs})
}

func (h *Handler) deposit(w http.ResponseWriter, r *http.Request) {
	var req DepositReq
	if !decodeJSON(w, r, &req) {
		return
	}
	tx, err := h.service.Deposit(r.Context(), auth.PrincipalFromContext(r.Context()), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": tx})
}

func (h *Handler) withdraw(w http.ResponseWriter, r *http.Request) {
	var req WithdrawReq
	if !decodeJSON(w, r, &req) {
		return
	}
	tx, err := h.service.Withdraw(r.Context(), auth.PrincipalFromContext(r.Context()), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": tx})
}

func (h *Handler) verifyTransaction(w http.ResponseWriter, r *http.Request) {
	var req VerifyTxReq
	if !decodeJSON(w, r, &req) {
		return
	}
	tx, err := h.service.VerifyTransaction(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": tx})
}

func (h *Handler) reverseTransaction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Description string `json:"description"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	tx, err := h.service.ReverseTransaction(r.Context(), auth.PrincipalFromContext(r.Context()), r.PathValue("id"), req.Description)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": tx})
}

func (h *Handler) getReconciliation(w http.ResponseWriter, r *http.Request) {
	reports, err := h.service.GetReconciliationReport(r.Context(), auth.PrincipalFromContext(r.Context()))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": reports})
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Data tidak valid.")
	case errors.Is(err, ErrProductNotFound):
		writeError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Produk tabungan tidak ditemukan.")
	case errors.Is(err, ErrAccountNotFound):
		writeError(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "Akun tabungan tidak ditemukan.")
	case errors.Is(err, ErrTransactionNotFound):
		writeError(w, http.StatusNotFound, "TRANSACTION_NOT_FOUND", "Transaksi tabungan tidak ditemukan.")
	case errors.Is(err, ErrDuplicateData):
		writeError(w, http.StatusConflict, "DUPLICATE_DATA", "Data tabungan sudah ada.")
	case errors.Is(err, ErrConstraint):
		writeError(w, http.StatusConflict, "CONSTRAINT_VIOLATION", "Tindakan melanggar aturan bisnis tabungan.")
	case errors.Is(err, ErrInvalidState):
		writeError(w, http.StatusConflict, "INVALID_STATE", "Status tidak mengizinkan tindakan ini.")
	case errors.Is(err, ErrInsufficientBalance):
		writeError(w, http.StatusConflict, "INSUFFICIENT_BALANCE", "Saldo tabungan tidak mencukupi.")
	case errors.Is(err, ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Akses ditolak (misal: verifikasi sendiri dilarang).")
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
