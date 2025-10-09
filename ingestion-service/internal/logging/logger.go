package logging

import (
	"os"

	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/config"
	"github.com/rs/zerolog"
)

func New(cfg *config.Config) zerolog.Logger {
	l := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return l
}
