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

type WalletSummary struct {
	ID            string
	Address       string
	Chain         string
	TotalUSDValue float64
	AssetCount    int32
}

func (r *Repo) GetPortfolioSummary(ctx context.Context, userID string) ([]WalletSummary, float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.db.Query(ctx, `
		SELECT 
			w.id,
			w.address,
			w.chain,
			COUNT(a.id) as asset_count,
			COALESCE(SUM(s.usd_value), 0) as total_usd_value
		FROM wallets w
		LEFT JOIN assets a ON w.id = a.wallet_id
		LEFT JOIN asset_snapshots s ON a.id = s.asset_id
		GROUP BY w.id, w.address, w.chain
		ORDER BY total_usd_value DESC
	`)
	if err != nil {
		return nil, 0, fmt.Errorf("query wallets: %w", err)
	}
	defer rows.Close()

	wallets := make([]WalletSummary, 0)
	var totalNetWorth float64

	for rows.Next() {
		var w WalletSummary
		if err := rows.Scan(&w.ID, &w.Address, &w.Chain, &w.AssetCount, &w.TotalUSDValue); err != nil {
			return nil, 0, fmt.Errorf("scan wallet: %w", err)
		}
		wallets = append(wallets, w)
		totalNetWorth += w.TotalUSDValue
	}

	return wallets, totalNetWorth, rows.Err()
}

type WalletDetails struct {
	WalletID      string
	Address       string
	Chain         string
	Assets        []Asset
	TotalUSDValue float64
}

func (r *Repo) GetWalletDetails(ctx context.Context, walletID, address string) (*WalletDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var details WalletDetails
	var walletUUID string

	query := `SELECT id, address, chain FROM wallets WHERE `
	var args []interface{}
	if walletID != "" {
		query += "id = $1"
		args = append(args, walletID)
	} else if address != "" {
		query += "address = $1"
		args = append(args, address)
	} else {
		return nil, errors.New("wallet_id or address required")
	}
	query += " LIMIT 1"

	err := r.db.QueryRow(ctx, query, args...).Scan(&walletUUID, &details.Address, &details.Chain)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("wallet not found")
		}
		return nil, fmt.Errorf("query wallet: %w", err)
	}
	details.WalletID = walletUUID

	rows, err := r.db.Query(ctx, `
		SELECT token_address, symbol, name, current_balance, COALESCE(s.usd_value, 0) as usd_value
		FROM assets a
		LEFT JOIN asset_snapshots s ON a.id = s.asset_id
		WHERE a.wallet_id = $1
		ORDER BY current_balance DESC
	`, walletUUID)
	if err != nil {
		return nil, fmt.Errorf("query assets: %w", err)
	}
	defer rows.Close()

	details.Assets = make([]Asset, 0)
	for rows.Next() {
		var asset Asset
		if err := rows.Scan(&asset.TokenAddress, &asset.Symbol, &asset.Name, &asset.CurrentBalance, &asset.USDValue); err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		details.Assets = append(details.Assets, asset)
		details.TotalUSDValue += asset.USDValue
	}

	return &details, rows.Err()
}

type TransactionResult struct {
	Transactions []Transaction
	TotalCount   int32
}

func (r *Repo) GetTransactionHistory(ctx context.Context, walletID, address string, page, limit int32, filterByType string) (*TransactionResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var walletUUID string
	whereClause := `WHERE w.id = $1 OR w.address = $1`
	if walletID != "" {
		walletUUID = walletID
	} else if address != "" {
		walletUUID = address
	} else {
		return nil, errors.New("wallet_id or address required")
	}

	var typeFilter string
	var args []interface{}
	args = append(args, walletUUID)
	argPos := 2

	if filterByType != "" {
		typeFilter = fmt.Sprintf(" AND t.type = $%d", argPos)
		args = append(args, filterByType)
		argPos++
	}

	var totalCount int32
	countQuery := `
		SELECT COUNT(*)
		FROM transactions_view t
		JOIN wallets w ON t.wallet_id = w.id
		` + whereClause + typeFilter
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count transactions: %w", err)
	}

	args = append(args, limit, offset)
	query := `
		SELECT t.tx_hash, t.from_addr, t.to_addr, t.amount, t.asset_symbol, COALESCE(a.token_address, ''), t.ts, t.type
		FROM transactions_view t
		JOIN wallets w ON t.wallet_id = w.id
		LEFT JOIN assets a ON a.wallet_id = t.wallet_id AND a.symbol = t.asset_symbol
		` + whereClause + typeFilter + `
		ORDER BY t.ts DESC
		LIMIT $` + fmt.Sprintf("%d", argPos) + ` OFFSET $` + fmt.Sprintf("%d", argPos+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]Transaction, 0)
	for rows.Next() {
		var txn Transaction
		if err := rows.Scan(&txn.Hash, &txn.From, &txn.To, &txn.Amount, &txn.Symbol, &txn.TokenAddress, &txn.Timestamp, &txn.Type); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		transactions = append(transactions, txn)
	}

	return &TransactionResult{
		Transactions: transactions,
		TotalCount:   totalCount,
	}, rows.Err()
}
