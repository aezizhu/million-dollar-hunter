package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v11"
	"google.golang.org/grpc"

	pb "github.com/aezizhu/million-dollar-hunter/services/portfolio-service/proto/portfolio/v1"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/config"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/kafka"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/repository"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/service"
)

type server struct {
	pb.UnimplementedPortfolioServiceServer
	svc *service.PortfolioService
}

func (s *server) GetPortfolio(ctx context.Context, req *pb.GetPortfolioRequest) (*pb.GetPortfolioResponse, error) {
	return s.svc.GetPortfolio(ctx, req)
}

func (s *server) GetPortfolioSummary(ctx context.Context, req *pb.GetPortfolioSummaryRequest) (*pb.GetPortfolioSummaryResponse, error) {
	return s.svc.GetPortfolioSummary(ctx, req)
}

func (s *server) GetWalletDetails(ctx context.Context, req *pb.GetWalletDetailsRequest) (*pb.GetWalletDetailsResponse, error) {
	return s.svc.GetWalletDetails(ctx, req)
}

func (s *server) GetTransactionHistory(ctx context.Context, req *pb.GetTransactionHistoryRequest) (*pb.GetTransactionHistoryResponse, error) {
	return s.svc.GetTransactionHistory(ctx, req)
}

func (s *server) Export(ctx context.Context, req *pb.ExportRequest) (*pb.ExportResponse, error) {
	return s.svc.Export(ctx, req)
}

func main() {
	var cfg config.Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("env parse: %v", err)
	}

	poolCfg := repository.PoolConfig{
		MaxConns:          cfg.DBMaxConns,
		MinConns:          cfg.DBMinConns,
		MaxConnLifetime:   cfg.DBMaxConnLifetime,
		MaxConnIdleTime:   cfg.DBMaxConnIdleTime,
		HealthCheckPeriod: cfg.DBHealthCheckPeriod,
	}

	db, err := repository.NewPostgres(cfg.DatabaseURL, poolCfg)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close(context.Background())

	svc := service.New(db, cfg)

	consumer, err := kafka.NewConsumer(cfg, svc)
	if err != nil {
		log.Fatalf("kafka consumer: %v", err)
	}
	go consumer.Run()

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterPortfolioServiceServer(grpcServer, &server{svc: svc})

	go func() {
		log.Printf("gRPC listening on %s", cfg.GRPCAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("serve: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	grpcServer.GracefulStop()
	consumer.Stop()
}
