package logger

import (
	"os"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var (
	globalLogger zerolog.Logger
)

// Init initializes the global logger with structured JSON logging
func Init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.TimestampFieldName = "timestamp"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"

	// Use JSON format for all logs (structured logging)
	globalLogger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger()
}

// Get returns the global logger instance with optional correlation ID
func Get(correlationID ...string) zerolog.Logger {
	logger := globalLogger
	if len(correlationID) > 0 && correlationID[0] != "" {
		logger = logger.With().Str("correlation_id", correlationID[0]).Logger()
	}
	return logger
}

// GenerateCorrelationID generates a new correlation ID
func GenerateCorrelationID() string {
	return uuid.New().String()
}

// SetLevel sets the logging level
func SetLevel(level string) {
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
