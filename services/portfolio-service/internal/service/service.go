package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	pb "github.com/aezizhu/million-dollar-hunter/services/portfolio-service/proto/portfolio/v1"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/config"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/ports"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Repository interface {
	VerifyWalletOwnership(ctx context.Context, userID, walletID string) error
	GetPortfolioSummary(ctx context.Context, userID string) ([]repository.WalletSummary, float64, error)
	GetWalletDetails(ctx context.Context, walletID, address string) (*repository.WalletDetails, error)
	GetTransactionHistory(ctx context.Context, walletID, address string, page, limit int32, filterByType string) (*repository.TransactionResult, error)
	GetPortfolioByWalletID(ctx context.Context, walletID string) (*repository.Portfolio, error)
	UserOwnsWallet(ctx context.Context, userID, walletIDOrAddr string) (bool, error)
	UpsertFromIngest(ctx context.Context, payload []byte) error
	GetCurrentTokenBalances(ctx context.Context, walletIDOrAddr string) ([]repository.TokenBalance, string, string, error)
	InsertAssetSnapshots(ctx context.Context, snapshots []repository.AssetSnapshotRow) error
	EnrichPortfolioWithPrices(ctx context.Context, portfolio *repository.Portfolio, marketDataClient ports.MarketDataClient) error
	Close(ctx context.Context)
}

type noopMarketData struct{}

func (n *noopMarketData) GetTokenPrice(ctx context.Context, tokenAddress, chain string) (float64, error) {
	return 0, nil
}

func (n *noopMarketData) GetTokenPrices(ctx context.Context, tokens map[string][]string) (map[string]map[string]float64, error) {
	return map[string]map[string]float64{}, nil
}

func (n *noopMarketData) Close() error { return nil }

type PortfolioService struct {
	repo             Repository
	cfg              config.Config
	marketDataClient ports.MarketDataClient
	metrics          interface {
		IncCounter(name string, labels map[string]string)
		ObserveDuration(name string, seconds float64, labels map[string]string)
	}
}

func New(repo Repository, cfg config.Config, marketDataClient ports.MarketDataClient) *PortfolioService {
	return &PortfolioService{
		repo:             repo,
		cfg:              cfg,
		marketDataClient: marketDataClient,
		metrics:          nil,
	}
}
func (s *PortfolioService) WithMetrics(rec interface {
	IncCounter(string, map[string]string)
	ObserveDuration(string, float64, map[string]string)
}) *PortfolioService {
	s.metrics = rec
	return s
}


func (s *PortfolioService) HandleTransactionDataIngested(ctx context.Context, raw []byte) error {
	if err := s.repo.UpsertFromIngest(ctx, raw); err != nil {
		return err
	}

	var evt repository.TransactionDataIngestedEvent
	if err := json.Unmarshal(raw, &evt); err != nil {
		if s.metrics != nil {
			s.metrics.IncCounter("enrichment.failure", map[string]string{"reason": "unmarshal"})
		}
		log.Printf("ingest: unmarshal failed err=%v", err)
		return err
	}

	log.Printf("enrich: start wallet=%s chain=%s", evt.WalletAddress, evt.Chain)

	balances, walletID, chain, err := s.repo.GetCurrentTokenBalances(ctx, evt.WalletAddress)
	if err != nil {
		if s.metrics != nil {
			s.metrics.IncCounter("enrichment.failure", map[string]string{"reason": "balances", "chain": evt.Chain})
		}
		log.Printf("enrich: GetCurrentTokenBalances failed wallet=%s chain=%s err=%v", evt.WalletAddress, evt.Chain, err)
		return nil
	}

	tokensByChain := make(map[string][]string)
	seen := map[string]struct{}{}
	lchain := strings.ToLower(chain)
	for _, b := range balances {
		addr := strings.ToLower(strings.TrimSpace(b.TokenAddress))
		if addr == "" {
			continue
		}
		key := lchain + ":" + addr
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		tokensByChain[lchain] = append(tokensByChain[lchain], addr)
	}
	log.Printf("enrich: tokens deduped wallet=%s chain=%s tokens=%d", evt.WalletAddress, chain, len(tokensByChain[lchain]))

	priceMap := map[string]float64{}
	if len(tokensByChain) > 0 && s.marketDataClient != nil {
		start := time.Now()
		prices, err := s.marketDataClient.GetTokenPrices(ctx, tokensByChain)
		dur := time.Since(start).Seconds()
		if s.metrics != nil {
			s.metrics.ObserveDuration("market_data.request.seconds", dur, map[string]string{
				"chain": chain,
			})
			if err != nil {
				s.metrics.IncCounter("market_data.request.failure", map[string]string{"chain": chain})
			} else {
				s.metrics.IncCounter("market_data.request.success", map[string]string{"chain": chain})
			}
		}
		if err != nil {
			log.Printf("enrich: GetTokenPrices failed wallet=%s chain=%s tokens=%d err=%v; defaulting price=0", evt.WalletAddress, chain, len(tokensByChain[lchain]), err)
		} else {
			for ch, m := range prices {
				lch := strings.ToLower(ch)
				for addr, price := range m {
					priceMap[lch+":"+strings.ToLower(addr)] = price
				}
			}
		}
	}

	rows := make([]repository.AssetSnapshotRow, 0, len(balances))
	for _, b := range balances {
		key := strings.ToLower(chain) + ":" + strings.ToLower(b.TokenAddress)
		price := priceMap[key]
		usd := b.Balance * price
		rows = append(rows, repository.AssetSnapshotRow{
			WalletID:     walletID,
			TokenAddress: b.TokenAddress,
			Balance:      b.Balance,
			USDValue:     usd,
		})
	}

	if len(rows) > 0 {
		if err := s.repo.InsertAssetSnapshots(ctx, rows); err != nil {
			if s.metrics != nil {
				s.metrics.IncCounter("enrichment.failure", map[string]string{"reason": "insert", "chain": chain})
			}
			log.Printf("enrich: InsertAssetSnapshots failed wallet=%s chain=%s rows=%d err=%v", evt.WalletAddress, chain, len(rows), err)
			return err
		}
		if s.metrics != nil {
			s.metrics.IncCounter("enrichment.snapshot_writes", map[string]string{"chain": chain})
		}
		log.Printf("enrich: snapshot write ok wallet=%s chain=%s rows=%d", evt.WalletAddress, chain, len(rows))
	}
	return nil
}

