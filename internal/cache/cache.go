package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/valkey-io/valkey-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"grpc-server/internal/config"
)

// Common cache errors
var (
	ErrCacheMiss = errors.New("cache miss")
)

// Cache defines the interface for cache operations
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Close() error
}

// ValkeyCache implements Cache interface using Valkey
type ValkeyCache struct {
	client valkey.Client
	logger *slog.Logger
	tracer trace.Tracer
}

// NewValkeyCache creates a new Valkey cache instance
func NewValkeyCache(cfg *config.CacheConfig, logger *slog.Logger) (*ValkeyCache, error) {
	// Parse the URL to extract the host:port
	// For now, simple parsing - expecting format: valkey://host:port
	url := cfg.URL
	if url == "" {
		url = "valkey://localhost:6380"
	}

	// Extract host:port from valkey://host:port
	address := "localhost:6380" // default
	if len(url) > 9 && url[:9] == "valkey://" {
		address = url[9:] // Extract everything after "valkey://"
	}

	logger.Info("Creating Valkey client", "address", address)

	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create valkey client: %w", err)
	}

	return &ValkeyCache{
		client: client,
		logger: logger,
		tracer: otel.Tracer("valkey.cache"),
	}, nil
}

// Get retrieves a value from cache
func (c *ValkeyCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.logger.Debug("Attempting cache get", "key", key)

	ctx, span := c.tracer.Start(ctx, "cache.get",
		trace.WithAttributes(
			attribute.String("cache.key", key),
		),
	)
	defer span.End()

	result := c.client.Do(ctx, c.client.B().Get().Key(key).Build())
	if err := result.Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			c.logger.Debug("Cache miss", "key", key)
			span.SetAttributes(attribute.Bool("cache.hit", false))
			return nil, ErrCacheMiss
		}
		c.logger.Error("Cache get operation failed", "key", key, "error", err)
		span.RecordError(err)
		return nil, fmt.Errorf("cache get failed: %w", err)
	}

	data, err := result.AsBytes()
	if err != nil {
		c.logger.Error("Failed to convert cache result to bytes", "key", key, "error", err)
		span.RecordError(err)
		return nil, fmt.Errorf("failed to convert result: %w", err)
	}

	span.SetAttributes(
		attribute.Bool("cache.hit", true),
		attribute.Int("cache.value_size", len(data)),
	)
	c.logger.Debug("Cache hit successful", "key", key, "value_size", len(data))
	return data, nil
}

// Set stores a value in cache with expiration
func (c *ValkeyCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	c.logger.Debug("Attempting cache set", "key", key, "expiration", expiration)

	ctx, span := c.tracer.Start(ctx, "cache.set",
		trace.WithAttributes(
			attribute.String("cache.key", key),
			attribute.String("cache.expiration", expiration.String()),
		),
	)
	defer span.End()

	var data []byte
	var err error

	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		data, err = json.Marshal(value)
		if err != nil {
			c.logger.Error("Failed to marshal value for cache", "key", key, "error", err)
			span.RecordError(err)
			return fmt.Errorf("failed to marshal value: %w", err)
		}
	}

	span.SetAttributes(attribute.Int("cache.value_size", len(data)))

	var cmd valkey.Completed
	if expiration > 0 {
		cmd = c.client.B().Set().Key(key).Value(string(data)).ExSeconds(int64(expiration.Seconds())).Build()
	} else {
		cmd = c.client.B().Set().Key(key).Value(string(data)).Build()
	}

	result := c.client.Do(ctx, cmd)
	if err := result.Error(); err != nil {
		c.logger.Error("Cache set operation failed", "key", key, "error", err, "value_size", len(data))
		span.RecordError(err)
		return fmt.Errorf("cache set failed: %w", err)
	}

	c.logger.Debug("Cache set successful", "key", key, "expiration", expiration, "value_size", len(data))
	return nil
}

// Delete removes a value from cache
func (c *ValkeyCache) Delete(ctx context.Context, key string) error {
	c.logger.Debug("Attempting cache delete", "key", key)

	ctx, span := c.tracer.Start(ctx, "cache.delete",
		trace.WithAttributes(
			attribute.String("cache.key", key),
		),
	)
	defer span.End()

	result := c.client.Do(ctx, c.client.B().Del().Key(key).Build())
	if err := result.Error(); err != nil {
		c.logger.Error("Cache delete operation failed", "key", key, "error", err)
		span.RecordError(err)
		return fmt.Errorf("cache delete failed: %w", err)
	}

	// Check how many keys were deleted
	deletedCount, err := result.AsInt64()
	if err != nil {
		c.logger.Error("Failed to get delete result", "key", key, "error", err)
		span.RecordError(err)
		return fmt.Errorf("failed to get delete result: %w", err)
	}

	span.SetAttributes(attribute.Int64("cache.deleted_count", deletedCount))
	c.logger.Debug("Cache delete completed", "key", key, "deleted_count", deletedCount)
	return nil
}

// Close closes the cache connection
func (c *ValkeyCache) Close() error {
	c.logger.Info("Closing Valkey cache connection")
	c.client.Close()
	c.logger.Info("Valkey cache connection closed")
	return nil
}
