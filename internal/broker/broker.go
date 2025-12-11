package broker

import (
	rabbitmq "Chronos/internal/broker/rabbitMQ"
	"Chronos/internal/config"
	"Chronos/internal/models"
	"context"
)

type Broker interface {
	Produce(ctx context.Context, notification models.Notification) error
}

func NewBroker(config config.Broker) (Broker, error) {
	return rabbitmq.NewBroker(config)
}
