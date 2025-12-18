package impl

import (
	"Chronos/internal/errs"
	"Chronos/internal/models"
	"context"
	"errors"
)

func (s *Service) CancelNotification(ctx context.Context, notificationID string) error {

	if cachedStatus, err := s.cache.GetStatus(ctx, notificationID); err == nil {
		if cachedStatus == models.StatusCanceled {
			return errs.ErrAlreadyCanceled
		}
		if cachedStatus == models.StatusSent || cachedStatus == models.StatusFailedToSendInTime {
			return errs.ErrCannotCancel
		}
	}

	if err := s.storage.SetStatus(ctx, notificationID, models.StatusCanceled); err != nil {

		switch {

		case errors.Is(err, errs.ErrNotificationNotFound):
			return errs.ErrNotificationNotFound

		case errors.Is(err, errs.ErrCannotCancel):
			currentStatus, err := s.storage.GetStatus(ctx, notificationID)
			if err != nil {
				return err
			}
			if currentStatus == models.StatusCanceled {
				if err := s.cache.SetStatus(ctx, notificationID, models.StatusCanceled); err != nil {
					s.logger.LogError("service — failed to set or update notification status in cache", err, "layer", "service.impl")
				}
				return errs.ErrAlreadyCanceled
			}
			return errs.ErrCannotCancel

		default:
			return err

		}

	}

	if err := s.cache.SetStatus(ctx, notificationID, models.StatusCanceled); err != nil {
		s.logger.LogError("service — failed to set or update notification status in cache", err, "layer", "service.impl")
	}

	return nil

}
