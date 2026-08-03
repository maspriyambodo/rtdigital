package reports

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

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
	mux.Handle("GET /reports/residents", h.require("resident.export", h.residents))
	mux.Handle("GET /reports/mutations", h.require("resident.export", h.mutations))
	mux.Handle("GET /reports/households", h.require("household.export", h.households))
	mux.Handle("GET /reports/invoices", h.require("invoice.export", h.invoices))
	mux.Handle("GET /reports/payments", h.require("finance.export", h.payments))
	mux.Handle("GET /reports/cash", h.require("finance.export", h.cash))
	mux.Handle("GET /reports/arrears", h.require("finance.export", h.arrears))
	mux.Handle("GET /reports/letters", h.require("letter_request.export", h.letters))
	mux.Handle("GET /reports/complaints", h.require("complaint.export", h.complaints))
}

func (h *Handler) require(permission string, next http.HandlerFunc) http.Handler {
	return auth.RequireAuthenticatedPermission(h.tokens, h.authz, permission, true, next)
}

func (h *Handler) residents(w http.ResponseWriter, r *http.Request) {
	principal := auth.PrincipalFromContext(r.Context())
	items, err := h.service.Residents(r.Context(), principal, parseFilter(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	if !isCSV(r) {
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
		return
	}
	if !h.auditExport(w, r, principal, "residents", len(items)) {
		return
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{item.ID, item.FullName, value(item.Gender), value(item.BirthDate), item.ResidentStatus, item.VerificationStatus, value(item.HouseUnitCode), value(item.HouseholdNumber), value(item.Relationship), item.CreatedAt})
	}
	writeCSV(w, "laporan_warga", []string{"ID", "Nama Lengkap", "Jenis Kelamin", "Tanggal Lahir", "Status Warga", "Status Verifikasi", "Unit Rumah", "Nomor Keluarga", "Hubungan", "Dibuat"}, rows)
}

func (h *Handler) mutations(w http.ResponseWriter, r *http.Request) {
	principal := auth.PrincipalFromContext(r.Context())
	items, err := h.service.Mutations(r.Context(), principal, parseFilter(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	if !isCSV(r) {
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
		return
	}
	if !h.auditExport(w, r, principal, "mutations", len(items)) {
		return
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{
			item.ID,
			item.FullName,
			item.HouseholdNumber,
			item.HouseUnitCode,
			item.Relationship,
			item.StartedAt,
			item.EndedAt,
			item.Status,
		})
	}
	writeCSV(w, "laporan_mutasi_warga", []string{"ID", "Nama Lengkap", "Nomor Keluarga", "Unit Rumah", "Hubungan", "Tanggal Mulai", "Tanggal Selesai", "Status"}, rows)
}

func (h *Handler) households(w http.ResponseWriter, r *http.Request) {
	principal := auth.PrincipalFromContext(r.Context())
	items, err := h.service.Households(r.Context(), principal, parseFilter(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	if !isCSV(r) {
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
		return
	}
	if !h.auditExport(w, r, principal, "households", len(items)) {
		return
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{item.ID, item.InternalNumber, item.HouseUnitCode, item.HeadResidentName, item.DomicileStatus, item.VerificationStatus, strconv.Itoa(item.ActiveMembersCount), item.MoveInDate})
	}
	writeCSV(w, "laporan_keluarga", []string{"ID", "Nomor Keluarga", "Unit Rumah", "Kepala Keluarga", "Status Domisili", "Status Verifikasi", "Anggota Aktif", "Tanggal Masuk"}, rows)
}

func (h *Handler) invoices(w http.ResponseWriter, r *http.Request) {
	principal := auth.PrincipalFromContext(r.Context())
	items, err := h.service.Invoices(r.Context(), principal, parseFilter(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	if !isCSV(r) {
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
		return
	}
	if !h.auditExport(w, r, principal, "invoices", len(items)) {
		return
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{item.ID, item.InvoiceNumber, item.HouseholdNumber, item.HouseUnitCode, item.DueTypeName, item.PeriodStart, item.PeriodEnd, item.DueDate, money(item.Amount), money(item.PaidAmount), money(item.AdjustmentAmount), item.Status})
	}
	writeCSV(w, "laporan_tagihan", []string{"ID", "Nomor Tagihan", "Nomor Keluarga", "Unit Rumah", "Jenis Iuran", "Periode Mulai", "Periode Akhir", "Jatuh Tempo", "Tagihan", "Terbayar", "Penyesuaian", "Status"}, rows)
}

func (h *Handler) payments(w http.ResponseWriter, r *http.Request) {
	principal := auth.PrincipalFromContext(r.Context())
	items, err := h.service.Payments(r.Context(), principal, parseFilter(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	if !isCSV(r) {
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
		return
	}
	if !h.auditExport(w, r, principal, "payments", len(items)) {
		return
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{item.ID, item.PaymentNumber, item.InvoiceNumber, item.HouseholdNumber, item.Method, money(item.Amount), item.PaidAt, item.VerificationStatus, value(item.VerifiedAt)})
	}
	writeCSV(w, "laporan_pembayaran", []string{"ID", "Nomor Pembayaran", "Nomor Tagihan", "Nomor Keluarga", "Metode", "Nominal", "Tanggal Bayar", "Status Verifikasi", "Tanggal Verifikasi"}, rows)
}

func (h *Handler) arrears(w http.ResponseWriter, r *http.Request) {
	principal := auth.PrincipalFromContext(r.Context())
	items, err := h.service.Arrears(r.Context(), principal, parseFilter(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	if !isCSV(r) {
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
		return
	}
	if !h.auditExport(w, r, principal, "arrears", len(items)) {
		return
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{
			item.HouseholdNumber,
			item.HouseUnitCode,
			item.HeadResidentName,
			strconv.Itoa(item.InvoiceCount),
			money(item.TotalArrears),
		})
	}
	writeCSV(w, "laporan_tunggakan", []string{"Nomor Keluarga", "Unit Rumah", "Kepala Keluarga", "Jumlah Tagihan", "Total Tunggakan"}, rows)
}

func (h *Handler) cash(w http.ResponseWriter, r *http.Request) {
	principal := auth.PrincipalFromContext(r.Context())
	items, err := h.service.Cash(r.Context(), principal, parseFilter(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	if !isCSV(r) {
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
		return
	}
	if !h.auditExport(w, r, principal, "cash", len(items)) {
		return
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{item.ID, item.TransactionNumber, item.Type, value(item.CategoryName), money(item.Amount), item.TransactionDate, item.Description, item.Status})
	}
	writeCSV(w, "laporan_kas", []string{"ID", "Nomor Transaksi", "Tipe", "Kategori", "Nominal", "Tanggal", "Deskripsi", "Status"}, rows)
}

func (h *Handler) letters(w http.ResponseWriter, r *http.Request) {
	principal := auth.PrincipalFromContext(r.Context())
	items, err := h.service.Letters(r.Context(), principal, parseFilter(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	if !isCSV(r) {
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
		return
	}
	if !h.auditExport(w, r, principal, "letters", len(items)) {
		return
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{item.ID, item.RequestNumber, value(item.LetterNumber), item.LetterTypeName, item.RequesterName, item.ResidentName, item.Status, value(item.SubmittedAt), value(item.IssuedAt)})
	}
	writeCSV(w, "laporan_surat", []string{"ID", "Nomor Pengajuan", "Nomor Surat", "Jenis Surat", "Pemohon", "Subjek Warga", "Status", "Diajukan", "Diterbitkan"}, rows)
}

func (h *Handler) complaints(w http.ResponseWriter, r *http.Request) {
	principal := auth.PrincipalFromContext(r.Context())
	items, err := h.service.Complaints(r.Context(), principal, parseFilter(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	if !isCSV(r) {
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
		return
	}
	if !h.auditExport(w, r, principal, "complaints", len(items)) {
		return
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{item.ID, item.TicketNumber, item.Category, item.Title, item.Priority, item.Status, item.ReporterName, value(item.AssignedToName), item.CreatedAt, value(item.ResolvedAt)})
	}
	writeCSV(w, "laporan_aduan", []string{"ID", "Nomor Tiket", "Kategori", "Judul", "Prioritas", "Status", "Pelapor", "Petugas", "Dibuat", "Diselesaikan"}, rows)
}

func (h *Handler) auditExport(w http.ResponseWriter, r *http.Request, principal *auth.Principal, reportType string, recordCount int) bool {
	if err := h.service.RecordExportAudit(r.Context(), principal, reportType, recordCount); err != nil {
		writeServiceError(w, r, err)
		return false
	}
	return true
}

func parseFilter(r *http.Request) Filter {
	query := r.URL.Query()
	return Filter{StartDate: query.Get("start_date"), EndDate: query.Get("end_date"), Status: query.Get("status")}
}

func isCSV(r *http.Request) bool {
	return r.URL.Query().Get("format") == "csv"
}

func writeCSV(w http.ResponseWriter, filename string, header []string, rows [][]string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-%s.csv"`, filename, time.Now().UTC().Format("20060102150405")))
	w.Header().Set("Cache-Control", "private, no-store")
	writer := csv.NewWriter(w)
	_ = writer.Write(header)
	_ = writer.WriteAll(rows)
}

func money(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

func value(item *string) string {
	if item == nil {
		return ""
	}
	return *item
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		writeError(w, r, http.StatusBadRequest, "VALIDATION_FAILED", "Filter tidak valid.")
	case errors.Is(err, ErrForbidden):
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "Akses tidak diizinkan.")
	default:
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memuat laporan.")
	}
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]any{
		"code":       code,
		"message":    message,
		"details":    []any{},
		"request_id": r.Header.Get("X-Request-ID"),
	}})
}