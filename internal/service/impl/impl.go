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

	if err := validateCreate(notification); err != nil {
		return 0, err
	}

	return s.Storage.CreateNotification(ctx, notification)

}

func (s *Service) GetNotification(ctx context.Context, notificationID int64) (string, error) {
	return s.Storage.GetNotification(ctx, notificationID)
}

func (s *Service) CancelNotification(ctx context.Context, notificationID int64) error {
	return s.Storage.CancelNotification(ctx, notificationID)
}
