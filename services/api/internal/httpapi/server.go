package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maspriyambodo/rtdigital/services/api/internal/assets"
	"github.com/maspriyambodo/rtdigital/services/api/internal/audit"
	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
	"github.com/maspriyambodo/rtdigital/services/api/internal/cash"
	"github.com/maspriyambodo/rtdigital/services/api/internal/communication"
	"github.com/maspriyambodo/rtdigital/services/api/internal/complaints"
	"github.com/maspriyambodo/rtdigital/services/api/internal/dashboard"
	"github.com/maspriyambodo/rtdigital/services/api/internal/files"
	"github.com/maspriyambodo/rtdigital/services/api/internal/invoices"
	"github.com/maspriyambodo/rtdigital/services/api/internal/letters"
	"github.com/maspriyambodo/rtdigital/services/api/internal/notifications"
	"github.com/maspriyambodo/rtdigital/services/api/internal/payments"
	"github.com/maspriyambodo/rtdigital/services/api/internal/reports"
	"github.com/maspriyambodo/rtdigital/services/api/internal/residents"
	"github.com/maspriyambodo/rtdigital/services/api/internal/savings"
	"github.com/maspriyambodo/rtdigital/services/api/internal/securityops"
	"github.com/maspriyambodo/rtdigital/services/api/internal/settings"
	"github.com/maspriyambodo/rtdigital/services/api/internal/users"
)

type requestIDKey struct{}

type Server struct {
	handler http.Handler
}

func NewServer(logger *slog.Logger, db *pgxpool.Pool, tokens *auth.TokenManager, authService *auth.Service, authz *auth.AuthorizationService, usersService *users.Service, residentsService *residents.Service, invoicesService *invoices.Service, filesService *files.Service, paymentsService *payments.Service, cashService *cash.Service, production bool, services ...any) *Server {
	var communicationService *communication.Service
	var lettersService *letters.Service
	var complaintsService *complaints.Service
	var notificationsService *notifications.Service
	var dashboardService *dashboard.Service
	var reportsService *reports.Service
	var settingsService *settings.Service
	var auditService *audit.Service
	var savingsService *savings.Service
	var assetsService *assets.Service
	var securityOpsService *securityops.Service
	var letterStorage letters.StorageClient
	for _, service := range services {
		switch value := service.(type) {
		case *communication.Service:
			communicationService = value
		case *letters.Service:
			lettersService = value
		case *complaints.Service:
			complaintsService = value
		case *notifications.Service:
			notificationsService = value
		case *dashboard.Service:
			dashboardService = value
		case *reports.Service:
			reportsService = value
		case *settings.Service:
			settingsService = value
		case *audit.Service:
			auditService = value
		case *savings.Service:
			savingsService = value
		case *assets.Service:
			assetsService = value
		case *securityops.Service:
			securityOpsService = value
		case letters.StorageClient:
			letterStorage = value
		}
	}

	root := http.NewServeMux()
	root.HandleFunc("GET /healthz", liveness)
	root.HandleFunc("GET /readyz", readiness(db))

	api := http.NewServeMux()
	auth.NewHandler(authService, tokens, authz, production).RegisterRoutes(api)
	users.NewHandler(usersService, tokens, authz).RegisterRoutes(api)
	residents.NewHandler(residentsService, tokens, authz).RegisterRoutes(api)
	invoices.NewHandler(invoicesService, tokens, authz).RegisterRoutes(api)
	if filesService != nil {
		files.NewHandler(filesService, tokens, authz).RegisterRoutes(api)
	}
	if paymentsService != nil {
		payments.NewHandler(paymentsService, tokens, authz).RegisterRoutes(api)
	}
	if cashService != nil {
		cash.NewHandler(cashService, tokens, authz).RegisterRoutes(api)
	}
	if communicationService != nil {
		communication.NewHandler(communicationService, tokens, authz).RegisterRoutes(api)
	}
	if lettersService != nil {
		letters.NewHandler(lettersService, tokens, authz, letterStorage, filesService).RegisterRoutes(api)
	}
	if complaintsService != nil {
		complaints.NewHandler(complaintsService, tokens, authz).RegisterRoutes(api)
	}
	if notificationsService != nil {
		notifications.NewHandler(notificationsService, tokens, authz).RegisterRoutes(api)
	}
	if dashboardService != nil {
		dashboard.NewHandler(dashboardService, tokens, authz).RegisterRoutes(api)
	}
	if reportsService != nil {
		reports.NewHandler(reportsService, tokens, authz).RegisterRoutes(api)
	}
	if settingsService != nil {
		settings.NewHandler(settingsService, tokens, authz).RegisterRoutes(api)
	}
	if auditService != nil {
		audit.NewHandler(auditService, tokens, authz).RegisterRoutes(api)
	}
	if savingsService != nil {
		savings.NewHandler(savingsService, tokens, authz).RegisterRoutes(api)
	}
	if assetsService != nil {
		assets.NewHandler(assetsService, tokens, authz).RegisterRoutes(api)
	}
	if securityOpsService != nil {
		securityops.NewHandler(securityOpsService, tokens, authz).RegisterRoutes(api)
	}
	root.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	origins := allowedOrigins()
	limiter := newRateLimiter(60, time.Minute)
	handler := withRequestID(
		withRecovery(
			logger,
			withSanitizedLogging(
				logger,
				withSecurityHeaders(
					production,
					withCORS(
						origins,
						withCSRF(
							origins,
							withRateLimit(
								limiter,
								withBodyLimit(root),
							),
						),
					),
				),
			),
		),
	)

	return &Server{handler: handler}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func liveness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func readiness(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unavailable"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	}
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}

func withRecovery(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.ErrorContext(r.Context(), "panic recovered",
					"panic", recovered,
					"stack", string(debug.Stack()),
					"request_id", RequestID(r.Context()),
				)
				http.Error(w, `{"code":"internal_error","message":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func newRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "unavailable"
	}
	return hex.EncodeToString(bytes)
}
