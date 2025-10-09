package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/handlers"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/middleware"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/ratelimit"
	"github.com/rs/zerolog"
)

const (
	HeaderRateLimit     = "X-RateLimit-Limit"
	HeaderRateRemaining = "X-RateLimit-Remaining"
	HeaderRateReset     = "X-RateLimit-Reset"
	HeaderRetryAfter    = "Retry-After"
)

func Register(r *gin.Engine, cfg config.Config, logger zerolog.Logger, reg *prometheus.Registry) {
	limiter := ratelimit.NewLocalTokenBucket(cfg.RateDefaultRPS, cfg.RateDefaultBurst, time.Second)

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	r.POST("/api/v1/auth/login", handlers.Login(cfg))
	r.POST("/api/v1/auth/refresh", handlers.Refresh())

	api := r.Group("/api/v1")
	api.Use(middleware.Auth(cfg))
	api.Use(middleware.RateLimit(limiter))

	api.GET("/portfolios", handlers.ListPortfolios())
	api.POST("/portfolios", handlers.AddWallet())
	api.GET("/wallets/:address", handlers.GetWallet())
	api.GET("/wallets/:address/transactions", handlers.GetTransactions())
	api.GET("/tokens/:tokenAddress/holders", handlers.TopHolders())
	api.GET("/export/wallet/:address", handlers.ExportWallet())
}
