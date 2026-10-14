package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	NewHandlerWithAuth(nil, []string{"http://127.0.0.1:5173"}, nil, nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestCORSAllowsConfiguredOriginOnly(t *testing.T) {
	tests := []struct {
		name       string
		origin     string
		wantOrigin string
	}{
		{name: "allowed", origin: "http://127.0.0.1:5173", wantOrigin: "http://127.0.0.1:5173"},
		{name: "blocked", origin: "https://evil.example", wantOrigin: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			req.Header.Set("Origin", tt.origin)
			rec := httptest.NewRecorder()

			NewHandlerWithAuth(nil, []string{"http://127.0.0.1:5173"}, nil, nil).ServeHTTP(rec, req)

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.wantOrigin {
				t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantOrigin)
			}
		})
	}
}