func (s *PortfolioService) GetPortfolio(ctx context.Context, req *pb.GetPortfolioRequest) (*pb.GetPortfolioResponse, error) {
	if req.GetWalletId() == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet_id is required")
	}

	portfolio, err := s.repo.GetPortfolioByWalletID(ctx, req.GetWalletId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetch portfolio: %v", err)
	}

	if s.marketDataClient != nil {
		if err := s.repo.EnrichPortfolioWithPrices(ctx, portfolio, s.marketDataClient); err != nil {
			log.Printf("WARNING: Failed to enrich portfolio with prices: %v. Returning portfolio with USD values as 0.", err)
		}
	}

	assets := make([]*pb.Asset, 0, len(portfolio.Assets))
	for _, a := range portfolio.Assets {
		assets = append(assets, &pb.Asset{
			TokenAddress: a.TokenAddress,
			Symbol:       a.Symbol,
			Name:         a.Name,
			Amount:       a.CurrentBalance,
			UsdValue:     a.USDValue,
		})
	}

	return &pb.GetPortfolioResponse{
		WalletId:      portfolio.WalletID,
		Assets:        assets,
		TotalUsdValue: portfolio.TotalUSDValue,
	}, nil
}

func (s *PortfolioService) Export(ctx context.Context, req *pb.ExportRequest) (*pb.ExportResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if req.GetWalletId() == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet_id is required")
	}

	if !isValidWalletID(req.GetWalletId()) {
		return nil, status.Error(codes.InvalidArgument, "invalid wallet_id format")
	}

	if err := s.repo.VerifyWalletOwnership(ctx, req.GetUserId(), req.GetWalletId()); err != nil {
		return nil, status.Error(codes.PermissionDenied, "wallet not found or access denied")
	}

	portfolio, err := s.repo.GetPortfolioByWalletID(ctx, req.GetWalletId())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "wallet not found")
		}
		return nil, status.Errorf(codes.Internal, "fetch portfolio: %v", err)
	}

	absExportDir, err := filepath.Abs(s.cfg.ExportDir)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invalid export dir config: %v", err)
	}

	if err := os.MkdirAll(absExportDir, 0o755); err != nil {
		return nil, status.Errorf(codes.Internal, "create export dir: %v", err)
	}

	safeUserID := sanitizeForFilename(req.GetUserId())
	safeWalletID := sanitizeForFilename(req.GetWalletId())
	exportID := uuid.New().String()
	filename := fmt.Sprintf("portfolio_%s_%s_%s", safeUserID, safeWalletID, exportID)
	var data []byte

	if req.GetFormat() == pb.ExportFormat_EXPORT_FORMAT_JSON {
		filename += ".json"
		data, err = json.MarshalIndent(portfolio, "", "  ")
		if err != nil {
			return nil, status.Errorf(codes.Internal, "marshal json: %v", err)
		}
	} else {
		filename += ".csv"
		data, err = s.generateCSV(portfolio)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "generate csv: %v", err)
		}
	}

	fullPath := filepath.Join(absExportDir, filename)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invalid export path: %v", err)
	}

	if !strings.HasPrefix(absPath, absExportDir+string(filepath.Separator)) {
		return nil, status.Error(codes.InvalidArgument, "invalid wallet_id: contains illegal characters")
	}

	if err := os.WriteFile(absPath, data, 0o644); err != nil {
		return nil, status.Errorf(codes.Internal, "write file: %v", err)
	}

	return &pb.ExportResponse{Path: filename}, nil
}

