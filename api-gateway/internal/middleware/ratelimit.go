package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/observability"
	"github.com/aezizhu/million-dollar-hunter/api-gateway/pkg/headers"
)

type Limiter interface {
	Allow(key string) (allowed bool, limit, remaining int, reset time.Time, retryAfter time.Duration)
}

func RateLimit(l Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		baseKey := c.FullPath()
		uid := c.GetString("user_id")
		if uid == "" {
			uid = c.ClientIP()
		}
		key := fmt.Sprintf("%s:%s", baseKey, uid)

		allowed, limit, remaining, reset, retry := l.Allow(key)
		c.Header(headers.RateLimit, strconv.Itoa(limit))
		c.Header(headers.RateRemaining, strconv.Itoa(remaining))
		c.Header(headers.RateReset, strconv.FormatInt(reset.Unix(), 10))
		if !allowed {
			if v, exists := c.Get("http_metrics"); exists {
				if m, ok := v.(*observability.HTTPMetrics); ok && m != nil {
					route := baseKey
					if route == "" {
						route = c.Request.URL.Path
					}
					m.RateLimitBlocked.WithLabelValues(route).Inc()
				}
			}
			c.Header(headers.RetryAfter, strconv.Itoa(int(retry.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate_limit", "message": "rate limit exceeded"})
			return
		}
		if v, exists := c.Get("http_metrics"); exists {
			if m, ok := v.(*observability.HTTPMetrics); ok && m != nil {
				route := baseKey
				if route == "" {
					route = c.Request.URL.Path
				}
				m.RateLimitAllowed.WithLabelValues(route).Inc()
			}
		}
		c.Next()
	}
}
