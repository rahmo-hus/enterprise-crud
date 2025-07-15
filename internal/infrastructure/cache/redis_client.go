// Package cache provides Redis caching infrastructure for the application.
// It includes Redis client management and caching services for domain entities.
package cache

import (
	"context"
	"fmt"

	"enterprise-crud/internal/config"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis client with application-specific configuration
type RedisClient struct {
	client *redis.Client
	config *config.RedisConfig
}

// NewRedisClient creates a new Redis client instance with the provided configuration
func NewRedisClient(cfg *config.RedisConfig) (*RedisClient, error) {
	// Create Redis client options
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		client: rdb,
		config: cfg,
	}, nil
}

// GetClient returns the underlying Redis client for direct usage
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// GetConfig returns the Redis configuration
func (r *RedisClient) GetConfig() *config.RedisConfig {
	return r.config
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Ping tests the Redis connection
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
