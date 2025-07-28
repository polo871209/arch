package database

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"

	"grpc-server/internal/config"
)

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
