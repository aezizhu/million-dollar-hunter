package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserStore interface {
	Create(ctx context.Context, email, passwordHash string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
}

type PGStore struct {
	Pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *PGStore {
	return &PGStore{Pool: pool}
}

func (s *PGStore) Create(ctx context.Context, email, passwordHash string) (User, error) {
	const q = `INSERT INTO users (email, password_hash) VALUES ($1,$2) RETURNING id, email, password_hash, created_at, updated_at`
	var u User
	err := s.Pool.QueryRow(ctx, q, email, passwordHash).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (s *PGStore) GetByEmail(ctx context.Context, email string) (User, error) {
	const q = `SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email=$1`
	var u User
	err := s.Pool.QueryRow(ctx, q, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

var ErrNotImplemented = errors.New("not implemented")
