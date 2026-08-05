package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maspriyambodo/rtdigital/services/api/internal/audit"
	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
	"github.com/maspriyambodo/rtdigital/services/api/internal/cash"
	"github.com/maspriyambodo/rtdigital/services/api/internal/communication"
	"github.com/maspriyambodo/rtdigital/services/api/internal/complaints"
	"github.com/maspriyambodo/rtdigital/services/api/internal/config"
	"github.com/maspriyambodo/rtdigital/services/api/internal/dashboard"
	"github.com/maspriyambodo/rtdigital/services/api/internal/files"
	"github.com/maspriyambodo/rtdigital/services/api/internal/httpapi"
	"github.com/maspriyambodo/rtdigital/services/api/internal/invoices"
	"github.com/maspriyambodo/rtdigital/services/api/internal/letters"
	"github.com/maspriyambodo/rtdigital/services/api/internal/notifications"
	"github.com/maspriyambodo/rtdigital/services/api/internal/payments"
	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
	"github.com/maspriyambodo/rtdigital/services/api/internal/reports"
	"github.com/maspriyambodo/rtdigital/services/api/internal/residents"
	"github.com/maspriyambodo/rtdigital/services/api/internal/settings"
	"github.com/maspriyambodo/rtdigital/services/api/internal/users"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	initCtx, cancelInit := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelInit()

	pool, err := platform.NewDatabase(initCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	tokens, err := auth.NewTokenManager(cfg.JWTSecret)
	if err != nil {
		logger.Error("failed to initialize token manager", "error", err)
		os.Exit(1)
	}

	crypter, err := auth.NewAESCrypter([]byte(cfg.DataEncryptionKey))
	if err != nil {
		logger.Error("failed to initialize crypter", "error", err)
		os.Exit(1)
	}

	storage, err := platform.NewStorage(initCtx, cfg.R2)
	if err != nil {
		logger.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}

	var mailer auth.Mailer = auth.NoopMailer{}
	resendKey := os.Getenv("RESEND_API_KEY")
	resendFrom := os.Getenv("RESEND_FROM_EMAIL")
	if resendKey != "" && resendFrom != "" {
		mailer = auth.NewResendMailer(resendKey, resendFrom)
	}

	var whatsapp notifications.WhatsAppSender = notifications.NoopWhatsAppSender{}
	if saungwaKey := os.Getenv("SAUNGWA_API_KEY"); saungwaKey != "" {
		client, err := notifications.NewSaungWAClient(
			saungwaKey,
			os.Getenv("SAUNGWA_ENDPOINT"),
			os.Getenv("APP_ENV") != "production",
		)
		if err != nil {
			logger.Error("failed to configure SaungWA client", "error", err)
		} else {
			whatsapp = client
		}
	}

	appBaseURL := os.Getenv("APP_URL")
	if appBaseURL == "" {
		appBaseURL = "http://localhost:3000"
	}

	authService := auth.NewService(pool, tokens, crypter, mailer, appBaseURL)
	authz := auth.NewAuthorizationService(pool)
	usersService := users.NewService(pool, mailer, appBaseURL)
	residentsService := residents.NewService(pool, crypter, cfg.DataEncryptionKey)
	invoicesService := invoices.NewService(pool)
	filesService := files.NewService(pool, storage)
	cashService := cash.NewService(pool)
	paymentsService := payments.NewService(pool, cashService)
	communicationService := communication.NewService(pool)
	lettersService := letters.NewService(pool)
	complaintsService := complaints.NewService(pool)
	notificationsService := notifications.NewService(pool)
	dashboardService := dashboard.NewService(pool)
	reportsService := reports.NewService(pool)
	settingsService := settings.NewService(pool)
	auditService := audit.NewService(pool)
	dispatcher := notifications.NewDispatcher(pool, notificationsService, mailer, whatsapp, logger)
	authService.SetNotificationDispatcher(dispatcher)
	usersService.SetNotificationDispatcher(dispatcher)
	paymentsService.SetNotificationDispatcher(dispatcher)
	invoicesService.SetNotificationDispatcher(dispatcher)
	residentsService.SetNotificationDispatcher(dispatcher)
	lettersService.SetNotificationDispatcher(dispatcher)
	complaintsService.SetNotificationDispatcher(dispatcher)
	communicationService.SetNotificationDispatcher(dispatcher)
	production := os.Getenv("APP_ENV") == "production"

	server := &http.Server{
		Addr:    cfg.Address(),
		Handler: httpapi.NewServer(logger, pool, tokens, authService, authz, usersService, residentsService, invoicesService, filesService, paymentsService, cashService, production, communicationService, lettersService, complaintsService, notificationsService, dashboardService, reportsService, settingsService, auditService, storage),
	}

	schedulerCtx, cancelScheduler := context.WithCancel(context.Background())
	defer cancelScheduler()
	go runEpic14Scheduler(schedulerCtx, logger, invoicesService, lettersService, complaintsService, residentsService)

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("starting server", "address", server.Addr)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case signal := <-signalContext():
		logger.Info("shutdown requested", "signal", signal)
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server exited cleanly")
}

func runEpic14Scheduler(
	ctx context.Context,
	logger *slog.Logger,
	invoiceService *invoices.Service,
	letterService *letters.Service,
	complaintService *complaints.Service,
	residentService *residents.Service,
) {
	run := func() {
		runCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()

		if result, err := invoiceService.RunScheduledGeneration(runCtx); err != nil {
			logger.Error("scheduled invoice generation failed", "error", err)
		} else {
			logger.Info("scheduled invoice generation completed",
				"attempted", result.Attempted,
				"created", result.Created,
				"skipped", result.Skipped,
				"failed", result.Failed,
			)
		}

		if reminders, err := invoiceService.RunDueReminders(runCtx); err != nil {
			logger.Error("scheduled due reminders failed", "error", err)
		} else {
			logger.Info("scheduled due reminders completed",
				"eligible", reminders.Eligible,
				"queued", reminders.Queued,
				"skipped", reminders.Skipped,
				"failures", reminders.Failures,
			)
		}

		if count, err := letterService.EscalateLetters(runCtx); err != nil {
			logger.Error("scheduled letter escalation failed", "error", err)
		} else if count > 0 {
			logger.Info("scheduled letter escalation completed", "escalated", count)
		}

		if count, err := complaintService.AutoCloseComplaints(runCtx); err != nil {
			logger.Error("scheduled complaint auto-close failed", "error", err)
		} else if count > 0 {
			logger.Info("scheduled complaint auto-close completed", "closed", count)
		}

		if count, err := residentService.RunDomicileReminders(runCtx); err != nil {
			logger.Error("scheduled domicile reminders failed", "error", err)
		} else if count > 0 {
			logger.Info("scheduled domicile reminders completed", "queued", count)
		}
	}

	run()
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

func signalContext() <-chan os.Signal {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	return signals
}
