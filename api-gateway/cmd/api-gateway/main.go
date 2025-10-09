package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/observability"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/server"
)

func main() {
	cfg := config.Load()

	logger := observability.InitLogger(cfg)
	tp := observability.InitTracing(cfg, logger)
	defer func() {
		if tp != nil {
			_ = tp.Shutdown(context.Background())
		}
	}()
	reg := observability.InitMetricsRegistry(cfg)

	engine := gin.New()
	engine.Use(gin.Recovery())

	server.Register(engine, cfg, logger, reg)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api-gateway listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
