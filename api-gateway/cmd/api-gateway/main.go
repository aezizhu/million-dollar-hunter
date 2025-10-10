package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-gateway/internal/config"
	"api-gateway/internal/handlers"
	"api-gateway/internal/logging"
	"api-gateway/internal/metrics"
	"api-gateway/internal/middleware"
	"api-gateway/internal/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := logging.New(cfg.LogLevel)
	logger.Info().Msg("starting api-gateway")

	metrics.Init()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer rdb.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to redis")
	}
	logger.Info().Str("addr", cfg.RedisAddr).Msg("connected to redis")

	rateLimiter := ratelimit.New(rdb, cfg.RateLimitPerMinute/60, cfg.RateLimitBurst)

	authMiddleware := middleware.NewAuthMiddleware(cfg, logger)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(rateLimiter, logger)

	healthHandler := handlers.NewHealthHandler(rdb, logger)
	authHandler := handlers.NewAuthHandler(authMiddleware, logger)
	portfolioHandler := handlers.NewPortfolioHandler(logger)
	walletHandler := handlers.NewWalletHandler(logger)
	tokenHandler := handlers.NewTokenHandler(logger)
	exportHandler := handlers.NewExportHandler(logger)

	if cfg.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(requestLogger(logger))
	router.Use(metrics.PrometheusMiddleware())
	router.Use(corsMiddleware())

	router.GET("/health", healthHandler.Health)
	router.GET("/healthz", healthHandler.Liveness)
	router.GET("/ready", healthHandler.Readiness)

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.Refresh)
		}

		protected := v1.Group("")
		protected.Use(authMiddleware.JWTAuth())
		protected.Use(rateLimitMiddleware.RateLimit())
		{
			protected.GET("/portfolios", portfolioHandler.ListPortfolios)
			protected.POST("/portfolios", portfolioHandler.AddWallet)

			protected.GET("/wallets/:address", walletHandler.GetWallet)
			protected.GET("/wallets/:address/transactions", walletHandler.GetTransactions)

			protected.GET("/tokens/:tokenAddress/holders", tokenHandler.GetTopHolders)

			protected.GET("/export/wallet/:address", exportHandler.ExportWallet)
		}
	}

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info().Str("port", cfg.HTTPPort).Msg("api-gateway listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("server forced to shutdown")
	}

	logger.Info().Msg("server exited")
}

func requestLogger(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logEvent := logger.Info()
		if statusCode >= 500 {
			logEvent = logger.Error()
		} else if statusCode >= 400 {
			logEvent = logger.Warn()
		}

		logEvent.
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", query).
			Int("status", statusCode).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Msg("request completed")
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
