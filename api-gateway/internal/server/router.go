package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/handlers"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/middleware"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/observability"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/ratelimit"
)

const (
	HeaderRateLimit     = "X-RateLimit-Limit"
	HeaderRateRemaining = "X-RateLimit-Remaining"
	HeaderRateReset     = "X-RateLimit-Reset"
	HeaderRetryAfter    = "Retry-After"
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
		if err := rdb.Ping(ctx).Err(); err == nil {
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

func Register(r *gin.Engine, cfg config.Config, logger zerolog.Logger, reg *prometheus.Registry) {
	httpMetrics := observability.NewHTTPMetrics(reg, cfg.PrometheusNamespace)
	r.Use(func(c *gin.Context) {
		c.Set("http_metrics", httpMetrics)
		c.Next()
	})

	limiter := newLimiter(cfg, logger)

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	r.POST("/api/v1/auth/login", handlers.Login(cfg))
	r.POST("/api/v1/auth/refresh", handlers.Refresh())

	api := r.Group("/api/v1")
	api.Use(middleware.Auth(cfg))
	api.Use(middleware.RateLimit(limiter))
	api.Use(middleware.Metrics(httpMetrics))
	api.Use(middleware.Tracing())

	api.GET("/portfolios", handlers.ListPortfolios())
	api.POST("/portfolios", handlers.AddWallet())
	api.GET("/wallets/:address", handlers.GetWallet())
	api.GET("/wallets/:address/transactions", handlers.GetTransactions())
	api.GET("/tokens/:tokenAddress/holders", handlers.TopHolders())
	api.GET("/export/wallet/:address", handlers.ExportWallet())
}
