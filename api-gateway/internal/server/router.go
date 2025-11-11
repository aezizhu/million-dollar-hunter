// Package server provides HTTP router and middleware setup for the API Gateway.
// Copyright (c) 2025 aezizhu. All rights reserved.
// Author: aezizhu
// Repository: github.com/aezizhu/million-dollar-hunter
package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"google.golang.org/grpc"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/clients"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/handlers"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/middleware"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/observability"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/ratelimit"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/pkg/headers"
)

func newHierLimiter(cfg config.Config, logger zerolog.Logger) *ratelimit.HierarchicalLimiter {
	var rdb *redis.Client
	var ipLim, userLim, routeBase ratelimit.SimpleLimiter

	if cfg.RedisURL != "" {
		opts := &redis.Options{Addr: cfg.RedisURL}
		rdb = redis.NewClient(opts)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			logger.Warn().Err(err).Msg("Redis unavailable, falling back to local rate limiter")
		} else {
			logger.Info().Msg("Using Redis for distributed rate limiting")
		}
	}

	ipRPS := cfg.IPRateLimitRPS
	if ipRPS == 0 {
		ipRPS = cfg.RateDefaultRPS
	}
	ipBurst := cfg.IPRateLimitBurst
	if ipBurst == 0 {
		ipBurst = cfg.RateDefaultBurst
	}
	userRPS := cfg.UserRateLimitRPS
	if userRPS == 0 {
		userRPS = cfg.RateDefaultRPS
	}
	userBurst := cfg.UserRateLimitBurst
	if userBurst == 0 {
		userBurst = cfg.RateDefaultBurst
	}

	if rdb != nil {
		ipLim = ratelimit.NewRedisTokenBucket(rdb, ipRPS, ipBurst, time.Second, "ratelimit")
		userLim = ratelimit.NewRedisTokenBucket(rdb, userRPS, userBurst, time.Second, "ratelimit")
		routeBase = ratelimit.NewRedisTokenBucket(rdb, cfg.RateDefaultRPS, cfg.RateDefaultBurst, time.Second, "ratelimit")
	} else {
		ipLim = ratelimit.LocalAdapter{Inner: ratelimit.NewLocalTokenBucket(ipRPS, ipBurst, time.Second)}
		userLim = ratelimit.LocalAdapter{Inner: ratelimit.NewLocalTokenBucket(userRPS, userBurst, time.Second)}
		routeBase = ratelimit.LocalAdapter{Inner: ratelimit.NewLocalTokenBucket(cfg.RateDefaultRPS, cfg.RateDefaultBurst, time.Second)}
	}

	overrides, _ := ratelimit.ParseRouteLimitsJSON(cfg.RouteLimitsJSON)
	routeLim := routeBase
	if len(overrides) > 0 {
		byKey := map[string]ratelimit.SimpleLimiter{}
		for route, lim := range overrides {
			if rdb != nil {
				byKey[route] = ratelimit.NewRedisTokenBucket(rdb, lim.RPS, lim.Burst, time.Second, "ratelimit")
			} else {
				byKey[route] = ratelimit.LocalAdapter{Inner: ratelimit.NewLocalTokenBucket(lim.RPS, lim.Burst, time.Second)}
			}
		}
		routeLim = ratelimit.NewMultiLimiter(routeBase, byKey)
	}

	return ratelimit.NewHierarchicalLimiter(ipLim, userLim, routeLim, cfg.RateLimitAllowlist)
}

