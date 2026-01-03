// Package redis provides a Redis-based implementation of the Cache interface.
// It handles storing and retrieving notification statuses with retry and expiration logic.
package redis

import (
	"Chronos/internal/config"
	"Chronos/internal/logger"
	"Chronos/internal/models"
	"context"

	r "github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
)

// Cache implements the Cache interface using a Redis backend.
type Cache struct {
	client *r.Client     // underlying Redis client
	logger logger.Logger // structured logger
	config config.Cache  // cache configuration
}

// Connect establishes a connection to Redis using the provided logger and configuration.
// Returns a Cache instance or an error if the connection fails.
func Connect(logger logger.Logger, config config.Cache) (*Cache, error) {
	client, err := r.Connect(r.Options{
		Address:   config.Host + ":" + config.Port,
		Password:  config.Password,
		MaxMemory: config.MaxMemory,
		Policy:    config.Policy})
	if err != nil {
		return nil, err
	}
	return &Cache{client: client, logger: logger, config: config}, nil
}

// SetStatus sets the value for a given key in Redis with expiration and retry strategy.
func (c *Cache) SetStatus(ctx context.Context, key string, value any) error {
	return c.client.SetWithExpirationAndRetry(ctx, retry.Strategy{
		Attempts: c.config.RetryStrategy.Attempts,
		Delay:    c.config.RetryStrategy.Delay,
		Backoff:  c.config.RetryStrategy.Backoff},
		key, value, c.config.ExpirationTime)
}

// GetStatus retrieves the value for a given key from Redis and refreshes its expiration.
func (c *Cache) GetStatus(ctx context.Context, key string) (string, error) {
	if err := c.client.Expire(ctx, key, c.config.ExpirationTime); err != nil {
		return "", err
	}
	return c.client.Get(ctx, key)
}

// MarkLates marks a list of notifications as late in Redis with expiration and retry strategy.
func (c *Cache) MarkLates(ctx context.Context, lates []string) error {
	if len(lates) > 0 {
		for _, key := range lates {
			if err := c.client.SetWithExpirationAndRetry(ctx, retry.Strategy{
				Attempts: c.config.RetryStrategy.Attempts,
				Delay:    c.config.RetryStrategy.Delay,
				Backoff:  c.config.RetryStrategy.Backoff},
				key, models.StatusLate, c.config.ExpirationTime); err != nil {
				return err
			}
		}
	}
	return nil
}

// Close shuts down the Redis client and logs the outcome.
func (c *Cache) Close() {
	if err := c.client.Close(); err != nil {
		c.logger.LogError("redis — failed to close properly", err, "layer", "cache.redis")
	} else {
		c.logger.LogInfo("redis — cache closed", "layer", "cache.redis")
	}
}
