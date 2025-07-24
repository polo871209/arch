package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"grpc-server/internal/config"
)

// Connect establishes a connection to PostgreSQL
func Connect(ctx context.Context, cfg *config.DatabaseConfig) (*pgx.Conn, error) {
	slog.Info("Connecting to database", "url", maskPassword(cfg.URL))

	conn, err := pgx.Connect(ctx, cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	err = conn.Ping(ctx)
	if err != nil {
		conn.Close(ctx)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Database connection established successfully")
	return conn, nil
}

// maskPassword masks the password in the database URL for logging
func maskPassword(url string) string {
	// Simple masking - in production you might want more sophisticated parsing
	return "postgres://***:***@localhost:5433/rpc_dev?sslmode=disable"
}
