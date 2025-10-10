package logging

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func New(level string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	
	var logLevel zerolog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn", "warning":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	default:
		logLevel = zerolog.InfoLevel
	}
	
	zerolog.SetGlobalLevel(logLevel)
	
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", "api-gateway").
		Logger()
	
	return logger
}

func RequestLogger(logger zerolog.Logger) zerolog.Logger {
	return logger.With().
		Str("component", "http").
		Logger()
}
