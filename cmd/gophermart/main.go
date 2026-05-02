package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/accrual"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/config"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/handler"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/repository"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/service"
)

func main() {
	cfg := config.Load()

	if cfg.DatabaseURI == "" {
		log.Fatal("DATABASE_URI is empty")
	}

	if err := repository.RunMigrations(cfg.DatabaseURI, "migrations"); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	db, err := repository.NewPostgresDB(cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
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

	if cfg.AccrualSystemAddress != "" {
		accrualClient := accrual.NewClient(cfg.AccrualSystemAddress)
		accrualWorker := service.NewAccrualWorker(orderRepo, accrualClient)
		accrualWorker.Start(workerCtx)
	}

	h := handler.NewHandler(authService, orderService, balanceService)

	srv := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: handler.NewRouter(h),
	}

	serverCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("starting server on %s", cfg.RunAddress)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-serverCtx.Done()
	workerCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}

	log.Println("server stopped")
}
