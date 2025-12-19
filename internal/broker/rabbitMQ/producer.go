package rabbitmq

import (
	"Chronos/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/retry"
)

func (b *Broker) Produce(ctx context.Context, notification models.Notification) error {

	sendAt := max(time.Until(notification.SendAt), 0)

	return retry.DoContext(ctx, retry.Strategy{
		Attempts: b.config.Producer.Attempts,
		Delay:    b.config.Producer.Delay,
		Backoff:  b.config.Producer.Backoff}, func() error {

		queueArgs := amqp.Table{
			"x-message-ttl":             int64(sendAt.Milliseconds()),
			"x-dead-letter-exchange":    mainExchange,
			"x-dead-letter-routing-key": b.config.QueueName,
			"x-expires":                 int64(sendAt.Milliseconds() + b.config.Producer.MessageQueueTTL.Milliseconds()),
		}

		err := b.client.DeclareQueue(notification.ID, mainExchange, notification.ID, false, true, true, queueArgs)
		if err != nil {
			if amqpErr, ok := err.(*amqp.Error); ok && amqpErr.Code == amqp.PreconditionFailed { // exception 406
				b.logger.Debug("producer — recovered notification is already in queue, skipping",
					"notificationID", notification.ID, "layer", "broker.rabbitMQ")
				return nil
			}
			return fmt.Errorf("failed to declare queue: %w", err)
		}

		ch, err := b.client.GetChannel()
		if err != nil {
			return fmt.Errorf("failed to get channel: %w", err)
		}
		defer func() { _ = ch.Close() }()

		body, err := json.Marshal(notification)
		if err != nil {
			return fmt.Errorf("failed to marshal notification to json: %w", err)
		}

		pub := amqp.Publishing{ContentType: contentType, Body: body}

		if err := ch.PublishWithContext(ctx, mainExchange, notification.ID, false, false, pub); err != nil {
			return fmt.Errorf("failed to publish with context: %w", err)
		}

		return nil

	})

}
