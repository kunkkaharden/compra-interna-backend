package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	JWTExpiryHours int
}

func Load() (*Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET env var is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL env var is required")
	}

	expiryHours := 24
	if raw := os.Getenv("JWT_EXPIRY_HOURS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("JWT_EXPIRY_HOURS must be an integer: %w", err)
		}
		expiryHours = parsed
	}

	return &Config{
		Port:           port,
		DatabaseURL:    databaseURL,
		JWTSecret:      secret,
		JWTExpiryHours: expiryHours,
	}, nil
}
