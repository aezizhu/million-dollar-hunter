package repository

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPortfolioSummary(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &Repo{db: mock}
	userID := uuid.New().String()

	t.Run("success with multiple wallets", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "address", "chain", "asset_count", "total_usd_value"}).
			AddRow(uuid.New().String(), "0x123", "ethereum", int32(5), float64(1000.50)).
			AddRow(uuid.New().String(), "0x456", "polygon", int32(3), float64(500.25))

		mock.ExpectQuery("SELECT.*FROM wallets").
			WithArgs(userID).
			WillReturnRows(rows)

		wallets, totalNetWorth, err := repo.GetPortfolioSummary(context.Background(), userID)

		assert.NoError(t, err)
		assert.Len(t, wallets, 2)
		assert.Equal(t, float64(1500.75), totalNetWorth)
		assert.Equal(t, "0x123", wallets[0].Address)
		assert.Equal(t, int32(5), wallets[0].AssetCount)
	})

	t.Run("empty results", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "address", "chain", "asset_count", "total_usd_value"})

		mock.ExpectQuery("SELECT.*FROM wallets").
			WithArgs(userID).
			WillReturnRows(rows)

		wallets, totalNetWorth, err := repo.GetPortfolioSummary(context.Background(), userID)

		assert.NoError(t, err)
		assert.Empty(t, wallets)
		assert.Equal(t, float64(0), totalNetWorth)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery("SELECT.*FROM wallets").
			WithArgs(userID).
			WillReturnError(errors.New("db connection failed"))

		_, _, err := repo.GetPortfolioSummary(context.Background(), userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "query wallets")
	})
}

func TestGetWalletDetails(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &Repo{db: mock}
	walletUUID := uuid.New().String()

	t.Run("success by wallet ID", func(t *testing.T) {
		walletRow := pgxmock.NewRows([]string{"id", "address", "chain"}).
			AddRow(walletUUID, "0x123", "ethereum")

		mock.ExpectQuery("SELECT id, address, chain FROM wallets WHERE").
			WithArgs(walletUUID).
			WillReturnRows(walletRow)

		assetRows := pgxmock.NewRows([]string{"token_address", "symbol", "current_balance", "usd_value"}).
			AddRow("0xabc", "USDC", float64(1000.5), float64(1000.50)).
			AddRow("0xdef", "USDT", float64(500.25), float64(500.25))

		mock.ExpectQuery("SELECT.*FROM transactions_view").
			WithArgs(walletUUID).
			WillReturnRows(assetRows)

		details, err := repo.GetWalletDetails(context.Background(), walletUUID, "")

		assert.NoError(t, err)
		assert.Equal(t, walletUUID, details.WalletID)
		assert.Equal(t, "0x123", details.Address)
		assert.Equal(t, "ethereum", details.Chain)
		assert.Len(t, details.Assets, 2)
		assert.Equal(t, float64(1500.75), details.TotalUSDValue)
	})

	t.Run("wallet not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, address, chain FROM wallets WHERE").
			WithArgs(walletUUID).
			WillReturnError(pgx.ErrNoRows)

		_, err := repo.GetWalletDetails(context.Background(), walletUUID, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wallet not found")
	})

	t.Run("missing wallet_id and address", func(t *testing.T) {
		_, err := repo.GetWalletDetails(context.Background(), "", "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wallet_id or address required")
	})
}

