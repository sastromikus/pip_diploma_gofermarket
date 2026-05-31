package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/accrual"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/auth"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/config"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/handler"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/repository"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/service"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg := config.Load()

	if cfg.DatabaseURI == "" {
		logger.Error("DATABASE_URI is empty")
		os.Exit(1)
	}

	if err := repository.RunMigrations(cfg.DatabaseURI, "migrations"); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	db, err := repository.NewPostgresDB(cfg.DatabaseURI)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	userRepo := repository.NewPostgresUserRepository(db)
	orderRepo := repository.NewPostgresOrderRepository(db)
	balanceRepo := repository.NewPostgresBalanceRepository(db)
	authService := service.NewAuthService(userRepo)
	orderService := service.NewOrderService(orderRepo)
	balanceService := service.NewBalanceService(balanceRepo)

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	var accrualWorker *service.AccrualWorker

	if cfg.AccrualSystemAddress != "" {
		accrualClient := accrual.NewClient(cfg.AccrualSystemAddress)
		accrualWorker = service.NewAccrualWorker(
			orderRepo,
			accrualClient,
			logger.With("component", "accrual_worker"),
		)
		accrualWorker.Start(workerCtx)
	}

	authManager := auth.NewManager(cfg.AuthSecret)
	h := handler.NewHandler(authService, orderService, balanceService, authManager)

	srv := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: handler.NewRouter(h),
	}

	serverCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("starting server", "addr", cfg.RunAddress)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	select {
	case <-serverCtx.Done():
	case err := <-serverErr:
		if err != nil {
			logger.Error("server failed", "error", err)
			workerCancel()
			if accrualWorker != nil {
				accrualWorker.Wait()
			}
			os.Exit(1)
		}
	}

	workerCancel()

	if accrualWorker != nil {
		accrualWorker.Wait()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
	}

	logger.Info("server stopped")
}
