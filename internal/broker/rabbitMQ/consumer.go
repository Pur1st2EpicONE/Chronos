package rabbitmq

import (
	"Chronos/internal/models"
	"context"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

func (b *Broker) Consume(ctx context.Context) error {
	if err := b.Consumer.Start(ctx); err != nil {
		return err
	}
	b.logger.LogInfo("consumer — stopped", "layer", "broker.rabbitMQ")
	return nil
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
		return b.send(ctx, notification)
	}

	return nil

}

func (b *Broker) send(ctx context.Context, notification models.Notification) error {
	fmt.Println(notification.Message)
	if err := b.storage.SetStatus(ctx, notification.ID, models.StatusSent); err != nil {
		return err
	}
	return nil
}
