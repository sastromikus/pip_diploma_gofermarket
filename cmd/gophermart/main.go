package main

import (
	"log"
	"net/http"

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
	authService := service.NewAuthService(userRepo)
	orderRepo := repository.NewPostgresOrderRepository(db)
	orderService := service.NewOrderService(orderRepo)

	h := handler.NewHandler(authService, orderService)

	log.Printf("starting server on %s", cfg.RunAddress)

	if err := http.ListenAndServe(cfg.RunAddress, handler.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}
