package server

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"time"

	"nochess/chesse-backend/internal/auth"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewHandler(pool *pgxpool.Pool, corsOrigins []string, jwtSecret []byte) http.Handler {
	store := auth.NewPostgresStore(pool)
	tokens := auth.NewTokenManager(jwtSecret, 15*time.Minute, 30*24*time.Hour)
	service := auth.NewService(store, tokens)
	return NewHandlerWithAuth(pool, corsOrigins, service, tokens)
}

func NewHandlerWithAuth(pool *pgxpool.Pool, corsOrigins []string, authService *auth.Service, tokens *auth.TokenManager) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", jsonStatus(http.StatusOK, "ok"))
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if pool == nil || pool.Ping(ctx) != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unready"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	authHandlers := auth.NewHandlers(authService, tokens)
	mux.HandleFunc("POST /auth/register", authHandlers.Register)
	mux.HandleFunc("POST /auth/login", authHandlers.Login)
	mux.HandleFunc("POST /auth/refresh", authHandlers.Refresh)
	mux.HandleFunc("POST /auth/logout", authHandlers.Logout)
	mux.HandleFunc("GET /me", authHandlers.RequireAuth(authHandlers.Me))
	return cors(corsOrigins, mux)
}

func jsonStatus(code int, status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, code, map[string]string{"status": status})
	}
}

func writeJSON(w http.ResponseWriter, code int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(value)
}

func cors(allowedOrigins []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && slices.Contains(allowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
