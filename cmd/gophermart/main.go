package main

import (
	"fmt"

	"github.com/sastromikus/pip_diploma_gofermarket/internal/config"
)

func main() {
	cfg := config.Load()

	fmt.Println("run address:", cfg.RunAddress)
	fmt.Println("database uri:", cfg.DatabaseURI)
	fmt.Println("accrual address:", cfg.AccrualSystemAddress)
}