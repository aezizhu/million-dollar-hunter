package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/observability"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/ratelimit"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/pkg/headers"
)

func RateLimitHier(h *ratelimit.HierarchicalLimiter, cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}
		uid := c.GetString("user_id")
		ip := c.ClientIP()
		bypass := false
		if cfg.RateLimitBypassHeader != "" && c.GetHeader(cfg.RateLimitBypassHeader) != "" && h.InAllowlist(ip) {
			bypass = true
		}

		dec := h.Allow(c.Request.Context(), ip, uid, route, bypass)

		c.Header(headers.RateLimit, strconv.Itoa(dec.Limit))
		c.Header(headers.RateRemaining, strconv.Itoa(dec.Remaining))
		c.Header(headers.RateReset, strconv.FormatInt(dec.Reset.Unix(), 10))

		if !dec.Allowed {
			if v, exists := c.Get("http_metrics"); exists {
				if m, ok := v.(*observability.HTTPMetrics); ok && m != nil {
					m.RateLimitBlocked.WithLabelValues(route, string(dec.Dimension)).Inc()
					m.HierarchicalDenials.WithLabelValues(route, string(dec.Dimension)).Inc()
					if dec.Dimension == ratelimit.DimIP {
						m.ViolationsByIP.Inc()
					}
				}
			}
			c.Header(headers.RetryAfter, strconv.Itoa(int(dec.RetryAfter.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit",
				"message": "rate limit exceeded",
				"details": gin.H{"dimension": string(dec.Dimension)},
			})
			return
		}
		if v, exists := c.Get("http_metrics"); exists {
			if m, ok := v.(*observability.HTTPMetrics); ok && m != nil {
				m.RateLimitAllowed.WithLabelValues(route, "none").Inc()
			}
		}
		c.Next()
	}
}
