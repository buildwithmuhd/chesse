package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPasswordHashVerifiesOnlyOriginalPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{name: "match", password: "correct horse battery staple", want: true},
		{name: "mismatch", password: "wrong password", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifyPassword(tt.password, hash)
			if err != nil {
				t.Fatalf("VerifyPassword() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("VerifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccessTokenRejectsExpiredToken(t *testing.T) {
	tokens := NewTokenManager([]byte("0123456789abcdef0123456789abcdef"), time.Minute, time.Hour)
	now := time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)
	tokens.now = func() time.Time { return now }

	token, err := tokens.NewAccessToken(User{ID: 1, Email: "a@example.com"})
	if err != nil {
		t.Fatalf("NewAccessToken() error = %v", err)
	}
	if claims, err := tokens.ParseAccessToken(token); err != nil || claims.UserID != 1 {
		t.Fatalf("ParseAccessToken() claims = %#v, err = %v", claims, err)
	}

	tokens.now = func() time.Time { return now.Add(2 * time.Minute) }
	if _, err := tokens.ParseAccessToken(token); err == nil {
		t.Fatal("ParseAccessToken() error = nil, want expiry error")
	}
}

func TestRefreshRotatesTokenAndRevokesFamilyOnReuse(t *testing.T) {
	ctx := context.Background()
	store := newMemoryStore()
	tokens := NewTokenManager([]byte("0123456789abcdef0123456789abcdef"), time.Minute, time.Hour)
	service := NewService(store, tokens)

	session, err := service.Register(ctx, "PLAYER@example.com", "strong-password")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if session.User.Email != "player@example.com" {
		t.Fatalf("email = %q", session.User.Email)
	}

	next, err := service.Refresh(ctx, session.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if next.RefreshToken == session.RefreshToken {
		t.Fatal("Refresh() reused refresh token")
	}

	if _, err := service.Refresh(ctx, session.RefreshToken); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("Refresh(reused) error = %v, want %v", err, ErrInvalidRefreshToken)
	}
	nextHash := HashRefreshToken(next.RefreshToken)
	nextStored, err := store.FindRefreshToken(ctx, nextHash)
	if err != nil {
		t.Fatalf("FindRefreshToken(next) error = %v", err)
	}
	if nextStored.RevokedAt == nil {
		t.Fatal("reused refresh token did not revoke token family")
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute)
	now := time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }

	if !limiter.Allow("login:127.0.0.1") || !limiter.Allow("login:127.0.0.1") {
		t.Fatal("first two attempts should pass")
	}
	if limiter.Allow("login:127.0.0.1") {
		t.Fatal("third attempt should be rate limited")
	}
	now = now.Add(time.Minute)
	if !limiter.Allow("login:127.0.0.1") {
		t.Fatal("attempt after window should pass")
	}
}

type memoryStore struct {
	nextUserID int64
	users     map[int64]User
	byEmail   map[string]int64
	refresh   map[string]RefreshToken
}

func newMemoryStore() *memoryStore {
	return &memoryStore{nextUserID: 1, users: map[int64]User{}, byEmail: map[string]int64{}, refresh: map[string]RefreshToken{}}
}

func (s *memoryStore) CreateUser(_ context.Context, email, passwordHash string) (User, error) {
	if _, ok := s.byEmail[email]; ok {
		return User{}, ErrEmailTaken
	}
	user := User{ID: s.nextUserID, Email: email, PasswordHash: passwordHash}
	s.nextUserID++
	s.users[user.ID] = user
	s.byEmail[email] = user.ID
	return user, nil
}

func (s *memoryStore) FindUserByEmail(_ context.Context, email string) (User, error) {
	id, ok := s.byEmail[email]
	if !ok {
		return User{}, ErrInvalidCredentials
	}
	return s.users[id], nil
}

func (s *memoryStore) FindUserByID(_ context.Context, id int64) (User, error) {
	user, ok := s.users[id]
	if !ok {
		return User{}, ErrInvalidCredentials
	}
	return user, nil
}

func (s *memoryStore) CreateRefreshToken(_ context.Context, token RefreshToken) error {
	s.refresh[token.TokenHash] = token
	return nil
}

func (s *memoryStore) FindRefreshToken(_ context.Context, tokenHash string) (RefreshToken, error) {
	token, ok := s.refresh[tokenHash]
	if !ok {
		return RefreshToken{}, ErrInvalidRefreshToken
	}
	return token, nil
}

func (s *memoryStore) RotateRefreshToken(_ context.Context, oldTokenID string, next RefreshToken) error {
	var oldHash string
	var old RefreshToken
	for hash, token := range s.refresh {
		if token.ID == oldTokenID {
			oldHash = hash
			old = token
			break
		}
	}
	if oldHash == "" {
		return ErrInvalidRefreshToken
	}
	if old.RotatedAt != nil {
		return ErrInvalidRefreshToken
	}
	now := time.Now()
	old.RotatedAt = &now
	s.refresh[oldHash] = old
	s.refresh[next.TokenHash] = next
	return nil
}

func (s *memoryStore) RevokeRefreshToken(_ context.Context, tokenHash string) error {
	token, ok := s.refresh[tokenHash]
	if !ok {
		return nil
	}
	return s.RevokeRefreshFamily(context.Background(), token.FamilyID)
}

func (s *memoryStore) RevokeRefreshFamily(_ context.Context, familyID string) error {
	now := time.Now()
	for hash, token := range s.refresh {
		if token.FamilyID == familyID {
			token.RevokedAt = &now
			s.refresh[hash] = token
		}
	}
	return nil
}
