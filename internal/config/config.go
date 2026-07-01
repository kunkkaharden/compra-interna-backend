package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port           string
	DBPath         string
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

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "compra_interna.db"
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
		DBPath:         dbPath,
		JWTSecret:      secret,
		JWTExpiryHours: expiryHours,
	}, nil
}
