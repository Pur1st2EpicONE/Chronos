package impl

import (
	"Chronos/internal/models"
	"context"
)

func (s *Service) GetAllStatuses(ctx context.Context) []models.Notification {
	statuses, err := s.storage.GetAllStatuses(ctx)
	if err != nil {
		s.logger.LogError("service — failed to get notification statuses from DB", err, "layer", "service.impl")
	}
	return statuses
}