func TestGetTransactionHistory(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &Repo{db: mock}
	walletUUID := uuid.New().String()

	t.Run("success with pagination", func(t *testing.T) {
		countRow := pgxmock.NewRows([]string{"count"}).AddRow(int32(100))
		mock.ExpectQuery("SELECT COUNT").
			WithArgs(walletUUID).
			WillReturnRows(countRow)

		txRows := pgxmock.NewRows([]string{"tx_hash", "from_addr", "to_addr", "amount", "asset_symbol", "token_address", "ts", "type"}).
			AddRow("0xhash1", "0xfrom", "0xto", "100.5", "USDC", "0xabc", time.Now(), "send").
			AddRow("0xhash2", "0xfrom2", "0xto2", "50.25", "USDT", "0xdef", time.Now(), "receive")

		mock.ExpectQuery("SELECT t.tx_hash").
			WithArgs(walletUUID, int32(10), int32(0)).
			WillReturnRows(txRows)

		result, err := repo.GetTransactionHistory(context.Background(), walletUUID, "", 1, 10, "")

		assert.NoError(t, err)
		assert.Equal(t, int32(100), result.TotalCount)
		assert.Len(t, result.Transactions, 2)
		assert.Equal(t, "0xhash1", result.Transactions[0].Hash)
	})

	t.Run("invalid filter type", func(t *testing.T) {
		_, err := repo.GetTransactionHistory(context.Background(), walletUUID, "", 1, 10, "invalid_type")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid filter_by_type")
	})

	t.Run("valid filter type accepted", func(t *testing.T) {
		countRow := pgxmock.NewRows([]string{"count"}).AddRow(int32(50))
		mock.ExpectQuery("SELECT COUNT").
			WithArgs(walletUUID, "send").
			WillReturnRows(countRow)

		txRows := pgxmock.NewRows([]string{"tx_hash", "from_addr", "to_addr", "amount", "asset_symbol", "token_address", "ts", "type"})
		mock.ExpectQuery("SELECT t.tx_hash").
			WithArgs(walletUUID, "send", int32(10), int32(0)).
			WillReturnRows(txRows)

		result, err := repo.GetTransactionHistory(context.Background(), walletUUID, "", 1, 10, "send")

		assert.NoError(t, err)
		assert.Equal(t, int32(50), result.TotalCount)
	})
}

func TestUpsertFromIngest(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &Repo{db: mock}

	t.Run("success", func(t *testing.T) {
		event := TransactionDataIngestedEvent{
			WalletAddress: "0x123",
			Chain:         "ethereum",
			Transactions: []Transaction{
				{
					Hash:         "0xhash1",
					From:         "0xfrom",
					To:           "0xto",
					Amount:       "100",
					Symbol:       "USDC",
					TokenAddress: "0xabc",
					Timestamp:    time.Now(),
					Type:         "send",
				},
			},
		}
		payload, _ := json.Marshal(event)

		walletUUID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO wallets").
			WithArgs("0x123", "ethereum").
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(walletUUID))

		mock.ExpectExec("INSERT INTO transactions_view").
			WithArgs(walletUUID, pgxmock.AnyArg(), "send", "0xfrom", "0xto", "USDC", "100", "0xhash1", "0xabc").
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectCommit()

		err := repo.UpsertFromIngest(context.Background(), payload)

		assert.NoError(t, err)
	})

	t.Run("empty payload", func(t *testing.T) {
		err := repo.UpsertFromIngest(context.Background(), []byte{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty payload")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		err := repo.UpsertFromIngest(context.Background(), []byte("invalid json"))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal event")
	})
}

func TestGetPortfolioByWalletID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &Repo{db: mock}
	walletID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT chain FROM wallets").
			WithArgs(walletID).
			WillReturnRows(pgxmock.NewRows([]string{"chain"}).AddRow("ethereum"))

		rows := pgxmock.NewRows([]string{"token_address", "symbol", "current_balance", "usd_value"}).
			AddRow("0xabc", "USDC", float64(1000.5), float64(1000.50)).
			AddRow("0xdef", "USDT", float64(500.25), float64(500.25))

		mock.ExpectQuery("SELECT.*FROM transactions_view").
			WithArgs(walletID).
			WillReturnRows(rows)

		portfolio, err := repo.GetPortfolioByWalletID(context.Background(), walletID)

		assert.NoError(t, err)
		assert.Equal(t, walletID, portfolio.WalletID)
		assert.Equal(t, "ethereum", portfolio.Chain)
		assert.Len(t, portfolio.Assets, 2)
		assert.Equal(t, float64(1500.75), portfolio.TotalUSDValue)
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT chain FROM wallets").
			WithArgs(walletID).
			WillReturnRows(pgxmock.NewRows([]string{"chain"}).AddRow("ethereum"))

		mock.ExpectQuery("SELECT.*FROM transactions_view").
			WithArgs(walletID).
			WillReturnError(errors.New("db error"))

		_, err := repo.GetPortfolioByWalletID(context.Background(), walletID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "query assets")
	})
}

