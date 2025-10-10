package handlers

import (
	"context"
	"net/http"
	"time"

	"api-gateway/internal/metrics"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type HealthHandler struct {
	redis  *redis.Client
	logger zerolog.Logger
}

func NewHealthHandler(redis *redis.Client, logger zerolog.Logger) *HealthHandler {
	return &HealthHandler{
		redis:  redis,
		logger: logger,
	}
}

type HealthResponse struct {
	Status       string                 `json:"status"`
	Version      string                 `json:"version"`
	Dependencies map[string]DepStatus   `json:"dependencies"`
}

type DepStatus struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	deps := make(map[string]DepStatus)
	overallHealthy := true

	redisStart := time.Now()
	if err := h.redis.Ping(ctx).Err(); err != nil {
		deps["redis"] = DepStatus{
			Status: "unhealthy",
			Error:  err.Error(),
		}
		metrics.DependencyUp.WithLabelValues("redis").Set(0)
		overallHealthy = false
		h.logger.Error().Err(err).Msg("redis health check failed")
	} else {
		deps["redis"] = DepStatus{
			Status:  "healthy",
			Latency: time.Since(redisStart).String(),
		}
		metrics.DependencyUp.WithLabelValues("redis").Set(1)
	}


	status := "healthy"
	statusCode := http.StatusOK
	if !overallHealthy {
		status = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, HealthResponse{
		Status:       status,
		Version:      "0.1.0",
		Dependencies: deps,
	})
}

func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.redis.Ping(ctx).Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}
