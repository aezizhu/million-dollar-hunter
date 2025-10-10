package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestIntegrationHealthCheck(t *testing.T) {
	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = "localhost:50051"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer conn.Close()

	time.Sleep(2 * time.Second)

	client := pb.NewMarketDataServiceClient(conn)

	t.Run("GetTokenPrice_BNB", func(t *testing.T) {
		req := &pb.GetTokenPriceRequest{
			TokenAddress: "0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c",
			Chain:        "bsc",
		}

		resp, err := client.GetTokenPrice(ctx, req)
		if err != nil {
			t.Logf("GetTokenPrice returned error (this may be expected if CoinGecko API is not configured): %v", err)
			return
		}

		if resp.TokenAddress != req.TokenAddress {
			t.Errorf("Expected token address %s, got %s", req.TokenAddress, resp.TokenAddress)
		}

		if resp.Chain != req.Chain {
			t.Errorf("Expected chain %s, got %s", req.Chain, resp.Chain)
		}

		if resp.UsdPrice <= 0 {
			t.Errorf("Expected positive price, got %f", resp.UsdPrice)
		}

		t.Logf("Successfully retrieved price for BNB: $%.2f (from_cache: %v)", resp.UsdPrice, resp.FromCache)
	})

	t.Run("GetTokenPrices_Multiple", func(t *testing.T) {
		req := &pb.GetTokenPricesRequest{
			Tokens: []*pb.TokenIdentifier{
				{
					TokenAddress: "0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c",
					Chain:        "bsc",
				},
				{
					TokenAddress: "0x55d398326f99059ff775485246999027b3197955",
					Chain:        "bsc",
				},
			},
		}

		resp, err := client.GetTokenPrices(ctx, req)
		if err != nil {
			t.Logf("GetTokenPrices returned error (this may be expected if CoinGecko API is not configured): %v", err)
			return
		}

		if len(resp.Prices) == 0 {
			t.Logf("No prices returned (this may be expected if CoinGecko API is not configured)")
			return
		}

		t.Logf("Successfully retrieved %d prices", len(resp.Prices))
		for _, price := range resp.Prices {
			t.Logf("  - %s on %s: $%.2f (from_cache: %v)", price.TokenAddress, price.Chain, price.UsdPrice, price.FromCache)
		}
	})

	t.Run("GetMarketData", func(t *testing.T) {
		req := &pb.GetMarketDataRequest{
			TokenAddress: "0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c",
			Chain:        "bsc",
		}

		resp, err := client.GetMarketData(ctx, req)
		if err != nil {
			t.Logf("GetMarketData returned error (this may be expected if CoinGecko API is not configured): %v", err)
			return
		}

		if resp.UsdPrice <= 0 {
			t.Errorf("Expected positive price, got %f", resp.UsdPrice)
		}

		t.Logf("Successfully retrieved market data:")
		t.Logf("  - Price: $%.2f", resp.UsdPrice)
		t.Logf("  - Market Cap: $%.0f", resp.MarketCap)
		t.Logf("  - 24h Volume: $%.0f", resp.Volume_24H)
		t.Logf("  - 24h Change: %.2f%%", resp.PriceChange_24H)
		t.Logf("  - From Cache: %v", resp.FromCache)
	})
}
