package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"google.golang.org/grpc"

	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/config"
	httpapi "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/http"
	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
)

func main() {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, _ := config.Parse()
	j := jwtmgr.New(cfg.JWTIssuer, cfg.JWTAudience, cfg.AccessTTL, cfg.RefreshTTL, cfg.JWTSigningKey)

	mux := http.NewServeMux()
	s := &httpapi.Server{Logger: &log, JWT: j}
	mux.HandleFunc("/healthz", s.Health)
	mux.HandleFunc("/api/v1/auth/login", s.Login)
	mux.HandleFunc("/api/v1/auth/logout", s.Logout)
	mux.HandleFunc("/api/v1/auth/refresh", s.Refresh)

	httpSrv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: hlog.NewHandler(log)(mux),
	}

	grpcSrv := grpc.NewServer()
	grpcLn, _ := net.Listen("tcp", ":"+cfg.GRPCPort)

	go func() {
		_ = httpSrv.ListenAndServe()
	}()
	go func() {
		_ = grpcSrv.Serve(grpcLn)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)
	grpcSrv.GracefulStop()
}
