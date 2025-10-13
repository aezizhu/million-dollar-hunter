package server

import (
	"context"
	"net/http"
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


func newLimiter(cfg config.Config, logger zerolog.Logger) middleware.Limiter {
	var (
		base ratelimit.SimpleLimiter
		rdb  *redis.Client
	)
	if cfg.RedisURL != "" {
		opts := &redis.Options{Addr: cfg.RedisURL}
		rdb = redis.NewClient(opts)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			logger.Warn().Err(err).Msg("Redis unavailable, falling back to local rate limiter")
		} else {
			logger.Info().Msg("Using Redis for distributed rate limiting")
			base = ratelimit.NewRedisTokenBucket(rdb, cfg.RateDefaultRPS, cfg.RateDefaultBurst, time.Second, "ratelimit")
		}
	}
	if base == nil {
		local := ratelimit.NewLocalTokenBucket(cfg.RateDefaultRPS, cfg.RateDefaultBurst, time.Second)
		base = ratelimit.LocalAdapter{Inner: local}
	}

	overrides, _ := ratelimit.ParseRouteLimitsJSON(cfg.RouteLimitsJSON)
	byKey := map[string]ratelimit.SimpleLimiter{}
	if len(overrides) > 0 {
		for route, lim := range overrides {
			if rdb != nil {
				byKey[route] = ratelimit.NewRedisTokenBucket(rdb, lim.RPS, lim.Burst, time.Second, "ratelimit")
			} else {
				byKey[route] = ratelimit.LocalAdapter{Inner: ratelimit.NewLocalTokenBucket(lim.RPS, lim.Burst, time.Second)}
			}
		}
	}
	ml := ratelimit.NewMultiLimiter(base, byKey)
	return limiterAdapter{ml}
}

type limiterAdapter struct {
	rl interface {
		Allow(ctx context.Context, key string) (bool, int, int, time.Time, time.Duration)
	}
}

func (l limiterAdapter) Allow(key string) (bool, int, int, time.Time, time.Duration) {
	return l.rl.Allow(context.Background(), key)
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
	grpcClients := clients.NewGRPCClients(cfg.PortfolioServiceURL, cfg.MarketDataServiceURL, cfg.AuthGRPCAddr, logger)

	httpMetrics := observability.NewHTTPMetrics(reg, cfg.PrometheusNamespace)
	authMetrics := observability.NewAuthGRPCMetrics(reg, cfg.PrometheusNamespace)
	r.Use(func(c *gin.Context) {
		c.Set("http_metrics", httpMetrics)
		c.Set("auth_grpc_metrics", authMetrics)
		c.Next()
	})

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
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

	limiter := newLimiter(cfg, logger)

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

	r.POST("/api/v1/auth/login", handlers.Login(cfg))
	r.POST("/api/v1/auth/refresh", handlers.Refresh(cfg))

	api := r.Group("/api/v1")
	api.Use(middleware.Metrics(httpMetrics))
	api.Use(middleware.RateLimit(limiter))
	api.Use(middleware.Auth(cfg, grpcClients.AuthConn))
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
