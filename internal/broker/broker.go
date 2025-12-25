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

type Broker interface {
	Consume() error
	Produce(notification models.Notification) error
	Shutdown()
}

func NewBroker(logger logger.Logger, config config.Broker, cache cache.Cache, storage repository.Storage, notifier notifier.Notifier) (Broker, error) {
	return rabbitmq.NewBroker(logger, config, cache, storage, notifier)
}
