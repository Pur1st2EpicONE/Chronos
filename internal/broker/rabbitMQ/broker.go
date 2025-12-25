package rabbitmq

import (
	"Chronos/internal/cache"
	"Chronos/internal/config"
	"Chronos/internal/logger"
	"Chronos/internal/notifier"
	"Chronos/internal/repository"
	"context"
	"fmt"
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
	cache    cache.Cache
	storage  repository.Storage
	notifier notifier.Notifier
	client   *rabbitmq.RabbitClient
}

func NewBroker(logger logger.Logger, config config.Broker, cache cache.Cache, storage repository.Storage, notifier notifier.Notifier) (*Broker, error) {

	client, err := rabbitmq.NewClient(rabbitmq.ClientConfig{

		URL:            config.URL,
		ConnectionName: config.ConnectionName,
		ConnectTimeout: config.ConnectTimeout,

		ReconnectStrat: retry.Strategy{
			Attempts: config.Reconnect.Attempts,
			Delay:    config.Reconnect.Delay,
			Backoff:  config.Reconnect.Backoff},

		ConsumingStrat: retry.Strategy{
			Attempts: config.Consumer.Attempts,
			Delay:    config.Reconnect.Delay,
			Backoff:  config.Reconnect.Backoff}})

	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}

	err = client.DeclareExchange(mainExchange, exchangeKind, true, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	err = client.DeclareQueue(config.QueueName, mainExchange, config.QueueName, true, false, true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	producer := rabbitmq.NewPublisher(client, mainExchange, contentType)

	b := &Broker{
		logger:   logger,
		config:   config,
		Consumer: nil,
		producer: producer,
		cache:    cache,
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

func (b *Broker) Shutdown() {
	if err := b.client.Close(); err != nil {
		b.logger.LogError("broker — failed to shutdown gracefully", err, "layer", "broker.rabbitMQ")
	} else {
		b.logger.LogInfo("broker — shutdown complete", "layer", "broker.rabbitMQ")
	}
}

func (b *Broker) sysmon(ctx context.Context) {

	b.storage.Cleanup(ctx)
	b.recover(ctx)

	cleaner := time.NewTicker(b.config.CleanupInterval)
	healthcheck := time.NewTicker(b.config.HealthcheckInterval)

	defer cleaner.Stop()
	defer healthcheck.Stop()

	var wasUnhealthy bool

	for {

		select {
		case <-ctx.Done():
			return
		case <-cleaner.C:
			b.storage.Cleanup(ctx)
		case <-healthcheck.C:
		}

		healthy := b.client.Healthy()

		if !healthy {
			wasUnhealthy = true
			lates, err := b.storage.MarkLates(ctx)
			if err != nil {
				b.logger.LogError("broker — failed to mark late notifications in db", err, "layer", "broker.rabbitMQ")
				continue
			}
			if err := b.cache.MarkLates(ctx, lates); err != nil {
				b.logger.LogError("broker — failed to mark late notifications in cache", err, "layer", "broker.rabbitMQ")
			}
			continue
		}

		if wasUnhealthy {
			b.recover(ctx)
			wasUnhealthy = false
		}

	}

}

func (b *Broker) recover(ctx context.Context) {
	notifications, err := b.storage.Recover(ctx)
	if err != nil {
		b.logger.LogError("broker — failed to recover notifications from db", err, "layer", "broker.rabbitMQ")
		return
	}
	for _, notification := range notifications {
		if err := b.Produce(notification); err != nil {
			b.logger.LogError("broker — failed to produce notification", err, "notificationID", notification.ID, "layer", "broker.rabbitMQ")
		}
	}
	b.logger.Debug("broker — recovered", "layer", "broker.rabbitMQ")
}
