package observability

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
)

type Logger = zerolog.Logger

func InitLogger(cfg config.Config) Logger {
	level := zerolog.InfoLevel
	zerolog.SetGlobalLevel(level)
	l := log.Output(zerolog.NewConsoleWriter())
	return l
}
