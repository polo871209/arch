package config

import (
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Logger   LoggerConfig
	Database DatabaseConfig
	Cache    CacheConfig
}

type ServerConfig struct {
	Port             string
	MaxRecvMsgSize   int
	MaxSendMsgSize   int
	EnableReflection bool
}

type LoggerConfig struct {
	Level  slog.Level
	Format string // "json" or "text"
}

type DatabaseConfig struct {
	URL         string
	MaxConns    int
	MinConns    int
	MaxIdleTime int // seconds
	MaxLifetime int // seconds
}

type CacheConfig struct {
	URL             string
	MaxConns        int
	MinConns        int
	ConnMaxIdleTime int // seconds
	ConnMaxLifetime int // seconds
}

func Load() *Config {
	slog.Debug("Loading application configuration")

	config := &Config{
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
			URL:         getEnv("DATABASE_URL", ""),
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

	slog.Info("Configuration loaded successfully",
		"server_port", config.Server.Port,
		"log_level", config.Logger.Level.String(),
		"log_format", config.Logger.Format,
	)

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	envVarStr, found := os.LookupEnv(key)
	if !found || envVarStr == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(envVarStr)
	if err != nil {
		slog.Error("Invalid integer value for environment variable", "key", key, "value", envVarStr, "error", err)
		return defaultValue
	}
	return val
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
