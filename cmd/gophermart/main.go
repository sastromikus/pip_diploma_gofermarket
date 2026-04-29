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

	userRepo := repository.NewMemoryUserRepository()
	authService := service.NewAuthService(userRepo)
	h := handler.NewHandler(authService)

	log.Printf("starting server on %s", cfg.RunAddress)

	if err := http.ListenAndServe(cfg.RunAddress, handler.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}