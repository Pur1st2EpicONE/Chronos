// Package broker provides abstractions for message brokers used in the application.
// It defines the Broker interface and exposes a constructor for concrete implementations.
package broker

import (
	rabbitmq "Chronos/internal/broker/rabbitMQ"
	"Chronos/internal/cache"
	"Chronos/internal/config"
	"Chronos/internal/logger"
	"Chronos/internal/models"
	"Chronos/internal/notifier"
	"Chronos/internal/repository"
)

// Broker defines the interface for a message broker used by the application.
// It supports consuming messages, producing notifications, and graceful shutdown.
type Broker interface {
	Consume() error                                 // Consume starts processing messages from the broker.
	Produce(notification models.Notification) error // Produce sends a notification message to the broker.
	Shutdown()                                      // Shutdown gracefully stops the broker and releases resources.
}

// NewBroker creates a new Broker instance. Currently, it returns a RabbitMQ-based broker.
// It initializes the broker with the provided logger, configuration, cache, storage, and notifier.
func NewBroker(logger logger.Logger, config config.Broker, cache cache.Cache, storage repository.Storage, notifier notifier.Notifier) (Broker, error) {
	return rabbitmq.NewBroker(logger, config, cache, storage, notifier)
}
