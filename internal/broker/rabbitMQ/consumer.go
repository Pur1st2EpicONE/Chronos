package rabbitmq

import (
	"Chronos/internal/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/retry"
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
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	status, err := b.storage.GetStatus(ctx, notification.ID)
	if err != nil {
		return err
	}

	if status != models.StatusCanceled {

		if err := b.notifier.Notify(notification); err != nil {
			if status != models.StatusFailed {
				b.updateStatus(ctx, notification.ID, notification.SendAt, models.StatusFailed)
			}
			return err
		}

		b.updateStatus(ctx, notification.ID, notification.SendAt, status)

	}

	return nil

}

func (b *Broker) updateStatus(ctx context.Context, notificationID string, sendAt time.Time, status string) {

	if status == models.StatusPending {
		status = models.StatusSent
	}

	if status == models.StatusLate || time.Now().UTC().Sub(sendAt) > time.Minute {
		status = models.StatusFailedToSendInTime
	}

	if err := retry.DoContext(ctx, retry.Strategy{
		Attempts: b.config.Consumer.Attempts,
		Delay:    b.config.Consumer.Delay,
		Backoff:  b.config.Consumer.Backoff,
	}, func() error {
		return b.cache.SetStatus(ctx, notificationID, status)
	}); err != nil {
		b.logger.LogError("broker — failed to update notification status in cache",
			err, "notificationID", notificationID, "layer", "broker.rabbitMQ")
	}

	if err := retry.DoContext(ctx, retry.Strategy{
		Attempts: b.config.Consumer.Attempts,
		Delay:    b.config.Consumer.Delay,
		Backoff:  b.config.Consumer.Backoff,
	}, func() error { return b.storage.SetStatus(ctx, notificationID, status) }); err != nil {
		b.logger.LogError("broker — failed to update notification status in db",
			err, "notificationID", notificationID, "layer", "broker.rabbitMQ")
	}

}
