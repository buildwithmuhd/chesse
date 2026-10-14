package auth

import (
	"context"
	"errors"
	"net/mail"
	"strings"
)

type Service struct {
	store  Store
	tokens *TokenManager
}

func NewService(store Store, tokens *TokenManager) *Service {
	return &Service{store: store, tokens: tokens}
}

func (s *Service) Register(ctx context.Context, email, password string) (Session, error) {
	email, err := cleanEmail(email)
	if err != nil {
		return Session{}, err
	}
	if len(password) < 8 || len(password) > 256 {
		return Session{}, errors.New("password must be between 8 and 256 characters")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return Session{}, err
	}
	user, err := s.store.CreateUser(ctx, email, hash)
	if err != nil {
		return Session{}, err
	}
	return s.issueSession(ctx, user, "", "")
}

func (s *Service) Login(ctx context.Context, email, password string) (Session, error) {
	email, err := cleanEmail(email)
	if err != nil {
		return Session{}, ErrInvalidCredentials
	}
	user, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		return Session{}, ErrInvalidCredentials
	}
	ok, err := VerifyPassword(password, user.PasswordHash)
	if err != nil || !ok {
		return Session{}, ErrInvalidCredentials
	}
	return s.issueSession(ctx, user, "", "")
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (Session, error) {
	current, err := s.store.FindRefreshToken(ctx, HashRefreshToken(refreshToken))
	if err != nil {
		return Session{}, ErrInvalidRefreshToken
	}
	if current.RevokedAt != nil || !s.tokens.now().Before(current.ExpiresAt) {
		return Session{}, ErrInvalidRefreshToken
	}
	if current.RotatedAt != nil {
		_ = s.store.RevokeRefreshFamily(ctx, current.FamilyID)
		return Session{}, ErrInvalidRefreshToken
	}
	user, err := s.store.FindUserByID(ctx, current.UserID)
	if err != nil {
		return Session{}, ErrInvalidRefreshToken
	}
	return s.issueSession(ctx, user, current.FamilyID, current.ID)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return s.store.RevokeRefreshToken(ctx, HashRefreshToken(refreshToken))
}

func (s *Service) issueSession(ctx context.Context, user User, familyID, previousRefreshID string) (Session, error) {
	access, err := s.tokens.NewAccessToken(user)
	if err != nil {
		return Session{}, err
	}
	refresh, err := s.tokens.NewRefreshToken()
	if err != nil {
		return Session{}, err
	}
	id, err := NewID()
	if err != nil {
		return Session{}, err
	}
	if familyID == "" {
		familyID, err = NewID()
		if err != nil {
			return Session{}, err
		}
	}
	next := RefreshToken{ID: id, FamilyID: familyID, UserID: user.ID, TokenHash: HashRefreshToken(refresh), ExpiresAt: s.tokens.RefreshExpiry()}
	if previousRefreshID != "" {
		if err := s.store.RotateRefreshToken(ctx, previousRefreshID, next); err != nil {
			return Session{}, err
		}
	} else if err := s.store.CreateRefreshToken(ctx, next); err != nil {
		return Session{}, err
	}
	return Session{User: user, AccessToken: access, RefreshToken: refresh}, nil
}

func cleanEmail(value string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))
	if len(email) > 320 {
		return "", errors.New("email is too long")
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return "", errors.New("invalid email")
	}
	return email, nil
}
