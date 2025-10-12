package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/repository"
)

func TestKafkaReplayIdempotency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	ctx := context.Background()
	dbURL := "postgres://portfolio:portfolio@localhost:5432/portfolio?sslmode=disable"
	
	poolCfg := repository.PoolConfig{
		MaxConns:          10,
		MinConns:          2,
		MaxConnLifetime:   1 * time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
	}
	
	repo, err := repository.NewPostgres(dbURL, poolCfg)
	require.NoError(t, err, "failed to connect to test database")
	defer repo.Close(ctx)

	event := repository.TransactionDataIngestedEvent{
		WalletAddress: "0xtestWallet",
		Chain:         "ethereum",
		Transactions: []repository.Transaction{
			{
				Hash:         "0xhash1",
				From:         "0xfrom",
				To:           "0xtestWallet",
				Amount:       "100.5",
				Symbol:       "USDC",
				TokenAddress: "0xUSDC",
				Timestamp:    time.Now(),
				Type:         "receive",
			},
		},
	}

	payload, err := json.Marshal(event)
	require.NoError(t, err)

	err = repo.UpsertFromIngest(ctx, payload)
	require.NoError(t, err, "first ingestion failed")

	time.Sleep(100 * time.Millisecond)

	err = repo.UpsertFromIngest(ctx, payload)
	require.NoError(t, err, "second ingestion (replay) failed")

	balances, walletID, _, err := repo.GetCurrentTokenBalances(ctx, "0xtestWallet")
	require.NoError(t, err)
	require.Len(t, balances, 1)

	snapshots := []repository.AssetSnapshotRow{
		{
			WalletID:     walletID,
			TokenAddress: "0xUSDC",
			Balance:      balances[0].Balance,
			USDValue:     "100.5",
		},
	}

	err = repo.InsertAssetSnapshots(ctx, snapshots)
	require.NoError(t, err, "first snapshot insert failed")

	time.Sleep(100 * time.Millisecond)

	err = repo.InsertAssetSnapshots(ctx, snapshots)
	require.NoError(t, err, "second snapshot insert (replay) failed")

	t.Log("Idempotency test completed: snapshots inserted successfully")
}

func TestMarketDataOutageScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	ctx := context.Background()
	dbURL := "postgres://portfolio:portfolio@localhost:5432/portfolio?sslmode=disable"
	
	poolCfg := repository.PoolConfig{
		MaxConns:          10,
		MinConns:          2,
		MaxConnLifetime:   1 * time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
	}
	
	repo, err := repository.NewPostgres(dbURL, poolCfg)
	require.NoError(t, err, "failed to connect to test database")
	defer repo.Close(ctx)

	event := repository.TransactionDataIngestedEvent{
		WalletAddress: "0xtestWallet2",
		Chain:         "ethereum",
		Transactions: []repository.Transaction{
			{
				Hash:         "0xhash2",
				From:         "0xfrom",
				To:           "0xtestWallet2",
				Amount:       "50.25",
				Symbol:       "USDT",
				TokenAddress: "0xUSDT",
				Timestamp:    time.Now(),
				Type:         "receive",
			},
		},
	}

	payload, err := json.Marshal(event)
	require.NoError(t, err)

	err = repo.UpsertFromIngest(ctx, payload)
	require.NoError(t, err, "ingestion failed")

	balances, walletID, _, err := repo.GetCurrentTokenBalances(ctx, "0xtestWallet2")
	require.NoError(t, err)
	require.Len(t, balances, 1)

	snapshots := []repository.AssetSnapshotRow{
		{
			WalletID:     walletID,
			TokenAddress: "0xUSDT",
			Balance:      balances[0].Balance,
			USDValue:     "0",
		},
	}

	err = repo.InsertAssetSnapshots(ctx, snapshots)
	require.NoError(t, err, "snapshot insert with zero USD value failed")

	t.Log("Market data outage scenario test completed: USD value set to 0 successfully")
}
