package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

type Repository struct {
	db     *sql.DB
	logger zerolog.Logger
}

type TokenPrice struct {
	TokenAddress   string
	Chain          string
	USDPrice       float64
	MarketCap      float64
	Volume24h      float64
	PriceChange24h float64
	LastUpdated    time.Time
}

func NewRepository(connString string, logger zerolog.Logger) (*Repository, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	logger.Info().Msg("Connected to PostgreSQL database")

	return &Repository{
		db:     db,
		logger: logger.With().Str("component", "repository").Logger(),
	}, nil
}

func (r *Repository) SaveTokenPrice(ctx context.Context, price *TokenPrice) error {
	query := `
		INSERT INTO token_prices (token_address, chain, usd_price, market_cap, volume_24h, price_change_24h, last_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (token_address, chain) 
		DO UPDATE SET
			usd_price = EXCLUDED.usd_price,
			market_cap = EXCLUDED.market_cap,
			volume_24h = EXCLUDED.volume_24h,
			price_change_24h = EXCLUDED.price_change_24h,
			last_updated = EXCLUDED.last_updated
	`

	_, err := r.db.ExecContext(ctx, query,
		price.TokenAddress,
		price.Chain,
		price.USDPrice,
		price.MarketCap,
		price.Volume24h,
		price.PriceChange24h,
		price.LastUpdated,
	)

	if err != nil {
		return fmt.Errorf("failed to save token price: %w", err)
	}

	r.logger.Debug().
		Str("token", price.TokenAddress).
		Str("chain", price.Chain).
		Float64("price", price.USDPrice).
		Msg("Saved token price to database")

	return nil
}

func (r *Repository) GetTokenPrice(ctx context.Context, tokenAddress, chain string) (*TokenPrice, error) {
	query := `
		SELECT token_address, chain, usd_price, market_cap, volume_24h, price_change_24h, last_updated
		FROM token_prices
		WHERE token_address = $1 AND chain = $2
	`

	var price TokenPrice
	err := r.db.QueryRowContext(ctx, query, tokenAddress, chain).Scan(
		&price.TokenAddress,
		&price.Chain,
		&price.USDPrice,
		&price.MarketCap,
		&price.Volume24h,
		&price.PriceChange24h,
		&price.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get token price: %w", err)
	}

	return &price, nil
}

func (r *Repository) SaveMultipleTokenPrices(ctx context.Context, prices []*TokenPrice) error {
	if len(prices) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO token_prices (token_address, chain, usd_price, market_cap, volume_24h, price_change_24h, last_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (token_address, chain) 
		DO UPDATE SET
			usd_price = EXCLUDED.usd_price,
			market_cap = EXCLUDED.market_cap,
			volume_24h = EXCLUDED.volume_24h,
			price_change_24h = EXCLUDED.price_change_24h,
			last_updated = EXCLUDED.last_updated
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, price := range prices {
		_, err := stmt.ExecContext(ctx,
			price.TokenAddress,
			price.Chain,
			price.USDPrice,
			price.MarketCap,
			price.Volume24h,
			price.PriceChange24h,
			price.LastUpdated,
		)
		if err != nil {
			return fmt.Errorf("failed to insert price for %s: %w", price.TokenAddress, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Debug().
		Int("count", len(prices)).
		Msg("Saved multiple token prices to database")

	return nil
}

func (r *Repository) GetAllTokenAddresses(ctx context.Context) (map[string][]string, error) {
	query := `
		SELECT DISTINCT chain, token_address
		FROM token_prices
		ORDER BY chain, token_address
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query token addresses: %w", err)
	}
	defer rows.Close()

	tokens := make(map[string][]string)
	for rows.Next() {
		var chain, address string
		if err := rows.Scan(&chain, &address); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		tokens[chain] = append(tokens[chain], address)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return tokens, nil
}

func (r *Repository) GetStaleTokens(ctx context.Context, staleDuration time.Duration) (map[string][]string, error) {
	query := `
		SELECT DISTINCT chain, token_address
		FROM token_prices
		WHERE last_updated < $1
		ORDER BY chain, token_address
	`

	cutoffTime := time.Now().Add(-staleDuration)
	rows, err := r.db.QueryContext(ctx, query, cutoffTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query stale tokens: %w", err)
	}
	defer rows.Close()

	tokens := make(map[string][]string)
	for rows.Next() {
		var chain, address string
		if err := rows.Scan(&chain, &address); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		tokens[chain] = append(tokens[chain], address)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return tokens, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}
