package config

import (
	"flag"
	"net"
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
		RunAddress:           "localhost:8080",
		DatabaseURI:          "",
		AccrualSystemAddress: "",
		AuthSecret:           "dev-secret",
	}

	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "server run address")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "database connection URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "accrual system address")
	flag.StringVar(&cfg.AuthSecret, "s", cfg.AuthSecret, "auth secret")
	flag.Parse()

	// Practicum checks expect environment variables to override command-line flags.
	// This also keeps the service compatible with both local runs and CI runs.
	if v := os.Getenv("RUN_ADDRESS"); v != "" {
		cfg.RunAddress = v
	}
	if v := os.Getenv("DATABASE_URI"); v != "" {
		cfg.DatabaseURI = v
	}
	if v := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); v != "" {
		cfg.AccrualSystemAddress = v
	}
	if v := os.Getenv("AUTH_SECRET"); v != "" {
		cfg.AuthSecret = v
	}

	cfg.RunAddress = normalizeRunAddress(cfg.RunAddress)

	return cfg
}

func normalizeRunAddress(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}

	// On Windows, listening on "localhost:port" can bind only IPv4 while the
	// test client resolves localhost to ::1. Listening on ":port" accepts both.
	if host == "localhost" || host == "127.0.0.1" || host == "::1" || host == "[::1]" {
		return ":" + port
	}

	return addr
}
