package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL string
	HTTPAddr    string
	CORSOrigins []string
	JWTSecret   string
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		HTTPAddr:    strings.TrimSpace(os.Getenv("HTTP_ADDR")),
		CORSOrigins: splitCSV(os.Getenv("CORS_ORIGINS")),
		JWTSecret:   strings.TrimSpace(os.Getenv("JWT_SECRET")),
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = "127.0.0.1:8080"
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if len(cfg.CORSOrigins) == 0 {
		return Config{}, errors.New("CORS_ORIGINS is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must be at least 32 characters")
	}
	return cfg, nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