func TestBalanceAccumulation(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &Repo{db: mock}

	t.Run("multiple transactions accumulate correctly", func(t *testing.T) {
		event := TransactionDataIngestedEvent{
			WalletAddress: "0x123",
			Chain:         "ethereum",
			Transactions: []Transaction{
				{Hash: "0xhash1", TokenAddress: "0xUSDC", Symbol: "USDC", Amount: "100", Type: "receive", Timestamp: time.Now(), From: "0xfrom1", To: "0x123"},
				{Hash: "0xhash2", TokenAddress: "0xUSDC", Symbol: "USDC", Amount: "50", Type: "receive", Timestamp: time.Now(), From: "0xfrom2", To: "0x123"},
				{Hash: "0xhash3", TokenAddress: "0xUSDC", Symbol: "USDC", Amount: "25", Type: "send", Timestamp: time.Now(), From: "0x123", To: "0xto1"},
			},
		}
		payload, _ := json.Marshal(event)

		walletUUID := uuid.New()

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO wallets").
			WithArgs("0x123", "ethereum").
			WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(walletUUID))

		mock.ExpectExec("INSERT INTO transactions_view").
			WithArgs(walletUUID, pgxmock.AnyArg(), "receive", "0xfrom1", "0x123", "USDC", "100", "0xhash1", "0xUSDC").
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectExec("INSERT INTO transactions_view").
			WithArgs(walletUUID, pgxmock.AnyArg(), "receive", "0xfrom2", "0x123", "USDC", "50", "0xhash2", "0xUSDC").
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectExec("INSERT INTO transactions_view").
			WithArgs(walletUUID, pgxmock.AnyArg(), "send", "0x123", "0xto1", "USDC", "25", "0xhash3", "0xUSDC").
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectCommit()

		err := repo.UpsertFromIngest(context.Background(), payload)
		assert.NoError(t, err)

		mock.ExpectQuery("SELECT chain FROM wallets").
			WithArgs(walletUUID.String()).
			WillReturnRows(pgxmock.NewRows([]string{"chain"}).AddRow("ethereum"))

		balanceRows := pgxmock.NewRows([]string{"token_address", "symbol", "current_balance", "usd_value"}).
			AddRow("0xUSDC", "USDC", float64(125), float64(125))

		mock.ExpectQuery("SELECT.*FROM transactions_view").
			WithArgs(walletUUID.String()).
			WillReturnRows(balanceRows)

		portfolio, err := repo.GetPortfolioByWalletID(context.Background(), walletUUID.String())

		assert.NoError(t, err)
		assert.Equal(t, "ethereum", portfolio.Chain)
		assert.Len(t, portfolio.Assets, 1)
		assert.Equal(t, "125.000000000000000000", portfolio.Assets[0].CurrentBalance)
	})
}

