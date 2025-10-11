package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/config"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/logging"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/repository"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	_ = os.Setenv("TZ", "UTC")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	logger := logging.New(cfg)

	db, err := repository.NewPostgres(ctx, cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("db connect")
	}
	defer db.Close()

	if err := repository.RunMigrations(ctx, cfg, logger); err != nil {
		logger.Fatal().Err(err).Msg("migrations")
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	svc, err := service.New(ctx, cfg, logger, db, reg)
	if err != nil {
		logger.Fatal().Err(err).Msg("service init")
	}
	defer func() {
		if err := svc.Close(); err != nil {
			logger.Error().Err(err).Msg("service close")
		}
	}()

	go svc.StartWorkers(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", svc.HealthCheckHandler)
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info().Str("addr", srv.Addr).Msg("listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("http server")
		}
	}()

	<-sigChan
	logger.Info().Msg("received shutdown signal, gracefully shutting down")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("http server shutdown")
	}

	logger.Info().Msg("shutdown complete")
}
