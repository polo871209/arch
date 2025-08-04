package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"grpc-server/internal/config"

	"github.com/valkey-io/valkey-go"
)

// Common cache errors
var (
	ErrCacheMiss = errors.New("cache miss")
)

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
}

func NewValkeyCache(cfg *config.CacheConfig, logger *slog.Logger) (*ValkeyCache, error) {
	const prefix = "valkey://"
	address, ok := strings.CutPrefix(cfg.URL, prefix)
	if !ok || address == "" {
		return nil, errors.New("invalid cache URL: must start with valkey:// and include host:port")
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
	}, nil
}

func (c *ValkeyCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.logger.Debug("Attempting cache get", "key", key)

	result := c.client.Do(ctx, c.client.B().Get().Key(key).Build())
	if err := result.Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			c.logger.Debug("Cache miss", "key", key)
			return nil, ErrCacheMiss
		}
		c.logger.Error("Cache get operation failed", "key", key, "error", err)
		return nil, fmt.Errorf("cache get failed: %w", err)
	}

	data, err := result.AsBytes()
	if err != nil {
		c.logger.Error("Failed to convert cache result to bytes", "key", key, "error", err)
		return nil, fmt.Errorf("failed to convert result: %w", err)
	}

	c.logger.Debug("Cache hit successful", "key", key, "value_size", len(data))
	return data, nil
}

func (c *ValkeyCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	c.logger.Debug("Attempting cache set", "key", key, "expiration", expiration)

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
			return fmt.Errorf("failed to marshal value: %w", err)
		}
	}

	if expiration <= 0 {
		expiration = time.Hour
		c.logger.Warn("No expiration provided, using default", "key", key, "default_expiration", expiration)
	}

	result := c.client.Do(ctx, c.client.B().Set().Key(key).Value(string(data)).ExSeconds(int64(expiration.Seconds())).Build())
	if err := result.Error(); err != nil {
		c.logger.Error("Cache set operation failed", "key", key, "error", err, "value_size", len(data))
		return fmt.Errorf("cache set failed: %w", err)
	}

	c.logger.Debug("Cache set successful", "key", key, "expiration", expiration, "value_size", len(data))
	return nil
}

func (c *ValkeyCache) Delete(ctx context.Context, key string) error {
	c.logger.Debug("Attempting cache delete", "key", key)

	result := c.client.Do(ctx, c.client.B().Del().Key(key).Build())
	if err := result.Error(); err != nil {
		c.logger.Error("Cache delete operation failed", "key", key, "error", err)
		return fmt.Errorf("cache delete failed: %w", err)
	}

	// Check how many keys were deleted
	deletedCount, err := result.AsInt64()
	if err != nil {
		c.logger.Error("Failed to get delete result", "key", key, "error", err)
		return fmt.Errorf("failed to get delete result: %w", err)
	}

	c.logger.Debug("Cache delete completed", "key", key, "deleted_count", deletedCount)
	return nil
}

func (c *ValkeyCache) Close() error {
	c.client.Close()
	c.logger.Info("Valkey cache connection closed")
	return nil
}
