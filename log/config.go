package log

import "os"

// LoggerType defines the type of logger to use
type LoggerType string

const (
	LoggerTypeLogrus  LoggerType = "logrus"
	LoggerTypeZerolog LoggerType = "zerolog"
)

var (
	// currentLoggerType stores the current logger type
	currentLoggerType LoggerType
)

func init() {
	// Read from environment variable, default to zerolog
	envType := os.Getenv("LOG_LIBRARY")
	switch envType {
	case "logrus":
		currentLoggerType = LoggerTypeLogrus
	case "zerolog":
		currentLoggerType = LoggerTypeZerolog
	default:
		// Default to zerolog
		currentLoggerType = LoggerTypeZerolog
	}
}

// GetLoggerType returns the current logger type
func GetLoggerType() LoggerType {
	return currentLoggerType
}

// SetLoggerType sets the logger type (must be called before logger initialization)
// Note: This should be called before any logging operations
func SetLoggerType(t LoggerType) {
	currentLoggerType = t
}
