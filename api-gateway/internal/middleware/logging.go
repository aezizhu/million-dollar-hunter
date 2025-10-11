package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func Logging(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

	traceIDVal, _ := c.Get("request_id")
		traceID := ""
		if tid, ok := traceIDVal.(string); ok {
			traceID = tid
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		event := logger.Info()
		if status >= 400 {
			event = logger.Error()
		}

		if raw != "" {
			path = path + "?" + raw
		}

		event.
			Str("trace_id", traceID).
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Dur("latency", latency).
			Str("client_ip", clientIP).
			Str("user_agent", userAgent).
			Msg("HTTP request")

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error().
					Str("trace_id", traceID).
					Str("path", path).
					Err(e.Err).
					Msg("Request error")
			}
		}
	}
}
