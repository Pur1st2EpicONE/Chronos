// Package impl contains the concrete implementation of the Service interface,
// coordinating the broker, cache, and storage layers.
package impl

import (
	"Chronos/internal/broker"
	"Chronos/internal/cache"
	"Chronos/internal/logger"
	"Chronos/internal/repository"
)

// Service implements the core business logic for notifications.
// It coordinates between the broker, cache, and storage layers to
// create, retrieve, update, and cancel notifications.
type Service struct {
	logger  logger.Logger      // logger for structured logging
	broker  broker.Broker      // broker layer for producing/consuming notifications
	cache   cache.Cache        // cache layer for fast status lookups
	storage repository.Storage // persistent storage for notifications
}

// NewService creates a new Service instance with the provided logger, broker, cache, and storage.
func NewService(logger logger.Logger, broker broker.Broker, cache cache.Cache, storage repository.Storage) *Service {
	return &Service{logger: logger, broker: broker, cache: cache, storage: storage}
}
