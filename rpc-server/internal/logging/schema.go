package logging

import (
	"context"
	"log/slog"
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

// Context-aware convenience methods
func (l *Logger) DebugCtx(ctx context.Context, msg string, args ...any) {
	l.Log(ctx, slog.LevelDebug, msg, args...)
}
func (l *Logger) InfoCtx(ctx context.Context, msg string, args ...any) {
	l.Log(ctx, slog.LevelInfo, msg, args...)
}
func (l *Logger) WarnCtx(ctx context.Context, msg string, args ...any) {
	l.Log(ctx, slog.LevelWarn, msg, args...)
}
func (l *Logger) ErrorCtx(ctx context.Context, msg string, args ...any) {
	l.Log(ctx, slog.LevelError, msg, args...)
}
