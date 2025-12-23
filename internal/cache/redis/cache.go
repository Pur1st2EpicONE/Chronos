package redis

import (
	"Chronos/internal/config"
	"Chronos/internal/logger"
	"Chronos/internal/models"
	"context"

	r "github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
)

type Cache struct {
	client *r.Client
	logger logger.Logger
	config config.Cache
}

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

func (c *Cache) SetStatus(ctx context.Context, key string, value any) error {
	return c.client.SetWithExpirationAndRetry(ctx, retry.Strategy{
		Attempts: c.config.RetryStrategy.Attempts,
		Delay:    c.config.RetryStrategy.Delay,
		Backoff:  c.config.RetryStrategy.Backoff},
		key, value, c.config.ExpirationTime)
}

func (c *Cache) GetStatus(ctx context.Context, key string) (string, error) {
	if err := c.client.Expire(ctx, key, c.config.ExpirationTime); err != nil {
		return "", err
	}
	return c.client.Get(ctx, key)
}

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

func (c *Cache) Close() {
	if err := c.client.Close(); err != nil {
		c.logger.LogError("redis — failed to close properly", err, "layer", "cache.redis")
	} else {
		c.logger.LogInfo("redis — cache closed", "layer", "cache.redis")
	}
}
