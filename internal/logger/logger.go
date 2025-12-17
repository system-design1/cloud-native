package logger

import (
	"os"
	"strings"

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
	// Output to stdout with JSON format (always JSON, never console format)
	// zerolog.New() always outputs JSON format by default
	globalLogger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger()

	// Set log level from environment variable
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info" // default level
	}
	SetLevel(logLevel)

	// Log the initialization (this will only show if level allows it)
	globalLogger.Info().
		Str("log_level", logLevel).
		Msg("Logger initialized with structured JSON logging")
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
// Supported levels: trace, debug, info, warn, error, fatal, panic, disabled
func SetLevel(level string) {
	level = strings.ToLower(strings.TrimSpace(level))
	switch level {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn", "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "disabled", "none", "off":
		zerolog.SetGlobalLevel(zerolog.Disabled)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
