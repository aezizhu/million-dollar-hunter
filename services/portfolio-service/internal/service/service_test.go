package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	pb "github.com/aezizhu/million-dollar-hunter/services/portfolio-service/proto/portfolio/v1"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/config"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/ports"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockRepo struct {
	mock.Mock
}

type MockMarketDataClient struct {
	mock.Mock
}

func (m *MockMarketDataClient) GetTokenPrice(ctx context.Context, tokenAddress, chain string) (float64, error) {
	args := m.Called(ctx, tokenAddress, chain)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockMarketDataClient) GetTokenPrices(ctx context.Context, tokens map[string][]string) (map[string]map[string]float64, error) {
	args := m.Called(ctx, tokens)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]map[string]float64), args.Error(1)
}

func (m *MockMarketDataClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRepo) VerifyWalletOwnership(ctx context.Context, userID, walletID string) error {
	args := m.Called(ctx, userID, walletID)
	return args.Error(0)
}

func (m *MockRepo) GetPortfolioSummary(ctx context.Context, userID string) ([]repository.WalletSummary, float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]repository.WalletSummary), args.Get(1).(float64), args.Error(2)
}

func (m *MockRepo) GetWalletDetails(ctx context.Context, walletID, address string) (*repository.WalletDetails, error) {
	args := m.Called(ctx, walletID, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.WalletDetails), args.Error(1)
}

func (m *MockRepo) GetTransactionHistory(ctx context.Context, walletID, address string, page, limit int32, filterByType string) (*repository.TransactionResult, error) {
	args := m.Called(ctx, walletID, address, page, limit, filterByType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.TransactionResult), args.Error(1)
}

func (m *MockRepo) GetPortfolioByWalletID(ctx context.Context, walletID string) (*repository.Portfolio, error) {
	args := m.Called(ctx, walletID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Portfolio), args.Error(1)
}

func (m *MockRepo) UpsertFromIngest(ctx context.Context, payload []byte) error {
	args := m.Called(ctx, payload)
	return args.Error(0)
}

func (m *MockRepo) Close(ctx context.Context) {
	m.Called(ctx)
}
func (m *MockRepo) UserOwnsWallet(ctx context.Context, userID, walletIDOrAddr string) (bool, error) {
	args := m.Called(ctx, userID, walletIDOrAddr)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepo) EnrichPortfolioWithPrices(ctx context.Context, portfolio *repository.Portfolio, marketDataClient ports.MarketDataClient) error {
	args := m.Called(ctx, portfolio, marketDataClient)
	return args.Error(0)
}


func TestGetPortfolioSummary(t *testing.T) {
	mockRepo := new(MockRepo)
	mockMarketDataClient := new(MockMarketDataClient)
	cfg := config.Config{ExportDir: "/tmp"}
	svc := New(mockRepo, cfg, mockMarketDataClient)

	t.Run("success", func(t *testing.T) {
		wallets := []repository.WalletSummary{
			{ID: "wallet1", Address: "0x123", Chain: "ethereum", TotalUSDValue: 1000, AssetCount: 5},
			{ID: "wallet2", Address: "0x456", Chain: "polygon", TotalUSDValue: 500, AssetCount: 3},
		}

		mockRepo.On("GetPortfolioSummary", mock.Anything, "user123").
			Return(wallets, float64(1500), nil).Once()

		req := &pb.GetPortfolioSummaryRequest{UserId: "user123"}
		resp, err := svc.GetPortfolioSummary(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Wallets, 2)
		assert.Equal(t, float64(1500), resp.TotalNetWorth)
		assert.Equal(t, "0x123", resp.Wallets[0].Address)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.On("GetPortfolioSummary", mock.Anything, "user123").
			Return([]repository.WalletSummary{}, float64(0), errors.New("db error")).Once()

		req := &pb.GetPortfolioSummaryRequest{UserId: "user123"}
		resp, err := svc.GetPortfolioSummary(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.Internal, status.Code(err))
		mockRepo.AssertExpectations(t)
	})
}

