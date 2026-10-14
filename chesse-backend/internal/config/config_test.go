package config

import (
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/chesse")
	t.Setenv("CORS_ORIGINS", "http://127.0.0.1:5173, https://chesse.example")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPAddr != "127.0.0.1:8080" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	wantOrigins := []string{"http://127.0.0.1:5173", "https://chesse.example"}
	if !reflect.DeepEqual(cfg.CORSOrigins, wantOrigins) {
		t.Fatalf("CORSOrigins = %#v, want %#v", cfg.CORSOrigins, wantOrigins)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("CORS_ORIGINS", "http://127.0.0.1:5173")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestLoadRequiresCORSOrigins(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/chesse")
	t.Setenv("CORS_ORIGINS", " ")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestLoadRequiresJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/chesse")
	t.Setenv("CORS_ORIGINS", "http://127.0.0.1:5173")
	t.Setenv("JWT_SECRET", "short")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}
