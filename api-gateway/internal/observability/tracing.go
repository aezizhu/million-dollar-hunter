package observability

import (
	"context"
	"os"

	"github.com/rs/zerolog"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type TracerProvider interface {
	Shutdown(ctx context.Context) error
}

func InitTracing(cfg config.Config, logger zerolog.Logger) TracerProvider {
	var tp *sdktrace.TracerProvider

	res, _ := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("api-gateway"),
		),
	)

	if cfg.OTLPEndpoint != "" {
		exp, err := otlptracehttp.New(context.Background(),
			otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
			otlptracehttp.WithInsecure(),
		)
		if err == nil {
			tp = sdktrace.NewTracerProvider(
				sdktrace.WithBatcher(exp),
				sdktrace.WithResource(res),
			)
		}
	}

	if tp == nil {
		w := os.Stdout
		exp, err := stdouttrace.New(stdouttrace.WithWriter(w), stdouttrace.WithPrettyPrint())
		if err != nil {
			return &noopTP{}
		}
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exp),
			sdktrace.WithResource(res),
		)
	}

	otel.SetTracerProvider(tp)
	return tp
}

type noopTP struct{}

func (n *noopTP) Shutdown(ctx context.Context) error { return nil }
