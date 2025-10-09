package observability

import (
	"context"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
	"github.com/rs/zerolog"
)

type TracerProvider interface {
	Shutdown(ctx context.Context) error
}

type nopTP struct{}

func (n *nopTP) Shutdown(ctx context.Context) error { return nil }

func InitTracing(cfg config.Config, logger zerolog.Logger) TracerProvider {
	return &nopTP{}
}
