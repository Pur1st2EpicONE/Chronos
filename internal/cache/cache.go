// Package cache provides abstractions for caching layer used in the application.
// It defines the Cache interface and exposes a constructor for concrete implementations.
package cache

import (
	"Chronos/internal/cache/redis"
	"Chronos/internal/config"
	"Chronos/internal/logger"
	"context"
)

// Cache defines the interface for a caching layer used by the application.
// It supports storing and retrieving notification statuses, marking late notifications, and closing the cache connection.
type Cache interface {
	SetStatus(ctx context.Context, key string, value any) error // SetStatus sets the status value for the given key in the cache.
	GetStatus(ctx context.Context, key string) (string, error)  // GetStatus retrieves the status value for the given key from the cache.
	MarkLates(ctx context.Context, lates []string) error        // MarkLates marks a list of notifications as late in the cache.
	Close()                                                     // Close closes the cache connection and releases resources.
}

// Connect creates a new Cache instance (currently Redis) using the provided logger and configuration.
// It returns the initialized cache or an error if connection fails.
func Connect(logger logger.Logger, config config.Cache) (Cache, error) {
	return redis.Connect(logger, config)
}
