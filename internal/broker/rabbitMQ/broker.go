// Package rabbitmq provides a RabbitMQ-based implementation of the Broker interface.
// It manages message publishing, consumption, and system health monitoring.
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

// Broker is a RabbitMQ implementation of the Broker interface.
// It holds references to logger, config, consumer, producer, cache, storage, notifier, and the underlying RabbitClient.
type Broker struct {
	logger   logger.Logger          // structured logger for logging broker events
	config   config.Broker          // broker configuration
	Consumer *rabbitmq.Consumer     // RabbitMQ consumer instance
	producer *rabbitmq.Publisher    // RabbitMQ publisher instance
	cache    cache.Cache            // cache interface for temporary storage
	storage  repository.Storage     // storage interface for persistent storage
	notifier notifier.Notifier      // notifier for sending notifications
	client   *rabbitmq.RabbitClient // underlying RabbitMQ client
}

// NewBroker creates and initializes a new RabbitMQ Broker instance.
// It sets up the RabbitMQ client, exchange, queue, producer, and consumer.
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

// Shutdown gracefully closes the underlying RabbitMQ client and logs the outcome.
func (b *Broker) Shutdown() {
	if err := b.client.Close(); err != nil {
		b.logger.LogError("broker — failed to shutdown gracefully", err, "layer", "broker.rabbitMQ")
	} else {
		b.logger.LogInfo("broker — shutdown complete", "layer", "broker.rabbitMQ")
	}
}

// sysmon runs system monitoring tasks including cleanup, health checks, and recovery.
// It periodically checks the health of the broker client and triggers recovery if needed.
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

// recover retrieves pending notifications from storage and re-queues them for processing.
// It logs any errors encountered during recovery.
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
