package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func Tracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		tr := otel.Tracer("api-gateway")
		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}
		ctx, span := tr.Start(c.Request.Context(), route, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		c.Request = c.Request.WithContext(ctx)
		start := time.Now()
		c.Next()
		dur := time.Since(start)

		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.route", route),
			attribute.String("http.target", c.Request.URL.Path),
		)
		span.SetAttributes(attribute.Int("http.status_code", c.Writer.Status()))
		if len(c.Errors) > 0 || c.Writer.Status() >= 500 {
			span.SetStatus(codes.Error, "request error")
		}
		span.SetAttributes(attribute.Float64("http.server_duration_ms", float64(dur.Milliseconds())))
	}
}