func TestEnrichPortfolioWithPrices_Deduplication(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &Repo{db: mock}

	type capturedTokens struct {
		mu     sync.Mutex
		tokens map[string][]string
	}

	captured := &capturedTokens{tokens: make(map[string][]string)}

	GetTokenPrices := func(ctx context.Context, tokens map[string][]string) (map[string]map[string]float64, error) {
		captured.mu.Lock()
		defer captured.mu.Unlock()
		for chain, addrs := range tokens {
			captured.tokens[chain] = append(captured.tokens[chain], addrs...)
		}
		
		result := make(map[string]map[string]float64)
		for chain, addrs := range tokens {
			result[chain] = make(map[string]float64)
			for _, addr := range addrs {
				result[chain][addr] = 1.0
			}
		}
		return result, nil
	}

	t.Run("deduplicates identical addresses", func(t *testing.T) {
		captured.tokens = make(map[string][]string)

		portfolio := &Portfolio{
			WalletID: "wallet1",
			Chain:    "Ethereum",
			Assets: []Asset{
				{TokenAddress: "0xABC", Symbol: "TOKEN", Name: "Token", CurrentBalance: "100"},
				{TokenAddress: "0xabc", Symbol: "TOKEN", Name: "Token", CurrentBalance: "50"},
				{TokenAddress: "0xABC", Symbol: "TOKEN", Name: "Token", CurrentBalance: "25"},
			},
		}

		mockMarketDataClient := &mockMarketDataClientForTest{
			getTokenPricesFn: GetTokenPrices,
		}

		err := repo.EnrichPortfolioWithPrices(context.Background(), portfolio, mockMarketDataClient)
		assert.NoError(t, err)

		captured.mu.Lock()
		defer captured.mu.Unlock()
		
		ethAddrs := captured.tokens["ethereum"]
		require.Len(t, ethAddrs, 1, "Expected only 1 unique token after deduplication")
		assert.Equal(t, "0xabc", ethAddrs[0], "Token address should be lowercased")
	})
}

type mockMarketDataClientForTest struct {
	getTokenPricesFn func(ctx context.Context, tokens map[string][]string) (map[string]map[string]float64, error)
}

func (m *mockMarketDataClientForTest) GetTokenPrice(ctx context.Context, tokenAddress, chain string) (float64, error) {
	return 0, nil
}

func (m *mockMarketDataClientForTest) GetTokenPrices(ctx context.Context, tokens map[string][]string) (map[string]map[string]float64, error) {
	if m.getTokenPricesFn != nil {
		return m.getTokenPricesFn(ctx, tokens)
	}
	return nil, nil
}

func (m *mockMarketDataClientForTest) Close() error {
	return nil
}

func TestEnrichPortfolioWithPrices_PerAssetParseError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := &Repo{db: mock}

	t.Run("handles parse error for one asset, enriches others", func(t *testing.T) {
		portfolio := &Portfolio{
			WalletID: "wallet1",
			Chain:    "ethereum",
			Assets: []Asset{
				{TokenAddress: "0xabc", Symbol: "GOOD", Name: "Good Token", CurrentBalance: "100.5"},
				{TokenAddress: "0xdef", Symbol: "BAD", Name: "Bad Token", CurrentBalance: "invalid_number"},
				{TokenAddress: "0x123", Symbol: "ALSO_GOOD", Name: "Also Good", CurrentBalance: "50.25"},
			},
		}

		mockClient := &mockMarketDataClientForTest{
			getTokenPricesFn: func(ctx context.Context, tokens map[string][]string) (map[string]map[string]float64, error) {
				return map[string]map[string]float64{
					"ethereum": {
						"0xabc": 2.0,
						"0xdef": 3.0,
						"0x123": 4.0,
					},
				}, nil
			},
		}

		err := repo.EnrichPortfolioWithPrices(context.Background(), portfolio, mockClient)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "1 asset parse errors")

		assert.Equal(t, 100.5*2.0, portfolio.Assets[0].USDValue, "First asset should be enriched")
		assert.Equal(t, float64(0), portfolio.Assets[1].USDValue, "Second asset (parse error) should have USD=0")
		assert.Equal(t, 50.25*4.0, portfolio.Assets[2].USDValue, "Third asset should be enriched")

		expectedTotal := (100.5 * 2.0) + (50.25 * 4.0)
		assert.Equal(t, expectedTotal, portfolio.TotalUSDValue, "Total should exclude the failed asset")
	})
}
