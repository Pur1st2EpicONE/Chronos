package impl

import (
	"Chronos/internal/config"
	"Chronos/internal/models"
	"Chronos/internal/repository"
	"context"

	"github.com/wb-go/wbf/zlog"
)

type Service struct {
	Storage repository.Storage
	logger  zlog.Zerolog
}

func NewService(config config.Service, storage repository.Storage) *Service {
	return &Service{Storage: storage, logger: zlog.Logger.With().Str("layer", "service.impl").Logger()}
}

func (s *Service) CreateNotification(ctx context.Context, notification models.Notification) (int64, error) {
	return s.Storage.CreateNotification(ctx, notification)
}
