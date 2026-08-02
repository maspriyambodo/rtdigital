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

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
	"github.com/maspriyambodo/rtdigital/services/api/internal/cash"
	"github.com/maspriyambodo/rtdigital/services/api/internal/config"
	"github.com/maspriyambodo/rtdigital/services/api/internal/files"
	"github.com/maspriyambodo/rtdigital/services/api/internal/httpapi"
	"github.com/maspriyambodo/rtdigital/services/api/internal/invoices"
	"github.com/maspriyambodo/rtdigital/services/api/internal/payments"
	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
	"github.com/maspriyambodo/rtdigital/services/api/internal/residents"
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
	production := os.Getenv("APP_ENV") == "production"

	server := &http.Server{
		Addr:    cfg.Address(),
		Handler: httpapi.NewServer(logger, pool, tokens, authService, authz, usersService, residentsService, invoicesService, filesService, paymentsService, cashService, production),
	}

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

func signalContext() <-chan os.Signal {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	return signals
}
