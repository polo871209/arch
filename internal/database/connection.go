package database

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"grpc-server/internal/config"
)

func Connect(ctx context.Context, cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	slog.Info("Connecting to database with connection pool", "url", maskPassword(cfg.URL))

	// Parse the connection URL to add pool configuration
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool settings for better concurrency
	poolConfig.MaxConns = 50       // Maximum number of connections (increased for 4+ workers)
	poolConfig.MinConns = 10       // Minimum number of connections to keep open
	poolConfig.MaxConnLifetime = 0 // Don't close connections due to age
	poolConfig.MaxConnIdleTime = 0 // Don't close connections due to idle time

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}

	// Test the connection
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Database connection pool established successfully",
		"max_conns", poolConfig.MaxConns,
		"min_conns", poolConfig.MinConns)
	return pool, nil
}

func maskPassword(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "<invalid-url>"
	}

	if parsedURL.User != nil {
		username := parsedURL.User.Username()
		parsedURL.User = url.UserPassword(username, "***") // mask password
	}

	// Hide query parameters too if needed
	masked := parsedURL.String()

	// Optionally: strip password from query if any (for edge cases)
	if strings.Contains(masked, "password=") {
		masked = strings.ReplaceAll(masked, "password="+parsedURL.Query().Get("password"), "password=***")
	}

	return masked
}
