package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/valkey-io/valkey-go"
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
	}, nil
}

// Get retrieves a value from cache
func (c *ValkeyCache) Get(ctx context.Context, key string) ([]byte, error) {
	result := c.client.Do(ctx, c.client.B().Get().Key(key).Build())
	if err := result.Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			return nil, ErrCacheMiss
		}
		c.logger.Error("cache get failed", "key", key, "error", err)
		return nil, fmt.Errorf("cache get failed: %w", err)
	}

	data, err := result.AsBytes()
	if err != nil {
		c.logger.Error("failed to convert cache result to bytes", "key", key, "error", err)
		return nil, fmt.Errorf("failed to convert result: %w", err)
	}

	c.logger.Debug("cache hit", "key", key)
	return data, nil
}

// Set stores a value in cache with expiration
func (c *ValkeyCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
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
			return fmt.Errorf("failed to marshal value: %w", err)
		}
	}

	var cmd valkey.Completed
	if expiration > 0 {
		cmd = c.client.B().Set().Key(key).Value(string(data)).ExSeconds(int64(expiration.Seconds())).Build()
	} else {
		cmd = c.client.B().Set().Key(key).Value(string(data)).Build()
	}

	result := c.client.Do(ctx, cmd)
	if err := result.Error(); err != nil {
		c.logger.Error("cache set failed", "key", key, "error", err)
		return fmt.Errorf("cache set failed: %w", err)
	}

	c.logger.Debug("cache set", "key", key, "expiration", expiration)
	return nil
}

// Delete removes a value from cache
func (c *ValkeyCache) Delete(ctx context.Context, key string) error {
	result := c.client.Do(ctx, c.client.B().Del().Key(key).Build())
	if err := result.Error(); err != nil {
		c.logger.Error("cache delete failed", "key", key, "error", err)
		return fmt.Errorf("cache delete failed: %w", err)
	}

	c.logger.Debug("cache delete", "key", key)
	return nil
}

// Close closes the cache connection
func (c *ValkeyCache) Close() error {
	c.client.Close()
	return nil
}
