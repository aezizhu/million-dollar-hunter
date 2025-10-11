package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

type PoolConfig struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

func NewPostgres(url string, poolCfg PoolConfig) (*Repo, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	cfg.MaxConns = poolCfg.MaxConns
	cfg.MinConns = poolCfg.MinConns
	cfg.MaxConnLifetime = poolCfg.MaxConnLifetime
	cfg.MaxConnIdleTime = poolCfg.MaxConnIdleTime
	cfg.HealthCheckPeriod = poolCfg.HealthCheckPeriod

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
	WalletID      string
	Assets        []Asset
	TotalUSDValue float64
}

type Asset struct {
	TokenAddress  string
	Symbol        string
	Name          string
	Amount        string
	USDValue      float64
	CurrentBalance string
}

type Transaction struct {
	Hash         string    `json:"hash"`
	From         string    `json:"from"`
	To           string    `json:"to"`
	Amount       string    `json:"amount"`
	Symbol       string    `json:"symbol"`
	TokenAddress string    `json:"token_address"`
	Timestamp    time.Time `json:"timestamp"`
	Type         string    `json:"type"`
}

type TransactionDataIngestedEvent struct {
	WalletAddress string        `json:"wallet_address"`
	Chain         string        `json:"chain"`
	Transactions  []Transaction `json:"transactions"`
}

func (r *Repo) UpsertFromIngest(ctx context.Context, payload []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if len(payload) == 0 {
		return errors.New("empty payload")
	}

	var event TransactionDataIngestedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	walletID, err := r.upsertWallet(ctx, tx, event.WalletAddress, event.Chain)
	if err != nil {
		return err
	}

	if err := r.processTransactions(ctx, tx, walletID, event.Transactions); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repo) upsertWallet(ctx context.Context, tx pgx.Tx, address, chain string) (uuid.UUID, error) {
	var walletID uuid.UUID
	err := tx.QueryRow(ctx, `
		INSERT INTO wallets (address, chain)
		VALUES ($1, $2)
		ON CONFLICT (address, chain) DO UPDATE SET address = EXCLUDED.address
		RETURNING id
	`, address, chain).Scan(&walletID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("upsert wallet: %w", err)
	}
	return walletID, nil
}

func (r *Repo) processTransactions(ctx context.Context, tx pgx.Tx, walletID uuid.UUID, transactions []Transaction) error {
	for _, txn := range transactions {
		if txn.Hash == "" {
			continue
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO transactions_view (wallet_id, ts, type, from_addr, to_addr, asset_symbol, amount, tx_hash)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (tx_hash) DO NOTHING
		`, walletID, txn.Timestamp, txn.Type, txn.From, txn.To, txn.Symbol, txn.Amount, txn.Hash)
		if err != nil {
			return fmt.Errorf("insert transaction: %w", err)
		}

		if txn.TokenAddress != "" {
			_, err = tx.Exec(ctx, `
				INSERT INTO assets (wallet_id, token_address, symbol, current_balance, updated_at)
				VALUES ($1, $2, $3, $4, now())
				ON CONFLICT (wallet_id, token_address) 
				DO UPDATE SET current_balance = EXCLUDED.current_balance, updated_at = now()
			`, walletID, txn.TokenAddress, txn.Symbol, txn.Amount)
			if err != nil {
				return fmt.Errorf("upsert asset: %w", err)
			}
		}
	}
	return nil
}

func (r *Repo) GetPortfolioByWalletID(ctx context.Context, walletID string) (*Portfolio, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	portfolio := &Portfolio{
		WalletID: walletID,
		Assets:   make([]Asset, 0),
	}

	rows, err := r.db.Query(ctx, `
		SELECT token_address, symbol, name, current_balance, COALESCE(usd_value, 0) as usd_value
		FROM assets a
		LEFT JOIN asset_snapshots s ON a.id = s.asset_id
		WHERE a.wallet_id = (SELECT id FROM wallets WHERE id = $1 OR address = $1 LIMIT 1)
		ORDER BY current_balance DESC
	`, walletID)
	if err != nil {
		return nil, fmt.Errorf("query assets: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var asset Asset
		if err := rows.Scan(&asset.TokenAddress, &asset.Symbol, &asset.Name, &asset.CurrentBalance, &asset.USDValue); err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		portfolio.Assets = append(portfolio.Assets, asset)
		portfolio.TotalUSDValue += asset.USDValue
	}

	return portfolio, rows.Err()
}
