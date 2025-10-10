package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/config"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/logging"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/repository"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/service"
)

func main() {
	_ = os.Setenv("TZ", "UTC")
	ctx := context.Background()

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

	svc, err := service.New(ctx, cfg, logger, db)
	if err != nil {
		logger.Fatal().Err(err).Msg("service init")
	}

	go svc.StartWorkers(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info().Str("addr", srv.Addr).Msg("listening")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal().Err(err).Msg("http server")
	}
}
