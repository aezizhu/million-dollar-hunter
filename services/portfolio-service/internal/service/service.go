package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	pb "github.com/aezizhu/million-dollar-hunter/services/portfolio-service/proto/portfolio/v1"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/config"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PortfolioService struct {
	repo *repository.Repo
	cfg  config.Config
}

func New(repo *repository.Repo, cfg config.Config) *PortfolioService {
	return &PortfolioService{repo: repo, cfg: cfg}
}

func (s *PortfolioService) HandleTransactionDataIngested(ctx context.Context, raw []byte) error {
	return s.repo.UpsertFromIngest(ctx, raw)
}

func (s *PortfolioService) GetPortfolio(ctx context.Context, req *pb.GetPortfolioRequest) (*pb.GetPortfolioResponse, error) {
	if req.GetWalletId() == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet_id is required")
	}

	portfolio, err := s.repo.GetPortfolioByWalletID(ctx, req.GetWalletId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetch portfolio: %v", err)
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
	if req.GetWalletId() == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet_id is required")
	}

	portfolio, err := s.repo.GetPortfolioByWalletID(ctx, req.GetWalletId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "wallet not found: %v", err)
	}

	if err := os.MkdirAll(s.cfg.ExportDir, 0o755); err != nil {
		return nil, status.Errorf(codes.Internal, "create export dir: %v", err)
	}

	filename := fmt.Sprintf("portfolio_%s_%d", req.GetWalletId(), time.Now().Unix())
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

	path := filepath.Join(s.cfg.ExportDir, filename)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return nil, status.Errorf(codes.Internal, "write file: %v", err)
	}

	return &pb.ExportResponse{Path: filename}, nil
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
