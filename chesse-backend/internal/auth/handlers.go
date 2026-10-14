package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const userContextKey contextKey = "auth_user"
const refreshCookieName = "chesse_refresh"

type Handlers struct {
	service *Service
	tokens  *TokenManager
	limit   *RateLimiter
}

func NewHandlers(service *Service, tokens *TokenManager) *Handlers {
	return &Handlers{service: service, tokens: tokens, limit: NewRateLimiter(5, time.Minute)}
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	if !h.allow(w, r, "register") {
		return
	}
	var req credentialsRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	session, err := h.service.Register(r.Context(), req.Email, req.Password)
	h.writeSession(w, session, err, http.StatusCreated)
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if !h.allow(w, r, "login") {
		return
	}
	var req credentialsRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	session, err := h.service.Login(r.Context(), req.Email, req.Password)
	h.writeSession(w, session, err, http.StatusOK)
}

func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	session, err := h.service.Refresh(r.Context(), cookie.Value)
	h.writeSession(w, session, err, http.StatusOK)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(refreshCookieName); err == nil {
		_ = h.service.Logout(r.Context(), cookie.Value)
	}
	clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(User)
	writeJSON(w, http.StatusOK, userResponse{ID: user.ID, Email: user.Email})
}

func (h *Handlers) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		claims, err := h.tokens.ParseAccessToken(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		user, err := h.service.store.FindUserByID(r.Context(), claims.UserID)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next(w, r.WithContext(ctx))
	}
}

func (h *Handlers) allow(w http.ResponseWriter, r *http.Request, action string) bool {
	if h.limit.Allow(action + ":" + clientIP(r)) {
		return true
	}
	writeError(w, http.StatusTooManyRequests, "rate limited")
	return false
}

func (h *Handlers) writeSession(w http.ResponseWriter, session Session, err error, successStatus int) {
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailTaken):
			writeError(w, http.StatusConflict, "email already registered")
		case errors.Is(err, ErrInvalidCredentials), errors.Is(err, ErrInvalidRefreshToken):
			writeError(w, http.StatusUnauthorized, "unauthorized")
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	setRefreshCookie(w, session.RefreshToken, h.tokens.refreshTTL)
	writeJSON(w, successStatus, sessionResponse{
		AccessToken: session.AccessToken,
		User:        userResponse{ID: session.User.ID, Email: session.User.Email},
	})
}

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type sessionResponse struct {
	AccessToken string       `json:"accessToken"`
	User        userResponse `json:"user"`
}

type userResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, code int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]string{"error": message})
}

func setRefreshCookie(w http.ResponseWriter, token string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/auth",
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}
