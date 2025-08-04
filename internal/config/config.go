package config

import (
	"fmt"
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
			Port:             requireEnv("GRPC_PORT"),
			MaxRecvMsgSize:   requireEnvInt("MAX_RECV_MSG_SIZE"),
			MaxSendMsgSize:   requireEnvInt("MAX_SEND_MSG_SIZE"),
			EnableReflection: requireEnvBool("ENABLE_REFLECTION"),
		},
		Logger: LoggerConfig{
			Level:  requireLogLevel("LOG_LEVEL"),
			Format: requireEnv("LOG_FORMAT"),
		},
		Database: DatabaseConfig{
			URL:         requireEnv("DATABASE_URL"),
			MaxConns:    requireEnvInt("DB_MAX_CONNS"),
			MinConns:    requireEnvInt("DB_MIN_CONNS"),
			MaxIdleTime: requireEnvInt("DB_MAX_IDLE_TIME"),
			MaxLifetime: requireEnvInt("DB_MAX_LIFETIME"),
		},
		Cache: CacheConfig{
			URL:             requireEnv("CACHE_URL"),
			MaxConns:        requireEnvInt("CACHE_MAX_CONNS"),
			MinConns:        requireEnvInt("CACHE_MIN_CONNS"),
			ConnMaxIdleTime: requireEnvInt("CACHE_MAX_IDLE_TIME"),
			ConnMaxLifetime: requireEnvInt("CACHE_MAX_LIFETIME"),
		},
	}

	slog.Info("Configuration loaded successfully",
		"server_port", config.Server.Port,
		"log_level", config.Logger.Level.String(),
		"log_format", config.Logger.Format,
	)

	return config
}

func requireEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("Environment variable %s is required but not set", key))
	}
	return value
}

func requireEnvInt(key string) int {
	envVarStr := requireEnv(key)
	val, err := strconv.Atoi(envVarStr)
	if err != nil {
		panic(fmt.Sprintf("Environment variable %s must be a valid integer, got: %s", key, envVarStr))
	}
	return val
}

func requireEnvBool(key string) bool {
	envVarStr := requireEnv(key)
	val, err := strconv.ParseBool(envVarStr)
	if err != nil {
		panic(fmt.Sprintf("Environment variable %s must be a valid boolean, got: %s", key, envVarStr))
	}
	return val
}

func requireLogLevel(key string) slog.Level {
	value := requireEnv(key)
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
		panic(fmt.Sprintf("Environment variable %s must be one of: DEBUG, INFO, WARN, ERROR, got: %s", key, value))
	}
}