func TestGetWalletDetails(t *testing.T) {
	mockRepo := new(MockRepo)
	mockMarketDataClient := new(MockMarketDataClient)
	cfg := config.Config{ExportDir: "/tmp"}
	svc := New(mockRepo, cfg, mockMarketDataClient)

	t.Run("success", func(t *testing.T) {
		userID := "user123"
		walletID := "wallet1"
		details := &repository.WalletDetails{
			WalletID: walletID,
			Address:  "0x123",
			Chain:    "ethereum",
			Assets: []repository.Asset{
				{TokenAddress: "0xabc", Symbol: "USDC", Name: "USD Coin", CurrentBalance: "1000", USDValue: 1000},
			},
			TotalUSDValue: 1000,
		}

		mockRepo.On("VerifyWalletOwnership", mock.Anything, userID, walletID).
			Return(nil).Once()
		mockRepo.On("GetWalletDetails", mock.Anything, walletID, "").
			Return(details, nil).Once()
		mockRepo.On("EnrichPortfolioWithPrices", mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Once()

		req := &pb.GetWalletDetailsRequest{UserId: userID, WalletId: walletID}
		resp, err := svc.GetWalletDetails(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "wallet1", resp.WalletId)
		assert.Equal(t, "0x123", resp.Address)
		assert.Len(t, resp.Assets, 1)
		assert.Equal(t, float64(1000), resp.TotalUsdValue)
		mockRepo.AssertExpectations(t)
	})

	t.Run("wallet not found", func(t *testing.T) {
		userID := "user123"
		walletID := "wallet1"
		mockRepo.On("VerifyWalletOwnership", mock.Anything, userID, walletID).
			Return(nil).Once()
		mockRepo.On("GetWalletDetails", mock.Anything, walletID, "").
			Return(nil, pgx.ErrNoRows).Once()

		req := &pb.GetWalletDetailsRequest{UserId: userID, WalletId: walletID}
		resp, err := svc.GetWalletDetails(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.NotFound, status.Code(err))
		mockRepo.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		userID := "user123"
		walletID := "wallet1"
		mockRepo.On("VerifyWalletOwnership", mock.Anything, userID, walletID).
			Return(nil).Once()
		mockRepo.On("GetWalletDetails", mock.Anything, walletID, "").
			Return(nil, errors.New("database connection failed")).Once()

		req := &pb.GetWalletDetailsRequest{UserId: userID, WalletId: walletID}
		resp, err := svc.GetWalletDetails(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.Internal, status.Code(err))
		mockRepo.AssertExpectations(t)
	})

	t.Run("missing user_id", func(t *testing.T) {
		req := &pb.GetWalletDetailsRequest{UserId: "", WalletId: "wallet1"}
		resp, err := svc.GetWalletDetails(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("wallet ownership verification fails", func(t *testing.T) {
		userID := "user123"
		walletID := "wallet1"
		mockRepo.On("VerifyWalletOwnership", mock.Anything, userID, walletID).
			Return(errors.New("wallet not found or access denied")).Once()

		req := &pb.GetWalletDetailsRequest{UserId: userID, WalletId: walletID}
		resp, err := svc.GetWalletDetails(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
		mockRepo.AssertExpectations(t)
	})
}

func TestGetTransactionHistory(t *testing.T) {
	mockRepo := new(MockRepo)
	mockMarketDataClient := new(MockMarketDataClient)
	cfg := config.Config{ExportDir: "/tmp"}
	svc := New(mockRepo, cfg, mockMarketDataClient)

	t.Run("success with pagination", func(t *testing.T) {
		result := &repository.TransactionResult{
			Transactions: []repository.Transaction{
				{Hash: "0xhash1", From: "0xfrom", To: "0xto", Amount: "100", Symbol: "USDC"},
			},
			TotalCount: 50,
		}

		mockRepo.On("GetTransactionHistory", mock.Anything, "wallet1", "", int32(1), int32(10), "send").
			Return(result, nil).Once()

		req := &pb.GetTransactionHistoryRequest{
			WalletId:     "wallet1",
			Page:         1,
			Limit:        10,
			FilterByType: "send",
		}
		resp, err := svc.GetTransactionHistory(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Transactions, 1)
		assert.Equal(t, int32(50), resp.TotalCount)
		assert.Equal(t, int32(1), resp.Page)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid filter type error", func(t *testing.T) {
		mockRepo.On("GetTransactionHistory", mock.Anything, "wallet1", "", int32(1), int32(10), "invalid").
			Return(nil, errors.New("invalid filter_by_type: must be one of send, receive, swap, approve, mint, burn")).Once()

		req := &pb.GetTransactionHistoryRequest{
			WalletId:     "wallet1",
			Page:         1,
			Limit:        10,
			FilterByType: "invalid",
		}
		resp, err := svc.GetTransactionHistory(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.Internal, status.Code(err))
		mockRepo.AssertExpectations(t)
	})
}

func TestGetPortfolio(t *testing.T) {
	mockRepo := new(MockRepo)
	mockMarketDataClient := new(MockMarketDataClient)
	cfg := config.Config{ExportDir: "/tmp"}
	svc := New(mockRepo, cfg, mockMarketDataClient)

	t.Run("success", func(t *testing.T) {
		portfolio := &repository.Portfolio{
			WalletID: "wallet1",
			Assets: []repository.Asset{
				{TokenAddress: "0xabc", Symbol: "USDC", Name: "USD Coin", CurrentBalance: "1000", USDValue: 1000},
			},
			TotalUSDValue: 1000,
		}

		mockRepo.On("GetPortfolioByWalletID", mock.Anything, "wallet1").
			Return(portfolio, nil).Once()
		mockRepo.On("EnrichPortfolioWithPrices", mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Once()

		req := &pb.GetPortfolioRequest{WalletId: "wallet1"}
		resp, err := svc.GetPortfolio(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "wallet1", resp.WalletId)
		assert.Len(t, resp.Assets, 1)
		assert.Equal(t, float64(1000), resp.TotalUsdValue)
		mockRepo.AssertExpectations(t)
	})

	t.Run("missing wallet_id", func(t *testing.T) {
		req := &pb.GetPortfolioRequest{WalletId: ""}
		resp, err := svc.GetPortfolio(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

func TestExport(t *testing.T) {
	mockRepo := new(MockRepo)
	mockMarketDataClient := new(MockMarketDataClient)
	cfg := config.Config{ExportDir: "/tmp/portfolio-exports-test"}
	svc := New(mockRepo, cfg, mockMarketDataClient)

	t.Run("success JSON export", func(t *testing.T) {
		userID := "user123"
		validWalletID := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"
		portfolio := &repository.Portfolio{
			WalletID: validWalletID,
			Assets: []repository.Asset{
				{TokenAddress: "0xabc", Symbol: "USDC", Name: "USD Coin", CurrentBalance: "1000", USDValue: 1000},
			},
			TotalUSDValue: 1000,
		}

		mockRepo.On("VerifyWalletOwnership", mock.Anything, userID, validWalletID).
			Return(nil).Once()
		mockRepo.On("GetPortfolioByWalletID", mock.Anything, validWalletID).
			Return(portfolio, nil).Once()

		req := &pb.ExportRequest{
			UserId:   userID,
			WalletId: validWalletID,
			Format:   pb.ExportFormat_EXPORT_FORMAT_JSON,
		}
		resp, err := svc.Export(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Path, "portfolio_"+userID+"_")
		assert.Contains(t, resp.Path, ".json")
		mockRepo.AssertExpectations(t)
	})

	t.Run("success CSV export", func(t *testing.T) {
		userID := "user123"
		validWalletID := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"
		portfolio := &repository.Portfolio{
			WalletID: validWalletID,
			Assets: []repository.Asset{
				{TokenAddress: "0xabc", Symbol: "USDC", Name: "USD Coin", CurrentBalance: "1000", USDValue: 1000},
			},
			TotalUSDValue: 1000,
		}

		mockRepo.On("VerifyWalletOwnership", mock.Anything, userID, validWalletID).
			Return(nil).Once()
		mockRepo.On("GetPortfolioByWalletID", mock.Anything, validWalletID).
			Return(portfolio, nil).Once()

		req := &pb.ExportRequest{
			UserId:   userID,
			WalletId: validWalletID,
			Format:   pb.ExportFormat_EXPORT_FORMAT_CSV,
		}
		resp, err := svc.Export(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Path, "portfolio_"+userID+"_")
		assert.Contains(t, resp.Path, ".csv")
		mockRepo.AssertExpectations(t)
	})

	t.Run("wallet not found", func(t *testing.T) {
		userID := "user123"
		validWalletID := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"
		mockRepo.On("VerifyWalletOwnership", mock.Anything, userID, validWalletID).
			Return(nil).Once()
		mockRepo.On("GetPortfolioByWalletID", mock.Anything, validWalletID).
			Return(nil, pgx.ErrNoRows).Once()

		req := &pb.ExportRequest{UserId: userID, WalletId: validWalletID, Format: pb.ExportFormat_EXPORT_FORMAT_JSON}
		resp, err := svc.Export(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.NotFound, status.Code(err))
		mockRepo.AssertExpectations(t)
	})

	t.Run("missing user_id", func(t *testing.T) {
		req := &pb.ExportRequest{UserId: "", WalletId: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", Format: pb.ExportFormat_EXPORT_FORMAT_JSON}
		resp, err := svc.Export(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("missing wallet_id", func(t *testing.T) {
		req := &pb.ExportRequest{UserId: "user123", WalletId: "", Format: pb.ExportFormat_EXPORT_FORMAT_JSON}
		resp, err := svc.Export(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("invalid wallet_id format", func(t *testing.T) {
		req := &pb.ExportRequest{UserId: "user123", WalletId: "invalid_wallet", Format: pb.ExportFormat_EXPORT_FORMAT_JSON}
		resp, err := svc.Export(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("path traversal attempt", func(t *testing.T) {
		req := &pb.ExportRequest{UserId: "user123", WalletId: "../../etc/passwd", Format: pb.ExportFormat_EXPORT_FORMAT_JSON}
		resp, err := svc.Export(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("wallet ownership verification fails", func(t *testing.T) {
		userID := "user123"
		validWalletID := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"
		mockRepo.On("VerifyWalletOwnership", mock.Anything, userID, validWalletID).
			Return(errors.New("wallet not found or access denied")).Once()

		req := &pb.ExportRequest{UserId: userID, WalletId: validWalletID, Format: pb.ExportFormat_EXPORT_FORMAT_JSON}
		resp, err := svc.Export(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
		mockRepo.AssertExpectations(t)
	})
}

func TestHandleTransactionDataIngested(t *testing.T) {
	mockRepo := new(MockRepo)
	mockMarketDataClient := new(MockMarketDataClient)
	cfg := config.Config{ExportDir: "/tmp"}
	svc := New(mockRepo, cfg, mockMarketDataClient)

	t.Run("success", func(t *testing.T) {
		payload := []byte(`{"wallet_address":"0x123","chain":"ethereum","transactions":[]}`)

		mockRepo.On("UpsertFromIngest", mock.Anything, payload).
			Return(nil).Once()

		err := svc.HandleTransactionDataIngested(context.Background(), payload)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		payload := []byte(`{"wallet_address":"0x123","chain":"ethereum","transactions":[]}`)

		mockRepo.On("UpsertFromIngest", mock.Anything, payload).
			Return(errors.New("db error")).Once()

		err := svc.HandleTransactionDataIngested(context.Background(), payload)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}
func TestGetPortfolioWithoutMarketDataService(t *testing.T) {
	mockRepo := new(MockRepo)
	cfg := config.Config{ExportDir: "/tmp"}
	svc := New(mockRepo, cfg, nil)

	t.Run("success without price enrichment", func(t *testing.T) {
		portfolio := &repository.Portfolio{
			WalletID: "wallet1",
			Chain:    "ethereum",
			Assets: []repository.Asset{
				{TokenAddress: "0xabc", Symbol: "USDC", Name: "USD Coin", CurrentBalance: "1000", USDValue: 0},
			},
			TotalUSDValue: 0,
		}

		mockRepo.On("GetPortfolioByWalletID", mock.Anything, "wallet1").
			Return(portfolio, nil).Once()

		req := &pb.GetPortfolioRequest{WalletId: "wallet1"}
		resp, err := svc.GetPortfolio(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "wallet1", resp.WalletId)
		assert.Len(t, resp.Assets, 1)
		assert.Equal(t, float64(0), resp.Assets[0].UsdValue)
		assert.Equal(t, float64(0), resp.TotalUsdValue)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetPortfolioEnrichmentError(t *testing.T) {
	mockRepo := new(MockRepo)
	mockMarketDataClient := new(MockMarketDataClient)
	cfg := config.Config{ExportDir: "/tmp"}
	svc := New(mockRepo, cfg, mockMarketDataClient)

	t.Run("enrichment fails gracefully", func(t *testing.T) {
		portfolio := &repository.Portfolio{
			WalletID: "wallet1",
			Chain:    "ethereum",
			Assets: []repository.Asset{
				{TokenAddress: "0xabc", Symbol: "USDC", Name: "USD Coin", CurrentBalance: "1000", USDValue: 0},
			},
			TotalUSDValue: 0,
		}

		mockRepo.On("GetPortfolioByWalletID", mock.Anything, "wallet1").
			Return(portfolio, nil).Once()
		mockRepo.On("EnrichPortfolioWithPrices", mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("market-data-service unavailable")).Once()

		req := &pb.GetPortfolioRequest{WalletId: "wallet1"}
		resp, err := svc.GetPortfolio(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "wallet1", resp.WalletId)
		assert.Len(t, resp.Assets, 1)
		assert.Equal(t, float64(0), resp.Assets[0].UsdValue)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetPortfolioDeduplication(t *testing.T) {
	mockRepo := new(MockRepo)
	mockMarketDataClient := new(MockMarketDataClient)
	cfg := config.Config{ExportDir: "/tmp"}
	svc := New(mockRepo, cfg, mockMarketDataClient)

	t.Run("deduplicate token addresses", func(t *testing.T) {
		portfolio := &repository.Portfolio{
			WalletID: "wallet1",
			Chain:    "ethereum",
			Assets: []repository.Asset{
				{TokenAddress: "0xABC", Symbol: "USDC", Name: "USD Coin", CurrentBalance: "1000", USDValue: 0},
				{TokenAddress: "0xabc", Symbol: "USDC", Name: "USD Coin", CurrentBalance: "500", USDValue: 0},
				{TokenAddress: "0xDEF", Symbol: "USDT", Name: "Tether", CurrentBalance: "250", USDValue: 0},
			},
			TotalUSDValue: 0,
		}

		mockRepo.On("GetPortfolioByWalletID", mock.Anything, "wallet1").
			Return(portfolio, nil).Once()

		mockRepo.On("EnrichPortfolioWithPrices", mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Once()

		req := &pb.GetPortfolioRequest{WalletId: "wallet1"}
		resp, err := svc.GetPortfolio(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "wallet1", resp.WalletId)
		assert.Len(t, resp.Assets, 3)
		mockRepo.AssertExpectations(t)
	})
}
