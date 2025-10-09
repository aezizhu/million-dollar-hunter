package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewPostgres(url string) (*Repo, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	return &Repo{db: pool}, nil
}

func (r *Repo) Close(ctx context.Context) {
	r.db.Close()
}

type Portfolio struct {
	WalletID string
}

func (r *Repo) UpsertFromIngest(ctx context.Context, payload []byte) error {
	if len(payload) == 0 {
		return errors.New("empty payload")
	}
	return nil
}
