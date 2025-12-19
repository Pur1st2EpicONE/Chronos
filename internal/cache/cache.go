package cache

import (
	"Chronos/internal/cache/redis"
	"Chronos/internal/config"
	"Chronos/internal/logger"
	"context"
)

type Cache interface {
	SetStatus(ctx context.Context, key string, value any) error
	GetStatus(ctx context.Context, key string) (string, error)
	MarkLates(ctx context.Context, lates []string) error
	Close()
}

func Connect(logger logger.Logger, config config.Cache) (Cache, error) {
	return redis.Connect(logger, config)
}
