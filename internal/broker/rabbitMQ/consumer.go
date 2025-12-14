package rabbitmq

import (
	"Chronos/internal/models"
	"context"
	"encoding/json"
	"errors"

	"github.com/rabbitmq/amqp091-go"
)

func (b *Broker) Consume(ctx context.Context) error {

	go b.sysmon(ctx)

	if err := b.Consumer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	b.logger.LogInfo("consumer — stopped", "layer", "broker.rabbitMQ")
	return nil
}

func (b *Broker) handler(ctx context.Context, msg amqp091.Delivery) error {

	var notification models.Notification

	if err := json.Unmarshal(msg.Body, &notification); err != nil {
		return err
	}

	status, err := b.storage.GetStatus(ctx, notification.ID)
	if err != nil {
		return err
	}

	if status == models.StatusPending {
		status = models.StatusSent
	} else {
		status = models.StatusFailed
	}

	if err := b.notifier.Notify(notification); err != nil {
		return err
	}

	if err := b.storage.SetStatus(ctx, notification.ID, status); err != nil {
		return err
	}

	return nil

}
