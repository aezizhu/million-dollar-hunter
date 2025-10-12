package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type RefreshToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

type RefreshStore interface {
	CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) (RefreshToken, error)
	GetValidRefreshToken(ctx context.Context, token string, now time.Time) (RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

func (s *PGStore) CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) (RefreshToken, error) {
	const q = `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1,$2,$3)
	           RETURNING id, user_id, token, expires_at, created_at, revoked_at`
	var rt RefreshToken
	err := s.Pool.QueryRow(ctx, q, userID, token, expiresAt).Scan(
		&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt, &rt.RevokedAt,
	)
	return rt, err
}

func (s *PGStore) GetValidRefreshToken(ctx context.Context, token string, now time.Time) (RefreshToken, error) {
	const q = `SELECT id, user_id, token, expires_at, created_at, revoked_at
	           FROM refresh_tokens
	           WHERE token=$1 AND (revoked_at IS NULL) AND expires_at > $2`
	var rt RefreshToken
	err := s.Pool.QueryRow(ctx, q, token, now).Scan(
		&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt, &rt.RevokedAt,
	)
	if err != nil {
		return RefreshToken{}, err
	}
	return rt, nil
}

func (s *PGStore) RevokeRefreshToken(ctx context.Context, token string) error {
	const q = `UPDATE refresh_tokens SET revoked_at=NOW() WHERE token=$1 AND revoked_at IS NULL`
	ct, err := s.Pool.Exec(ctx, q, token)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *PGStore) RevokeAllForUser(ctx context.Context, userID string) error {
	const q = `UPDATE refresh_tokens SET revoked_at=NOW() WHERE user_id=$1 AND revoked_at IS NULL`
	_, err := s.Pool.Exec(ctx, q, userID)
	return err
}

func (s *PGStore) StoreRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	_, err := s.CreateRefreshToken(ctx, userID, token, expiresAt)
	return err
}

func (s *PGStore) ValidateRefreshToken(ctx context.Context, token string) (bool, string, error) {
	rt, err := s.GetValidRefreshToken(ctx, token, time.Now())
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, "", nil
		}
		return false, "", err
	}
	return true, rt.UserID, nil
}
