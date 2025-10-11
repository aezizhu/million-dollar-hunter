package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"google.golang.org/grpc"

	gen "github.com/aezizhu/million-dollar-hunter/services/auth-service/api/gen"
	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/config"
	httpapi "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/http"
	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
	grpcserver "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/grpc"
	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/store"
)

func main() {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.Parse()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse config")
	}

	multiUser := false
	if val := os.Getenv("ENABLE_MULTI_USER"); val != "" {
		switch strings.ToLower(val) {
		case "true", "1", "yes":
			multiUser = true
		case "false", "0", "no":
			multiUser = false
		default:
			log.Fatal().Str("value", val).Msg("ENABLE_MULTI_USER must be true/false/1/0/yes/no")
		}
	}

	if !multiUser {
		if os.Getenv("MVP_USERNAME") == "" {
			log.Fatal().Msg("MVP_USERNAME is required in MVP mode")
		}
		mvpPass := os.Getenv("MVP_PASSWORD")
		if mvpPass == "" {
			log.Fatal().Msg("MVP_PASSWORD is required in MVP mode")
		}
		if len(mvpPass) != 60 || (!strings.HasPrefix(mvpPass, "$2a$") && !strings.HasPrefix(mvpPass, "$2b$") && !strings.HasPrefix(mvpPass, "$2y$")) {
			log.Warn().Int("length", len(mvpPass)).
				Msg("MVP_PASSWORD may not be a valid bcrypt hash (expected 60 chars starting with $2a/$2b/$2y)")
		}
		log.Info().Msg("Running in MVP mode")
	} else {
		if os.Getenv("DATABASE_URL") == "" {
			log.Fatal().Msg("DATABASE_URL is required in multi-user mode")
		}
		log.Info().Msg("Running in multi-user mode")
	}

	j := jwtmgr.New(cfg.JWTIssuer, cfg.JWTAudience, cfg.AccessTTL, cfg.RefreshTTL, cfg.JWTSigningKey)

	mux := http.NewServeMux()
	s := &httpapi.Server{Logger: &log, JWT: j}
	if multiUser {
		dsn := os.Getenv("DATABASE_URL")
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
	gen.RegisterAuthServiceServer(grpcSrv, grpcserver.New(j))

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
