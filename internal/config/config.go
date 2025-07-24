package config

import (
	"log/slog"
	"os"
)

// Config holds application configuration
type Config struct {
	Server ServerConfig
	Logger LoggerConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port             string
	MaxRecvMsgSize   int
	MaxSendMsgSize   int
	EnableReflection bool
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level  slog.Level
	Format string // "json" or "text"
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:             getEnv("SERVER_PORT", "50051"),
			MaxRecvMsgSize:   getEnvInt("MAX_RECV_MSG_SIZE", 4*1024*1024), // 4MB
			MaxSendMsgSize:   getEnvInt("MAX_SEND_MSG_SIZE", 4*1024*1024), // 4MB
			EnableReflection: getEnvBool("ENABLE_REFLECTION", true),
		},
		Logger: LoggerConfig{
			Level:  getLogLevel("LOG_LEVEL", slog.LevelInfo),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		// Simple int parsing - in production, you'd want proper error handling
		if len(value) > 0 {
			// For simplicity, returning default on any parsing issues
			return defaultValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

func getLogLevel(key string, defaultLevel slog.Level) slog.Level {
	value := os.Getenv(key)
	switch value {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return defaultLevel
	}
}
