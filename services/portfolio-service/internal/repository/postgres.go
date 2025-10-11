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

type PgxPool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Close()
}

type Repo struct {
	db PgxPool
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
type OwnershipCheckResult struct {
	Owned bool
}

func (r *Repo) UserOwnsWallet(ctx context.Context, userID, walletIDOrAddr string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var owned bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM wallets
			WHERE user_id = $1
			  AND (id = $2 OR address = $2)
			LIMIT 1
		)
	`, userID, walletIDOrAddr).Scan(&owned)
	if err != nil {
		return false, fmt.Errorf("check ownership: %w", err)
	}
	return owned, nil
}


func (r *Repo) Close(ctx context.Context) {
	r.db.Close()
}

func (r *Repo) VerifyWalletOwnership(ctx context.Context, userID, walletID string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM wallets 
			WHERE (id = $1 OR address = $1) AND user_id = $2
		)
	`, walletID, userID).Scan(&exists)
	
	if err != nil {
		return fmt.Errorf("check wallet ownership: %w", err)
	}
	
	if !exists {
		return errors.New("wallet not found or access denied")
	}
	
	return nil
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
			INSERT INTO transactions_view (wallet_id, ts, type, from_addr, to_addr, asset_symbol, amount, tx_hash, token_address)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (tx_hash) DO NOTHING
		`, walletID, txn.Timestamp, txn.Type, txn.From, txn.To, txn.Symbol, txn.Amount, txn.Hash, txn.TokenAddress)
		if err != nil {
			return fmt.Errorf("insert transaction: %w", err)
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
		SELECT 
			t.token_address,
			t.asset_symbol as symbol,
			SUM(CASE 
				WHEN t.type IN ('receive', 'mint') THEN CAST(t.amount AS NUMERIC)
				WHEN t.type IN ('send', 'burn', 'approve') THEN -CAST(t.amount AS NUMERIC)
				ELSE 0
			END) as current_balance,
			COALESCE(MAX(s.usd_value), 0) as usd_value
		FROM transactions_view t
		LEFT JOIN LATERAL (
			SELECT usd_value FROM asset_snapshots 
			WHERE wallet_id = t.wallet_id AND token_address = t.token_address
			ORDER BY ts DESC 
			LIMIT 1
		) s ON true
		WHERE t.wallet_id = (SELECT id FROM wallets WHERE id = $1 OR address = $1 LIMIT 1)
			AND t.token_address IS NOT NULL 
			AND t.token_address != ''
		GROUP BY t.token_address, t.asset_symbol
		HAVING SUM(CASE 
			WHEN t.type IN ('receive', 'mint') THEN CAST(t.amount AS NUMERIC)
			WHEN t.type IN ('send', 'burn', 'approve') THEN -CAST(t.amount AS NUMERIC)
			ELSE 0
		END) > 0
		ORDER BY usd_value DESC, current_balance DESC
	`, walletID)
	if err != nil {
		return nil, fmt.Errorf("query assets: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var asset Asset
		var balance float64
		if err := rows.Scan(&asset.TokenAddress, &asset.Symbol, &balance, &asset.USDValue); err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		asset.CurrentBalance = fmt.Sprintf("%.18f", balance)
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
			COUNT(DISTINCT subq.token_address) as asset_count,
			COALESCE(SUM(subq.usd_value), 0) as total_usd_value
		FROM wallets w
		LEFT JOIN (
			SELECT DISTINCT ON (t.wallet_id, t.token_address)
				t.wallet_id,
				t.token_address,
				COALESCE(s.usd_value, 0) as usd_value
			FROM transactions_view t
			LEFT JOIN LATERAL (
				SELECT usd_value FROM asset_snapshots 
				WHERE wallet_id = t.wallet_id AND token_address = t.token_address
				ORDER BY ts DESC 
				LIMIT 1
			) s ON true
			WHERE t.token_address IS NOT NULL AND t.token_address != ''
			GROUP BY t.wallet_id, t.token_address, s.usd_value
			HAVING SUM(CASE 
				WHEN t.type IN ('receive', 'mint') THEN CAST(t.amount AS NUMERIC)
				WHEN t.type IN ('send', 'burn', 'approve') THEN -CAST(t.amount AS NUMERIC)
				ELSE 0
			END) > 0
		) subq ON w.id = subq.wallet_id
		WHERE w.user_id = $1
		GROUP BY w.id, w.address, w.chain
		ORDER BY total_usd_value DESC
	`, userID)
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
		SELECT 
			t.token_address,
			t.asset_symbol as symbol,
			SUM(CASE 
				WHEN t.type IN ('receive', 'mint') THEN CAST(t.amount AS NUMERIC)
				WHEN t.type IN ('send', 'burn', 'approve') THEN -CAST(t.amount AS NUMERIC)
				ELSE 0
			END) as current_balance,
			COALESCE(MAX(s.usd_value), 0) as usd_value
		FROM transactions_view t
		LEFT JOIN LATERAL (
			SELECT usd_value FROM asset_snapshots 
			WHERE wallet_id = t.wallet_id AND token_address = t.token_address
			ORDER BY ts DESC 
			LIMIT 1
		) s ON true
		WHERE t.wallet_id = $1
			AND t.token_address IS NOT NULL 
			AND t.token_address != ''
		GROUP BY t.token_address, t.asset_symbol
		HAVING SUM(CASE 
			WHEN t.type IN ('receive', 'mint') THEN CAST(t.amount AS NUMERIC)
			WHEN t.type IN ('send', 'burn', 'approve') THEN -CAST(t.amount AS NUMERIC)
			ELSE 0
		END) > 0
		ORDER BY usd_value DESC, current_balance DESC
	`, walletUUID)
	if err != nil {
		return nil, fmt.Errorf("query assets: %w", err)
	}
	defer rows.Close()

	details.Assets = make([]Asset, 0)
	for rows.Next() {
		var asset Asset
		var balance float64
		if err := rows.Scan(&asset.TokenAddress, &asset.Symbol, &balance, &asset.USDValue); err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		asset.CurrentBalance = fmt.Sprintf("%.18f", balance)
		details.Assets = append(details.Assets, asset)
		details.TotalUSDValue += asset.USDValue
	}

	return &details, rows.Err()
}

