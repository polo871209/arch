package cache

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TracedCache wraps a Cache implementation with OpenTelemetry tracing
type TracedCache struct {
	cache  Cache
	tracer trace.Tracer
}

// NewTracedCache creates a new traced cache wrapper
func NewTracedCache(cache Cache, serviceName string) *TracedCache {
	return &TracedCache{
		cache:  cache,
		tracer: otel.Tracer(fmt.Sprintf("%s/cache", serviceName)),
	}
}

func (tc *TracedCache) Get(ctx context.Context, key string) ([]byte, error) {
	ctx, span := tc.tracer.Start(ctx, "cache.get",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("cache.operation", "get"),
			attribute.String("cache.key", key),
		),
	)
	defer span.End()

	data, err := tc.cache.Get(ctx, key)
	if err != nil {
		span.SetAttributes(attribute.Bool("cache.hit", false))
		if err == ErrCacheMiss {
			span.SetStatus(codes.Ok, "cache miss")
		} else {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		}
		return nil, err
	}

	span.SetAttributes(
		attribute.Bool("cache.hit", true),
		attribute.Int("cache.value_size", len(data)),
	)
	span.SetStatus(codes.Ok, "cache hit")
	return data, nil
}

func (tc *TracedCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	ctx, span := tc.tracer.Start(ctx, "cache.set",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("cache.operation", "set"),
			attribute.String("cache.key", key),
			attribute.String("cache.expiration", expiration.String()),
		),
	)
	defer span.End()

	if err := tc.cache.Set(ctx, key, value, expiration); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	span.SetStatus(codes.Ok, "cache set successful")
	return nil
}

func (tc *TracedCache) Delete(ctx context.Context, key string) error {
	ctx, span := tc.tracer.Start(ctx, "cache.delete",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("cache.operation", "delete"),
			attribute.String("cache.key", key),
		),
	)
	defer span.End()

	if err := tc.cache.Delete(ctx, key); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return err
	}

	span.SetStatus(codes.Ok, "cache delete successful")
	return nil
}

func (tc *TracedCache) Close() error {
	return tc.cache.Close()
}

// DeleteUntraced performs a delete operation without creating a span
// Useful for bulk operations where individual spans would create too much noise
func (tc *TracedCache) DeleteUntraced(ctx context.Context, key string) error {
	return tc.cache.Delete(ctx, key)
}
