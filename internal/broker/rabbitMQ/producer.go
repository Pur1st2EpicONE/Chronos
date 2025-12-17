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
			return err
		}

		ch, err := b.client.GetChannel()
		if err != nil {
			return err
		}
		defer ch.Close()

		body, err := json.Marshal(notification)
		if err != nil {
			return fmt.Errorf("failed to marshal notification to json: %w", err)
		}

		pub := amqp.Publishing{ContentType: "application/json", Body: body}

		if err := ch.PublishWithContext(ctx, mainExchange, notification.ID, false, false, pub); err != nil {
			return err
		}

		return nil

	})

}
