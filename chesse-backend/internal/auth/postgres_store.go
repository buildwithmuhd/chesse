package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) CreateUser(ctx context.Context, email, passwordHash string) (User, error) {
	var user User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, password_hash`,
		email, passwordHash,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if isUniqueViolation(err) {
		return User{}, ErrEmailTaken
	}
	return user, err
}

func (s *PostgresStore) FindUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, password_hash FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)
	return user, err
}

func (s *PostgresStore) FindUserByID(ctx context.Context, id int64) (User, error) {
	var user User
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, password_hash FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)
	return user, err
}

func (s *PostgresStore) CreateRefreshToken(ctx context.Context, token RefreshToken) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO refresh_token_families (id, user_id) VALUES ($1, $2)`,
		token.FamilyID, token.UserID,
	)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO refresh_tokens (id, family_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)`,
		token.ID, token.FamilyID, token.TokenHash, token.ExpiresAt,
	)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *PostgresStore) FindRefreshToken(ctx context.Context, tokenHash string) (RefreshToken, error) {
	var token RefreshToken
	err := s.pool.QueryRow(ctx, `
		SELECT rt.id, rt.family_id, rtf.user_id, rt.token_hash, rt.expires_at, rt.rotated_at, rtf.revoked_at
		FROM refresh_tokens rt
		JOIN refresh_token_families rtf ON rtf.id = rt.family_id
		WHERE rt.token_hash = $1`,
		tokenHash,
	).Scan(&token.ID, &token.FamilyID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.RotatedAt, &token.RevokedAt)
	return token, err
}

func (s *PostgresStore) RotateRefreshToken(ctx context.Context, oldTokenID string, next RefreshToken) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var rotatedAt *time.Time
	err = tx.QueryRow(ctx, `SELECT rotated_at FROM refresh_tokens WHERE id = $1 FOR UPDATE`, oldTokenID).Scan(&rotatedAt)
	if err != nil {
		return err
	}
	if rotatedAt != nil {
		_, _ = tx.Exec(ctx, `UPDATE refresh_token_families SET revoked_at = now() WHERE id = $1`, next.FamilyID)
		return ErrInvalidRefreshToken
	}
	if _, err := tx.Exec(ctx, `UPDATE refresh_tokens SET rotated_at = now() WHERE id = $1`, oldTokenID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO refresh_tokens (id, family_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)`,
		next.ID, next.FamilyID, next.TokenHash, next.ExpiresAt,
	); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *PostgresStore) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE refresh_token_families
		SET revoked_at = now()
		WHERE id = (SELECT family_id FROM refresh_tokens WHERE token_hash = $1)`,
		tokenHash,
	)
	return err
}

func (s *PostgresStore) RevokeRefreshFamily(ctx context.Context, familyID string) error {
	_, err := s.pool.Exec(ctx, `UPDATE refresh_token_families SET revoked_at = now() WHERE id = $1`, familyID)
	return err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
