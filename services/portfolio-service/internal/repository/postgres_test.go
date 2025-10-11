package repository

import (
	"context"
	"encoding/json"
	"errors"
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

		assetRows := pgxmock.NewRows([]string{"token_address", "symbol", "name", "current_balance", "usd_value"}).
			AddRow("0xabc", "USDC", "USD Coin", "1000.5", float64(1000.50)).
			AddRow("0xdef", "USDT", "Tether", "500.25", float64(500.25))

		mock.ExpectQuery("SELECT DISTINCT ON").
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
			WithArgs(walletUUID, pgxmock.AnyArg(), "send", "0xfrom", "0xto", "USDC", "100", "0xhash1").
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectExec("INSERT INTO assets").
			WithArgs(walletUUID, "0xabc", "USDC", "100").
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
		rows := pgxmock.NewRows([]string{"token_address", "symbol", "name", "current_balance", "usd_value"}).
			AddRow("0xabc", "USDC", "USD Coin", "1000.5", float64(1000.50)).
			AddRow("0xdef", "USDT", "Tether", "500.25", float64(500.25))

		mock.ExpectQuery("SELECT DISTINCT ON").
			WithArgs(walletID).
			WillReturnRows(rows)

		portfolio, err := repo.GetPortfolioByWalletID(context.Background(), walletID)

		assert.NoError(t, err)
		assert.Equal(t, walletID, portfolio.WalletID)
		assert.Len(t, portfolio.Assets, 2)
		assert.Equal(t, float64(1500.75), portfolio.TotalUSDValue)
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT DISTINCT ON").
			WithArgs(walletID).
			WillReturnError(errors.New("db error"))

		_, err := repo.GetPortfolioByWalletID(context.Background(), walletID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "query assets")
	})
}
