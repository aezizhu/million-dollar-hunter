package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	pb "github.com/aezizhu/million-dollar-hunter/services/portfolio-service/proto/portfolio/v1"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/config"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/repository"
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
	return &pb.GetPortfolioResponse{
		WalletId: req.GetWalletId(),
		Assets:   []*pb.Asset{},
	}, nil
}

func (s *PortfolioService) Export(ctx context.Context, req *pb.ExportRequest) (*pb.ExportResponse, error) {
	_ = os.MkdirAll(s.cfg.ExportDir, 0o755)
	path := filepath.Join(s.cfg.ExportDir, fmt.Sprintf("portfolio_%s.%s", req.GetWalletId(), req.GetFormat().String()))
	if req.GetFormat() == pb.ExportFormat_EXPORT_FORMAT_JSON {
		if err := os.WriteFile(path, []byte(`{"walletId":"`+req.GetWalletId()+`","assets":[]}`), 0o644); err != nil {
			return nil, err
		}
	} else {
		if err := os.WriteFile(path, []byte("wallet_id,symbol,amount,usd_value\n"), 0o644); err != nil {
			return nil, err
		}
	}
	return &pb.ExportResponse{Path: path}, nil
}
