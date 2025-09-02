package logging

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// TraceContextHandler wraps a slog.Handler and injects trace_id from the context into every record.
// Use this to ensure all logs include trace_id when context carries an OpenTelemetry span.
// Wrap your base handler with NewTraceContextHandler in main when creating the logger.

type TraceContextHandler struct {
	h slog.Handler
}

func NewTraceContextHandler(h slog.Handler) slog.Handler {
	return &TraceContextHandler{h: h}
}

func (t *TraceContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return t.h.Enabled(ctx, level)
}

func (t *TraceContextHandler) Handle(ctx context.Context, r slog.Record) error {
	sc := trace.SpanContextFromContext(ctx)
	if sc.IsValid() {
		r.AddAttrs(slog.String(TraceID, sc.TraceID().String()))
	}
	return t.h.Handle(ctx, r)
}

func (t *TraceContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceContextHandler{h: t.h.WithAttrs(attrs)}
}

func (t *TraceContextHandler) WithGroup(name string) slog.Handler {
	return &TraceContextHandler{h: t.h.WithGroup(name)}
}
