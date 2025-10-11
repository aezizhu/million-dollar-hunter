package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"google.golang.org/grpc"

	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/config"
	httpapi "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/http"
	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/store"
)

func main() {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.Parse()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse config")
	}

	if os.Getenv("ENABLE_MULTI_USER") != "true" {
		if os.Getenv("MVP_USERNAME") == "" {
			log.Fatal().Msg("MVP_USERNAME environment variable is required when ENABLE_MULTI_USER is false")
		}
		if os.Getenv("MVP_PASSWORD") == "" {
			log.Fatal().Msg("MVP_PASSWORD environment variable is required when ENABLE_MULTI_USER is false")
		}
		log.Info().Msg("Running in MVP mode with hardcoded credentials")
	} else {
		if os.Getenv("DATABASE_URL") == "" {
			log.Fatal().Msg("DATABASE_URL environment variable is required when ENABLE_MULTI_USER is true")
		}
		log.Info().Msg("Running in multi-user mode with database authentication")
	}

	j := jwtmgr.New(cfg.JWTIssuer, cfg.JWTAudience, cfg.AccessTTL, cfg.RefreshTTL, cfg.JWTSigningKey)

	mux := http.NewServeMux()
	s := &httpapi.Server{Logger: &log, JWT: j}
	if os.Getenv("ENABLE_MULTI_USER") == "true" {
		if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
			pool, err := pgxpool.New(context.Background(), dsn)
			if err != nil {
				log.Fatal().Err(err).Msg("database connection failed")
			}
			pg := &store.PGStore{Pool: pool}
			s.Store = pg
			s.RefreshTokens = pg
			s.Audit = pg
			defer pool.Close()
		}
	}
	mux.HandleFunc("/healthz", s.Health)
	mux.HandleFunc("/api/v1/auth/login", s.Login)
	mux.HandleFunc("/api/v1/auth/register", s.Register)
	mux.HandleFunc("/api/v1/auth/logout", s.Logout)
	mux.HandleFunc("/api/v1/auth/refresh", s.Refresh)

	httpSrv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: hlog.NewHandler(log)(mux),
	}

	grpcSrv := grpc.NewServer()
	grpcLn, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start gRPC listener")
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()
	go func() {
		if err := grpcSrv.Serve(grpcLn); err != nil {
			log.Fatal().Err(err).Msg("gRPC server failed")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)
	grpcSrv.GracefulStop()
}
