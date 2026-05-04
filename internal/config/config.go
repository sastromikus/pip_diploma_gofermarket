package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	AuthSecret           string
}

func Load() Config {
	cfg := Config{
		RunAddress:           getEnv("RUN_ADDRESS", "localhost:8080"),
		DatabaseURI:          getEnv("DATABASE_URI", ""),
		AccrualSystemAddress: getEnv("ACCRUAL_SYSTEM_ADDRESS", ""),
		AuthSecret:           getEnv("AUTH_SECRET", "dev-secret"),
	}

	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "server run address")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "database connection URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "accrual system address")
	flag.StringVar(&cfg.AuthSecret, "s", cfg.AuthSecret, "auth secret")
	flag.Parse()

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
