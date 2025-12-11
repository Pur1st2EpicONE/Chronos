package rabbitmq

import (
	"Chronos/internal/models"
	"context"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

func (b *Broker) Consume(ctx context.Context) error {
	return b.Consumer.Start(ctx)
}

func (b *Broker) handler(ctx context.Context, msg amqp091.Delivery) error {

	var notification models.Notification

	if err := json.Unmarshal(msg.Body, &notification); err != nil {
		return msg.Nack(false, false)
	}

	status, err := b.storage.GetStatus(ctx, notification.ID)
	if err != nil {
		return msg.Nack(false, false)
	}

	if status == models.StatusPending {
		fmt.Println(notification)
	}

	return nil

}
