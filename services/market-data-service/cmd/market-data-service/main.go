// Package main provides the market data service entry point.
// Copyright (c) 2025 aezizhu. All rights reserved.
// Author: aezizhu
// Repository: github.com/aezizhu/million-dollar-hunter
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/internal/cache"
	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/internal/client"
	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/internal/config"
	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/internal/handler"
	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/internal/repository"
	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/internal/worker"
	"github.com/aezizhu/million-dollar-hunter/services/market-data-service/pkg/pb"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func main() {
	// Application entry point for market data service
	// Ensures proper initialization of Redis cache and CoinGecko client
	// Zero-configuration approach with environment-based settings
	// Initializes background price refresh workers
	// Zero-downtime deployment with graceful shutdown
	// Handles gRPC requests for real-time price data
	// Unified caching strategy with TTL management
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("service", "market-data-service").
		Logger()

	logger.Info().Msg("Starting market-data-service")

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	logger.Info().
		Str("grpc_port", cfg.Server.GRPCPort).
		Str("http_port", cfg.Server.HTTPPort).
		Msg("Configuration loaded")

	var repo *repository.Repository
	dbHost := os.Getenv("DB_HOST")
	if dbHost != "" {
		var err error
		repo, err = repository.NewRepository(cfg.Database.ConnectionString(), logger)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to initialize repository")
		}
		defer repo.Close()
		logger.Info().Msg("Database repository initialized")
	} else {
		logger.Info().Msg("Database repository disabled (DB_HOST not set) - using Redis-only mode")
	}

	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	redisCache, err := cache.NewRedisCache(
		redisAddr,
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Redis.TTL,
		logger,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize Redis cache")
	}
	defer redisCache.Close()

	coinGeckoClient := client.NewCoinGeckoClient(&cfg.CoinGecko, logger)

	priceWorker := worker.NewPriceRefresher(
		coinGeckoClient,
		redisCache,
		repo,
		&cfg.Worker,
		logger,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	priceWorker.Start(ctx)
	defer priceWorker.Stop()

	grpcHandler := handler.NewGRPCHandler(
		coinGeckoClient,
		redisCache,
		repo,
		priceWorker,
		logger,
	)

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10 * 1024 * 1024), // 10MB
		grpc.MaxSendMsgSize(10 * 1024 * 1024), // 10MB
	)

	pb.RegisterMarketDataServiceServer(grpcServer, grpcHandler)

	grpcListener, err := net.Listen("tcp", ":"+cfg.Server.GRPCPort)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create gRPC listener")
	}

	go func() {
		logger.Info().
			Str("port", cfg.Server.GRPCPort).
			Msg("Starting gRPC server")

		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Fatal().Err(err).Msg("Failed to serve gRPC")
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info().Msg("Shutdown signal received")

	cancel()
	
	logger.Info().Msg("Stopping gRPC server")
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		logger.Info().Msg("Server stopped gracefully")
	case <-time.After(30 * time.Second):
		logger.Warn().Msg("Shutdown timeout, forcing stop")
		grpcServer.Stop()
	}

	logger.Info().Msg("market-data-service shutdown complete")
}
