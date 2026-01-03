package impl

import (
	"Chronos/internal/errs"
	"context"
	"errors"
)

// GetStatus retrieves the current status of a notification.
// It first checks the cache, and if not found, falls back to the database.
// After fetching from the database, the status is updated in the cache for future calls.
// This ensures a single source of truth while optimizing for read performance.
func (s *Service) GetStatus(ctx context.Context, notificationID string) (string, error) {

	status, err := s.cache.GetStatus(ctx, notificationID)
	if err != nil {

		status, err = s.storage.GetStatus(ctx, notificationID)
		if err != nil {
			if errors.Is(err, errs.ErrNotificationNotFound) {
				s.logger.Debug("service — notification status fetched from DB", "notificationID", notificationID, "layer", "service.impl")
			}
			s.logger.LogError("service — failed to get notification status from DB", err, "notificationID", notificationID, "layer", "service.impl")
			return "", err
		}

		if err := s.cache.SetStatus(ctx, notificationID, status); err != nil {
			s.logger.LogError("service — failed to set notification status in cache", err, "notificationID", notificationID, "layer", "service.impl")
		}

		s.logger.Debug("service — notification status fetched from DB", "notificationID", notificationID, "layer", "service.impl")
	} else {
		s.logger.Debug("service — notification status fetched from cache", "notificationID", notificationID, "layer", "service.impl")
	}

	return status, nil

}
