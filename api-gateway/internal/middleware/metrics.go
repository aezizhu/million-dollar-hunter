package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/observability"
)

func Metrics(m *observability.HTTPMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		if m != nil {
			m.InFlight.Inc()
		}
		start := time.Now()
		c.Next()
		status := strconv.Itoa(c.Writer.Status())
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}
		observability.Observe(m, c.Request.Method, route, status, start)
	}
}
