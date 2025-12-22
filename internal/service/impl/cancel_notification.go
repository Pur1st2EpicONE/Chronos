package impl

import (
	"Chronos/internal/errs"
	"Chronos/internal/models"
	"context"
	"errors"
)

func (s *Service) CancelNotification(ctx context.Context, notificationID string) error {

	if cachedStatus, err := s.cache.GetStatus(ctx, notificationID); err == nil {
		switch cachedStatus {
		case models.StatusCanceled:
			return errs.ErrAlreadyCanceled
		case models.StatusSent, models.StatusFailedToSendInTime:
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
				s.logger.LogError("service — failed to get notification status from DB", err, "layer", "service.impl")
				return err
			}
			if err := s.cache.SetStatus(ctx, notificationID, currentStatus); err != nil {
				s.logger.LogError("service — failed to set notification status in cache", err, "layer", "service.impl")
			}
			if currentStatus == models.StatusCanceled {
				return errs.ErrAlreadyCanceled
			}
			return errs.ErrCannotCancel

		default:
			s.logger.LogError("service — failed to set notification status in DB", err, "layer", "service.impl")
			return err

		}

	}

	if err := s.cache.SetStatus(ctx, notificationID, models.StatusCanceled); err != nil {
		s.logger.LogError("service — failed to set notification status in cache", err, "layer", "service.impl")
	}

	return nil

}
