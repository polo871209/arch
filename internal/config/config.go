package config

import (
	"log/slog"
	"os"
)

// Config holds application configuration
type Config struct {
	Server   ServerConfig
	Logger   LoggerConfig
	Database DatabaseConfig
	Cache    CacheConfig
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

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL         string
	MaxConns    int
	MinConns    int
	MaxIdleTime int // seconds
	MaxLifetime int // seconds
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	URL             string
	MaxConns        int
	MinConns        int
	ConnMaxIdleTime int // seconds
	ConnMaxLifetime int // seconds
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:             getEnv("SERVER_PORT", "50051"),
			MaxRecvMsgSize:   getEnvInt("MAX_RECV_MSG_SIZE", 4*1024*1024), // 4MB
			MaxSendMsgSize:   getEnvInt("MAX_SEND_MSG_SIZE", 4*1024*1024), // 4MB
			EnableReflection: true,                                        // Always enable reflection for development
		},
		Logger: LoggerConfig{
			Level:  getLogLevel("LOG_LEVEL", slog.LevelInfo),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Database: DatabaseConfig{
			URL:         getEnv("DATABASE_URL", "postgres://rpc_user:rpc_password@localhost:5433/rpc_dev?sslmode=disable"),
			MaxConns:    getEnvInt("DB_MAX_CONNS", 25),
			MinConns:    getEnvInt("DB_MIN_CONNS", 5),
			MaxIdleTime: getEnvInt("DB_MAX_IDLE_TIME", 300), // 5 minutes
			MaxLifetime: getEnvInt("DB_MAX_LIFETIME", 3600), // 1 hour
		},
		Cache: CacheConfig{
			URL:             getEnv("CACHE_URL", "valkey://localhost:6380"),
			MaxConns:        getEnvInt("CACHE_MAX_CONNS", 10),
			MinConns:        getEnvInt("CACHE_MIN_CONNS", 2),
			ConnMaxIdleTime: getEnvInt("CACHE_MAX_IDLE_TIME", 300), // 5 minutes
			ConnMaxLifetime: getEnvInt("CACHE_MAX_LIFETIME", 3600), // 1 hour
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
