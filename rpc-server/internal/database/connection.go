package database

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"grpc-server/internal/config"
)

type spanContextKey struct{}

func Connect(ctx context.Context, cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	slog.Info("Connecting to database with connection pool", "url", maskPassword(cfg.URL))

	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool settings
	poolConfig.MaxConns = int32(cfg.MaxConns)
	poolConfig.MinConns = int32(cfg.MinConns)
	poolConfig.MaxConnLifetime = time.Duration(cfg.MaxLifetime) * time.Second
	poolConfig.MaxConnIdleTime = time.Duration(cfg.MaxIdleTime) * time.Second

	// Add OpenTelemetry tracing
	poolConfig.ConnConfig.Tracer = &pgxTracer{
		tracer: otel.Tracer("rpc-server.rpc/database"),
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Database connection pool established successfully",
		"max_conns", poolConfig.MaxConns,
		"min_conns", poolConfig.MinConns)
	return pool, nil
}

// pgxTracer implements pgx tracing interface
type pgxTracer struct {
	tracer trace.Tracer
}

func (t *pgxTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx, span := t.tracer.Start(ctx, "db.query",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "query"),
			attribute.String("db.statement", data.SQL),
		),
	)
	return context.WithValue(ctx, spanContextKey{}, span)
}

func (t *pgxTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	span, ok := ctx.Value(spanContextKey{}).(trace.Span)
	if !ok {
		return
	}
	defer span.End()

	if data.Err != nil {
		span.SetStatus(codes.Error, data.Err.Error())
		span.RecordError(data.Err)
		return
	}

	span.SetStatus(codes.Ok, "query successful")
	if rowsAffected := data.CommandTag.RowsAffected(); rowsAffected >= 0 {
		span.SetAttributes(attribute.Int64("db.rows_affected", rowsAffected))
	}
}

func maskPassword(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "<invalid-url>"
	}

	if parsedURL.User != nil {
		username := parsedURL.User.Username()
		parsedURL.User = url.UserPassword(username, "***")
	}

	masked := parsedURL.String()
	if strings.Contains(masked, "password=") {
		if password := parsedURL.Query().Get("password"); password != "" {
			masked = strings.ReplaceAll(masked, "password="+password, "password=***")
		}
	}

	return masked
}
