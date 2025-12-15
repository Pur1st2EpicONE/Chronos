package impl

import (
	"Chronos/internal/broker"
	"Chronos/internal/models"
	"Chronos/internal/repository"
	"context"
	"time"

	"github.com/wb-go/wbf/zlog"
)

type Service struct {
	logger  zlog.Zerolog
	broker  broker.Broker
	storage repository.Storage
}

func NewService(broker broker.Broker, storage repository.Storage) *Service {
	return &Service{storage: storage, broker: broker, logger: zlog.Logger.With().Str("layer", "service.impl").Logger()}
}

func (s *Service) CreateNotification(ctx context.Context, notification models.Notification) (int64, error) {

	var err error

	if err = validateCreate(notification); err != nil {
		return 0, err
	}

	now := time.Now().UTC()
	notification.CreatedAt = now
	notification.UpdatedAt = now
	notification.Status = models.StatusPending

	notification.ID, err = s.storage.CreateNotification(ctx, notification)
	if err != nil {
		return 0, err
	}

	_ = s.broker.Produce(ctx, notification)

	return notification.ID, nil

}

func (s *Service) GetStatus(ctx context.Context, notificationID int64) (string, error) {
	return s.storage.GetStatus(ctx, notificationID)
}

func (s *Service) CancelNotification(ctx context.Context, notificationID int64) error {
	return s.storage.SetStatus(ctx, notificationID, models.StatusCanceled)
}
