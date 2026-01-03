package impl

import (
	"Chronos/internal/errs"
	"Chronos/internal/models"
	"context"
	"time"

	"github.com/wb-go/wbf/helpers"
)

const brokerRecoveryWindow = time.Hour
const localDateTime = "2006-01-02 15:04:05"

// CreateNotification validates, initializes, stores, and enqueues a notification.
// If the broker fails and the notification is scheduled soon (within brokerRecoveryWindow),
// it will be removed from storage and an ErrUrgentDeliveryFailed is returned.
func (s *Service) CreateNotification(ctx context.Context, notification models.Notification) (string, error) {

	if err := validateCreate(&notification); err != nil {
		return "", err
	}

	initialize(&notification)

	if err := s.storage.CreateNotification(ctx, notification); err != nil {
		s.logger.LogError("service — failed to create notification", err, "layer", "service.impl")
		return "", err
	}

	if err := s.broker.Produce(notification); err != nil {

		s.logger.LogError("service — failed to produce notification", err, "layer", "service.impl")

		if time.Until(notification.SendAt) < brokerRecoveryWindow {

			err = s.storage.DeleteNotification(ctx, notification.ID)
			if err != nil {
				s.logger.LogError("service — failed to delete notification from db", err, "layer", "service.impl")
			}
			return "", errs.ErrUrgentDeliveryFailed

		}

	}

	return notification.ID, nil

}

// initialize sets the notification ID, status, updated timestamp, and local send time.
func initialize(notification *models.Notification) {
	notification.UpdatedAt = time.Now().UTC()
	notification.SendAtLocal = notification.SendAt.Local().Format(localDateTime)
	notification.ID = helpers.CreateUUID()
	notification.Status = models.StatusPending
}
