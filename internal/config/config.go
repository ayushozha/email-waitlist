package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port             int
	DatabaseURL      string
	AdminKey         string
	RateLimit        int  // requests per minute per IP
	TrustProxy       bool // trust X-Forwarded-For from a reverse proxy
	ResendAPIKey     string
	DefaultFromEmail string
}

func Load() (*Config, error) {
	port := 8090
	if v := os.Getenv("PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT: %w", err)
		}
		port = p
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	adminKey := os.Getenv("ADMIN_KEY")
	if adminKey == "" {
		return nil, fmt.Errorf("ADMIN_KEY is required")
	}

	rateLimit := 30
	if v := os.Getenv("RATE_LIMIT"); v != "" {
		r, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid RATE_LIMIT: %w", err)
		}
		rateLimit = r
	}

	trustProxy := false
	if v := os.Getenv("TRUST_PROXY"); v != "" {
		t, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("invalid TRUST_PROXY: %w", err)
		}
		trustProxy = t
	}

	resendKey := os.Getenv("RESEND_API_KEY")
	fromEmail := os.Getenv("DEFAULT_FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = "Waitlist <waitlist@ayushojha.com>"
	}

	return &Config{
		Port:             port,
		DatabaseURL:      dbURL,
		AdminKey:         adminKey,
		RateLimit:        rateLimit,
		TrustProxy:       trustProxy,
		ResendAPIKey:     resendKey,
		DefaultFromEmail: fromEmail,
	}, nil
}
