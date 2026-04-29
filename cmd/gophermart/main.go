package main

import (
	"log"
	"net/http"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/config"
	"github.com/sastromikus/pip_diploma_gofermarket/internal/handler"
)

func main() {
	cfg := config.Load()

	log.Printf("starting server on %s", cfg.RunAddress)

	if err := http.ListenAndServe(cfg.RunAddress, handler.NewRouter()); err != nil {
		log.Fatal(err)
	}
}