package impl

import "context"

func (s *Service) GetStatus(ctx context.Context, notificationID string) (string, error) {

	status, err := s.cache.GetStatus(ctx, notificationID)
	if err != nil {

		status, err = s.storage.GetStatus(ctx, notificationID)
		if err != nil {
			return "", err
		}

		if err := s.cache.SetStatus(ctx, notificationID, status); err != nil {
			s.logger.LogError("service — failed to set notification status in cache", err, "layer", "service.impl")
		}

	}

	return status, nil

}
