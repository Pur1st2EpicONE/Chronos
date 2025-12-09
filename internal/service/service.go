package service

import (
	"Chronos/internal/config"
	"Chronos/internal/models"
	"Chronos/internal/repository"
	"Chronos/internal/service/impl"
	"context"
)

type Service interface {
	CreateNotification(ctx context.Context, notification models.Notification) (int64, error)
	GetNotification(ctx context.Context, notificationID int64) (string, error)
	CancelNotification(ctx context.Context, id int64) error
}

func NewService(config config.Service, storage repository.Storage) Service {
	return impl.NewService(config, storage)
}
