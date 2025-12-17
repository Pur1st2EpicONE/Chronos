package broker

import (
	rabbitmq "Chronos/internal/broker/rabbitMQ"
	"Chronos/internal/cache"
	"Chronos/internal/config"
	"Chronos/internal/logger"
	"Chronos/internal/models"
	"Chronos/internal/notifier"
	"Chronos/internal/repository"
	"context"
)

type Broker interface {
	Consume(ctx context.Context) error
	Produce(ctx context.Context, notification models.Notification) error
}

func NewBroker(logger logger.Logger, config config.Broker, cache cache.Cache, storage repository.Storage, notifier notifier.Notifier) (Broker, error) {
	return rabbitmq.NewBroker(logger, config, cache, storage, notifier)
}
