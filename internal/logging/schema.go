package logging

import "log/slog"

// Standard field names - use these consistently across all logging
const (
	UserID    = "user_id"
	UserEmail = "user_email"
	CacheKey  = "cache_key"
	Error     = "error"
)

// Logger wraps slog.Logger with consistent field names
type Logger struct {
	*slog.Logger
}

// New creates a new standardized logger
func New(logger *slog.Logger) *Logger {
	return &Logger{Logger: logger}
}