func isValidWalletID(id string) bool {
	if matched, _ := regexp.MatchString(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, strings.ToLower(id)); matched {
		return true
	}
	if matched, _ := regexp.MatchString(`^0x[0-9a-fA-F]{40}$`, id); matched {
		return true
	}
	return false
}

func sanitizeForFilename(input string) string {
	safe := strings.ReplaceAll(input, "..", "_")
	safe = strings.ReplaceAll(safe, "/", "_")
	safe = strings.ReplaceAll(safe, "\\", "_")
	safe = strings.ReplaceAll(safe, "\x00", "_")
	return safe
}

func (s *PortfolioService) generateCSV(portfolio *repository.Portfolio) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)

	if err := w.Write([]string{"wallet_id", "token_address", "symbol", "name", "balance", "usd_value"}); err != nil {
		return nil, err
	}

	for _, asset := range portfolio.Assets {
		if err := w.Write([]string{
			portfolio.WalletID,
			asset.TokenAddress,
			asset.Symbol,
			asset.Name,
			asset.CurrentBalance,
			strconv.FormatFloat(asset.USDValue, 'f', 2, 64),
		}); err != nil {
			return nil, err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *PortfolioService) GetPortfolioSummary(ctx context.Context, req *pb.GetPortfolioSummaryRequest) (*pb.GetPortfolioSummaryResponse, error) {
	wallets, totalNetWorth, err := s.repo.GetPortfolioSummary(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetch portfolio summary: %v", err)
	}

	pbWallets := make([]*pb.Wallet, 0, len(wallets))
	for _, w := range wallets {
		pbWallets = append(pbWallets, &pb.Wallet{
			Id:            w.ID,
			Address:       w.Address,
			Chain:         w.Chain,
			TotalUsdValue: w.TotalUSDValue,
			AssetCount:    w.AssetCount,
		})
	}

	return &pb.GetPortfolioSummaryResponse{
		Wallets:        pbWallets,
		TotalNetWorth:  totalNetWorth,
	}, nil
}

func (s *PortfolioService) GetWalletDetails(ctx context.Context, req *pb.GetWalletDetailsRequest) (*pb.GetWalletDetailsResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	walletIdentifier := req.GetWalletId()
	if walletIdentifier == "" {
		walletIdentifier = req.GetAddress()
	}
	if walletIdentifier == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet_id or address is required")
	}

	if err := s.repo.VerifyWalletOwnership(ctx, req.GetUserId(), walletIdentifier); err != nil {
		return nil, status.Error(codes.PermissionDenied, "wallet not found or access denied")
	}

	details, err := s.repo.GetWalletDetails(ctx, req.GetWalletId(), req.GetAddress())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || err.Error() == "wallet not found" {
			return nil, status.Error(codes.NotFound, "wallet not found")
		}
		return nil, status.Errorf(codes.Internal, "fetch wallet details: %v", err)
	}

	portfolio := &repository.Portfolio{
		WalletID:      details.WalletID,
		Chain:         details.Chain,
		Assets:        details.Assets,
		TotalUSDValue: details.TotalUSDValue,
	}
	if s.marketDataClient != nil {
		if err := s.repo.EnrichPortfolioWithPrices(ctx, portfolio, s.marketDataClient); err != nil {
			log.Printf("WARNING: Failed to enrich wallet %s with prices: %v. Returning wallet with USD values as 0.", details.WalletID, err)
		}
	}

	assets := make([]*pb.Asset, 0, len(portfolio.Assets))
	for _, a := range portfolio.Assets {
		assets = append(assets, &pb.Asset{
			TokenAddress: a.TokenAddress,
			Symbol:       a.Symbol,
			Name:         a.Name,
			Amount:       a.CurrentBalance,
			UsdValue:     a.USDValue,
		})
	}

	return &pb.GetWalletDetailsResponse{
		WalletId:      details.WalletID,
		Address:       details.Address,
		Chain:         details.Chain,
		Assets:        assets,
		TotalUsdValue: portfolio.TotalUSDValue,
	}, nil
}

func (s *PortfolioService) GetTransactionHistory(ctx context.Context, req *pb.GetTransactionHistoryRequest) (*pb.GetTransactionHistoryResponse, error) {
	result, err := s.repo.GetTransactionHistory(
		ctx,
		req.GetWalletId(),
		req.GetAddress(),
		req.GetPage(),
		req.GetLimit(),
		req.GetFilterByType(),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetch transaction history: %v", err)
	}

	transactions := make([]*pb.Transaction, 0, len(result.Transactions))
	for _, t := range result.Transactions {
		transactions = append(transactions, &pb.Transaction{
			Hash:         t.Hash,
			From:         t.From,
			To:           t.To,
			Amount:       t.Amount,
			Symbol:       t.Symbol,
			TokenAddress: t.TokenAddress,
			Timestamp:    t.Timestamp.Unix(),
			Type:         t.Type,
		})
	}

	return &pb.GetTransactionHistoryResponse{
		Transactions: transactions,
		TotalCount:   result.TotalCount,
		Page:         req.GetPage(),
		Limit:        req.GetLimit(),
	}, nil
}
