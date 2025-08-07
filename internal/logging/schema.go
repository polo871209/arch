package logging

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// Standard field names - use these consistently across all logging
const (
	UserID    = "user_id"
	UserEmail = "user_email"
	CacheKey  = "cache_key"
	Error     = "error"
	TraceID   = "trace_id"
)

// Logger wraps slog.Logger with consistent field names
type Logger struct {
	*slog.Logger
}

// New creates a new standardized logger
func New(logger *slog.Logger) *Logger {
	return &Logger{Logger: logger}
}

// WithTrace returns a logger enriched with trace_id from context if available
func WithTrace(ctx context.Context, logger *Logger) *Logger {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return logger
	}
	return &Logger{Logger: logger.With(
		TraceID, sc.TraceID().String(),
	)}
}
