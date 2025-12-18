package redis

import (
	"Chronos/internal/config"
	"Chronos/internal/logger"
	"Chronos/internal/models"
	"context"
	"time"

	r "github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
)

type Cache struct {
	client *r.Client
	logger logger.Logger
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
	return &Cache{client: client, logger: logger}, nil
}

func (c *Cache) SetStatus(ctx context.Context, key string, value any) error {
	return c.client.SetWithExpirationAndRetry(ctx, retry.Strategy{Attempts: 3, Delay: 2, Backoff: 2}, key, value, 2*time.Minute)
}

func (c *Cache) GetStatus(ctx context.Context, key string) (string, error) {
	if err := c.client.Expire(ctx, key, 2*time.Minute); err != nil {
		return "", err
	}
	return c.client.GetWithRetry(ctx, retry.Strategy{Attempts: 3, Delay: 2, Backoff: 2}, key)
}

func (c *Cache) MarkLates(ctx context.Context, lates []string) {
	if len(lates) > 0 {
		for _, key := range lates {
			if err := c.client.SetWithExpirationAndRetry(ctx, retry.Strategy{Attempts: 3, Delay: 2, Backoff: 2}, key, models.StatusLate, 2*time.Minute); err != nil {
				c.logger.LogError("redis — failed to set notification status", err, "layer", "cache.redis")
			}
		}
	}
}

func (c *Cache) Close() {
	if err := c.client.Close(); err != nil {
		c.logger.LogError("redis — failed to close properly", err, "layer", "cache.redis")
	} else {
		c.logger.LogInfo("redis — cache closed", "layer", "cache.redis")
	}
}
