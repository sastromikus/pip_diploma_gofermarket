package config

import (
	"flag"
	"net"
	"os"
	"strings"
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

	applyEnv(&cfg)

	cfg.RunAddress = normalizeRunAddress(cfg.RunAddress)
	cfg.AccrualSystemAddress = normalizeHTTPAddress(cfg.AccrualSystemAddress)

	return cfg
}

func applyEnv(cfg *Config) {
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
}

func normalizeRunAddress(address string) string {
	address = strings.TrimSpace(address)
	if address == "" {
		return address
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return address
	}

	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return ":" + port
	}

	return address
}

func normalizeHTTPAddress(address string) string {
	address = strings.TrimSpace(address)
	if address == "" {
		return address
	}

	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		address = "http://" + address
	}

	// On Windows localhost often resolves to ::1 first. The test binaries and
	// local servers are commonly bound to IPv4, so prefer 127.0.0.1 for loopback.
	address = strings.Replace(address, "http://localhost:", "http://127.0.0.1:", 1)
	address = strings.Replace(address, "https://localhost:", "https://127.0.0.1:", 1)

	return strings.TrimRight(address, "/")
}
