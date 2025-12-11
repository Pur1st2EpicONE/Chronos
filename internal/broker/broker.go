package broker

import (
	rabbitmq "Chronos/internal/broker/rabbitMQ"
	"Chronos/internal/config"
	"Chronos/internal/models"
	"Chronos/internal/repository"
	"context"
)

type Broker interface {
	Consume(ctx context.Context) error
	Produce(ctx context.Context, notification models.Notification) error
}

func NewBroker(config config.Broker, storage repository.Storage) (Broker, error) {
	return rabbitmq.NewBroker(config, storage)
}
