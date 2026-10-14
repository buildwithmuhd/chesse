package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrEmailTaken          = errors.New("email already registered")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRateLimited         = errors.New("rate limited")
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
}

type RefreshToken struct {
	ID        string
	FamilyID  string
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	RotatedAt *time.Time
	RevokedAt *time.Time
}

type Store interface {
	CreateUser(ctx context.Context, email, passwordHash string) (User, error)
	FindUserByEmail(ctx context.Context, email string) (User, error)
	FindUserByID(ctx context.Context, id int64) (User, error)
	CreateRefreshToken(ctx context.Context, token RefreshToken) error
	FindRefreshToken(ctx context.Context, tokenHash string) (RefreshToken, error)
	RotateRefreshToken(ctx context.Context, oldTokenID string, next RefreshToken) error
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeRefreshFamily(ctx context.Context, familyID string) error
}

type Session struct {
	User         User
	AccessToken  string
	RefreshToken string
}
