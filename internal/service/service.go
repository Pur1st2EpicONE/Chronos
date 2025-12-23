package service

import (
	"Chronos/internal/broker"
	"Chronos/internal/cache"
	"Chronos/internal/logger"
	"Chronos/internal/models"
	"Chronos/internal/repository"
	"Chronos/internal/service/impl"
	"context"
)

type Service interface {
	CreateNotification(ctx context.Context, notification models.Notification) (string, error)
	GetAllStatuses(ctx context.Context) []models.Notification
	GetStatus(ctx context.Context, notificationID string) (string, error)
	CancelNotification(ctx context.Context, notificationID string) error
}

func NewService(logger logger.Logger, broker broker.Broker, cache cache.Cache, storage repository.Storage) Service {
	return impl.NewService(logger, broker, cache, storage)
}
