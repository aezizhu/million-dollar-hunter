package config

type Config struct {
	GRPCAddr            string `env:"GRPC_ADDR,notEmpty" envDefault:":8081"`
	DatabaseURL         string `env:"DATABASE_URL,notEmpty" envDefault:"postgres://portfolio:portfolio@localhost:5432/portfolio?sslmode=disable"`
	KafkaBrokers        string `env:"KAFKA_BROKERS,notEmpty" envDefault:"localhost:9092"`
	TopicTxIngested     string `env:"TOPIC_TRANSACTION_INGESTED,notEmpty" envDefault:"TransactionDataIngested"`
	TopicPortfolioUpdated string `env:"TOPIC_PORTFOLIO_UPDATED,notEmpty" envDefault:"PortfolioUpdated"`
	GroupID             string `env:"KAFKA_GROUP_ID,notEmpty" envDefault:"portfolio-service"`
	ExportDir           string `env:"EXPORT_DIR,notEmpty" envDefault:"/data/exports"`
}
</create_file_file>
<create_file path="/home/ubuntu/repos/million-dollar-hunter/services/portfolio-service/internal/repository/postgres.go">package repository

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