func Register(r *gin.Engine, cfg config.Config, logger zerolog.Logger, reg *prometheus.Registry) *clients.GRPCClients {
	if cfg.AuthValidateMode == "grpc" {
		if cfg.AuthGRPCAddr == "" {
			logger.Fatal().Msg("AUTH_GRPC_ADDR is required when AUTH_VALIDATE_MODE=grpc")
		}
		if cfg.JWTAudience == "" {
			logger.Fatal().Msg("JWT_AUDIENCE is required when AUTH_VALIDATE_MODE=grpc")
		}
	}
	cfg.Validate(logger)
	grpcClients := clients.NewGRPCClients(cfg.PortfolioServiceURL, cfg.MarketDataServiceURL, cfg.AuthGRPCAddr, logger)

	httpMetrics := observability.NewHTTPMetrics(reg, cfg.PrometheusNamespace)
	authMetrics := observability.NewAuthGRPCMetrics(reg, cfg.PrometheusNamespace)
	r.Use(func(c *gin.Context) {
		c.Set("http_metrics", httpMetrics)
		c.Set("auth_grpc_metrics", authMetrics)
		c.Next()
	})
	r.SetTrustedProxies([]string{"127.0.0.1/32"})
	r.Use(func(c *gin.Context) {
		c.Header("Vary", "Origin")
		c.Next()
	})

	// Parse comma-separated origins
	allowedOrigins := []string{}
	if cfg.FrontendURL != "" && cfg.FrontendURL != "*" {
		origins := strings.Split(cfg.FrontendURL, ",")
		for _, origin := range origins {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowedOrigins = append(allowedOrigins, origin)
			}
		}
	}
	// If no valid origins, use empty slice (CORS will reject all)
	if len(allowedOrigins) == 0 && cfg.FrontendURL != "*" {
		allowedOrigins = []string{}
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", headers.RequestID},
		ExposeHeaders:    []string{headers.RateLimit, headers.RateRemaining, headers.RateReset, headers.RetryAfter},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(middleware.SecurityHeaders(cfg))

	r.Use(func(c *gin.Context) {
		rid := c.GetHeader(headers.RequestID)
		if rid == "" {
			rid = time.Now().UTC().Format("20060102150405.000000000")
		}
		c.Set("request_id", rid)
		c.Header(headers.RequestID, rid)
		c.Next()
	})

	r.Use(middleware.Logging(logger))

	hierLimiter := newHierLimiter(cfg, logger)

	r.GET("/healthz", func(c *gin.Context) {
		health := gin.H{"ok": true}
		status := http.StatusOK
		if cfg.RedisURL != "" {
			opts := &redis.Options{Addr: cfg.RedisURL}
			rdb := redis.NewClient(opts)
			ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
			defer cancel()
			if err := rdb.Ping(ctx).Err(); err != nil {
				health["redis"] = "unhealthy"
				status = http.StatusServiceUnavailable
			} else {
				health["redis"] = "ok"
			}
		}
		c.JSON(status, health)
	})
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))
	r.OPTIONS("/*path", func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == cfg.FrontendURL {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Vary", "Origin")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, "+headers.RequestID)
		c.Status(http.StatusNoContent)
	})

	r.POST("/api/v1/auth/login", handlers.Login(cfg))
	r.POST("/api/v1/auth/refresh", handlers.Refresh(cfg))

	api := r.Group("/api/v1")
	// Auth must run before rate limiting so unauthenticated requests do not consume rate limit quota.
	api.Use(middleware.Metrics(httpMetrics))
	api.Use(middleware.Auth(cfg, grpcClients.AuthConn))
	api.Use(middleware.RateLimitHier(hierLimiter, cfg))
	api.Use(middleware.Tracing())

	var portfolioConn, marketDataConn *grpc.ClientConn
	if grpcClients != nil {
		portfolioConn = grpcClients.PortfolioConn
		marketDataConn = grpcClients.MarketDataConn
	}

	api.GET("/portfolios", handlers.ListPortfolios(portfolioConn, logger))
	api.POST("/portfolios", handlers.AddWallet(portfolioConn, logger))
	api.GET("/wallets/:address", handlers.GetWallet(portfolioConn, logger))
	api.GET("/wallets/:address/transactions", handlers.GetTransactions(portfolioConn, logger))
	api.GET("/tokens/:tokenAddress/holders", handlers.TopHolders(marketDataConn, logger))
	api.GET("/export/wallet/:address", handlers.ExportWallet(portfolioConn, logger))
	return grpcClients
}
