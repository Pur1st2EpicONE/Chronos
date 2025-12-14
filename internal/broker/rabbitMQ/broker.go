package rabbitmq

import (
	"Chronos/internal/config"
	"Chronos/internal/logger"
	"Chronos/internal/notifier"
	"Chronos/internal/repository"
	"context"
	"time"

	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	mainExchange = "mainExchange"
	contentType  = "application/json"
	exchangeKind = "direct"
)

type Broker struct {
	logger   logger.Logger
	config   config.Broker
	Consumer *rabbitmq.Consumer
	producer *rabbitmq.Publisher
	storage  repository.Storage
	notifier notifier.Notifier
	client   *rabbitmq.RabbitClient
}

func NewBroker(logger logger.Logger, config config.Broker, storage repository.Storage, notifier notifier.Notifier) (*Broker, error) {

	client, err := rabbitmq.NewClient(rabbitmq.ClientConfig{

		URL:            config.URL,
		ConnectionName: config.ConnectionName,
		ConnectTimeout: config.ConnectTimeout,
		Heartbeat:      config.Heartbeat,

		ReconnectStrat: retry.Strategy{
			Attempts: config.Reconnect.Attempts,
			Delay:    config.Reconnect.Delay,
			Backoff:  config.Reconnect.Backoff}})

	if err != nil {
		return nil, err
	}

	err = client.DeclareExchange(mainExchange, exchangeKind, true, false, false, nil)
	if err != nil {
		client.Close()
		return nil, err
	}

	err = client.DeclareQueue(config.QueueName, mainExchange, config.QueueName, true, false, true, nil)
	if err != nil {
		client.Close()
		return nil, err
	}

	producer := rabbitmq.NewPublisher(client, mainExchange, contentType)

	b := &Broker{
		logger:   logger,
		config:   config,
		Consumer: nil,
		producer: producer,
		storage:  storage,
		notifier: notifier,
		client:   client}

	b.Consumer = rabbitmq.NewConsumer(client, rabbitmq.ConsumerConfig{
		Queue:         config.QueueName,
		ConsumerTag:   config.Consumer.ConsumerTag,
		AutoAck:       config.Consumer.AutoAck,
		Ask:           rabbitmq.AskConfig{Multiple: false},
		Nack:          rabbitmq.NackConfig{Multiple: false, Requeue: true},
		Args:          amqp.Table{},
		Workers:       config.Consumer.Workers,
		PrefetchCount: config.Consumer.PrefetchCount,
	}, func(ctx context.Context, msg amqp.Delivery) error { return b.handler(ctx, msg) })

	return b, nil

}

func (b *Broker) sysmon(ctx context.Context) {

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var wasUnhealthy bool

	for {

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		healthy := b.client.Healthy()

		if !healthy {

			if err := b.storage.MarkLate(ctx); err != nil {
				continue
			}
			wasUnhealthy = true

		} else {

			if wasUnhealthy {
				notifications, err := b.storage.Recover(ctx)
				if err != nil {
					continue
				}
				for _, n := range notifications {
					if err := b.Produce(ctx, n); err != nil {
						//
					}
				}
				wasUnhealthy = false
			}

		}

	}

}
