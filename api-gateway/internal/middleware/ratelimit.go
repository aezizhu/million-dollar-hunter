package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	headerRateLimit     = "X-RateLimit-Limit"
	headerRateRemaining = "X-RateLimit-Remaining"
	headerRateReset     = "X-RateLimit-Reset"
	headerRetryAfter    = "Retry-After"
)

type Limiter interface {
	Allow(key string) (allowed bool, limit, remaining int, reset time.Time, retryAfter time.Duration)
}

func RateLimit(l Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.FullPath()
		allowed, limit, remaining, reset, retry := l.Allow(key)
		c.Header(headerRateLimit, strconv.Itoa(limit))
		c.Header(headerRateRemaining, strconv.Itoa(remaining))
		c.Header(headerRateReset, strconv.FormatInt(reset.Unix(), 10))
		if !allowed {
			c.Header(headerRetryAfter, strconv.Itoa(int(retry.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate_limit", "message": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