type TransactionResult struct {
	Transactions []Transaction
	TotalCount   int32
}

const (
	DefaultPageLimit = 50
	MaxPageLimit     = 1000
)

func (r *Repo) GetTransactionHistory(ctx context.Context, walletID, address string, page, limit int32, filterByType string) (*TransactionResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = DefaultPageLimit
	} else if limit > MaxPageLimit {
		limit = MaxPageLimit
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var walletUUID string
	if walletID != "" {
		walletUUID = walletID
	} else if address != "" {
		walletUUID = address
	} else {
		return nil, errors.New("wallet_id or address required")
	}

	if filterByType != "" {
		validTypes := map[string]bool{"send": true, "receive": true, "swap": true, "approve": true, "mint": true, "burn": true}
		if !validTypes[filterByType] {
			return nil, fmt.Errorf("invalid filter_by_type: must be one of send, receive, swap, approve, mint, burn")
		}
	}

	var args []interface{}
	var countQuery string
	var query string

	if filterByType != "" {
		args = []interface{}{walletUUID, filterByType}
		countQuery = `
			SELECT COUNT(*)
			FROM transactions_view t
			JOIN wallets w ON t.wallet_id = w.id
			WHERE (w.id = $1 OR w.address = $1) AND t.type = $2`

		args = append(args, limit, offset)
		query = `
			SELECT t.tx_hash, t.from_addr, t.to_addr, t.amount, t.asset_symbol, COALESCE(t.token_address, ''), t.ts, t.type
			FROM transactions_view t
			JOIN wallets w ON t.wallet_id = w.id
			WHERE (w.id = $1 OR w.address = $1) AND t.type = $2
			ORDER BY t.ts DESC
			LIMIT $3 OFFSET $4`
	} else {
		args = []interface{}{walletUUID}
		countQuery = `
			SELECT COUNT(*)
			FROM transactions_view t
			JOIN wallets w ON t.wallet_id = w.id
			WHERE w.id = $1 OR w.address = $1`

		args = append(args, limit, offset)
		query = `
			SELECT t.tx_hash, t.from_addr, t.to_addr, t.amount, t.asset_symbol, COALESCE(t.token_address, ''), t.ts, t.type
			FROM transactions_view t
			JOIN wallets w ON t.wallet_id = w.id
			WHERE w.id = $1 OR w.address = $1
			ORDER BY t.ts DESC
			LIMIT $2 OFFSET $3`
	}

	var totalCount int32
	countArgs := args[:len(args)-2]
	err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count transactions: %w", err)
	}

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
