package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"api-gateway/internal/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type RateLimitMiddleware struct {
	limiter *ratelimit.RateLimiter
	logger  zerolog.Logger
}

func NewRateLimitMiddleware(limiter *ratelimit.RateLimiter, logger zerolog.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: limiter,
		logger:  logger,
	}
}

func (r *RateLimitMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			identifier = fmt.Sprintf("user:%v", userID)
		}

		allowed, retryAfter, remaining, resetTime, err := r.limiter.Allow(c.Request.Context(), identifier)
		if err != nil {
			r.logger.Error().Err(err).Str("identifier", identifier).Msg("rate limit check failed")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "rate limit check failed",
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(r.limiter.Capacity()))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

		if !allowed {
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			r.logger.Warn().
				Str("identifier", identifier).
				Dur("retry_after", retryAfter).
				Msg("rate limit exceeded")
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "too many requests",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
