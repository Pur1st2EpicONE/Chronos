package service

import (
	"Chronos/internal/broker"
	"Chronos/internal/models"
	"Chronos/internal/repository"
	"Chronos/internal/service/impl"
	"context"
)

type Service interface {
	CreateNotification(ctx context.Context, notification models.Notification) (int64, error)
	GetStatus(ctx context.Context, notificationID int64) (string, error)
	CancelNotification(ctx context.Context, notificationID int64) error
}

func NewService(broker broker.Broker, storage repository.Storage) Service {
	return impl.NewService(broker, storage)
}
