package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/melihgurlek/backend-path/pkg/metrics"
)

// RedisCache provides Redis-based caching functionality
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(redisURL string) (*RedisCache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info().Msg("Redis cache connected successfully")

	return &RedisCache{
		client: client,
	}, nil
}

// Get retrieves a value from cache
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	start := time.Now()
	defer func() {
		metrics.CacheOperations.WithLabelValues("get", "success").Inc()
		metrics.CacheOperationDuration.WithLabelValues("get").Observe(time.Since(start).Seconds())
	}()

	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			metrics.CacheOperations.WithLabelValues("get", "miss").Inc()
			return false, nil // Cache miss, not an error
		}
		metrics.CacheOperations.WithLabelValues("get", "error").Inc()
		return false, fmt.Errorf("failed to get from cache: %w", err)
	}

	// Unmarshal JSON value
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		metrics.CacheOperations.WithLabelValues("get", "error").Inc()
		return false, fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	metrics.CacheOperations.WithLabelValues("get", "hit").Inc()
	return true, nil
}

// Set stores a value in cache with TTL
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		metrics.CacheOperations.WithLabelValues("set", "success").Inc()
		metrics.CacheOperationDuration.WithLabelValues("set").Observe(time.Since(start).Seconds())
	}()

	// Marshal value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		metrics.CacheOperations.WithLabelValues("set", "error").Inc()
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		metrics.CacheOperations.WithLabelValues("set", "error").Inc()
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		metrics.CacheOperations.WithLabelValues("delete", "success").Inc()
		metrics.CacheOperationDuration.WithLabelValues("delete").Observe(time.Since(start).Seconds())
	}()

	if err := c.client.Del(ctx, key).Err(); err != nil {
		metrics.CacheOperations.WithLabelValues("delete", "error").Inc()
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

// DeletePattern removes all keys matching a pattern
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	start := time.Now()
	defer func() {
		metrics.CacheOperations.WithLabelValues("delete_pattern", "success").Inc()
		metrics.CacheOperationDuration.WithLabelValues("delete_pattern").Observe(time.Since(start).Seconds())
	}()

	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			metrics.CacheOperations.WithLabelValues("delete_pattern", "error").Inc()
			return fmt.Errorf("failed to delete key %s: %w", iter.Val(), err)
		}
	}

	if err := iter.Err(); err != nil {
		metrics.CacheOperations.WithLabelValues("delete_pattern", "error").Inc()
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	return nil
}

// Exists checks if a key exists in cache
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return result > 0, nil
}

// TTL gets the remaining TTL for a key
func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// GetStats returns cache statistics
func (c *RedisCache) GetStats(ctx context.Context) (*redis.PoolStats, error) {
	stats := c.client.PoolStats()
	return stats, nil
}
