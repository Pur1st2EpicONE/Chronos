// Package service contains the business logic for notifications.
// It coordinates operations between the broker, cache, and storage layers.
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

// Service defines the business logic interface for notifications.
// It orchestrates operations across the broker, cache, and storage layers.
type Service interface {
	CreateNotification(ctx context.Context, notification models.Notification) (string, error) // CreateNotification creates a new notification and returns its ID.
	GetAllStatuses(ctx context.Context) []models.Notification                                 // GetAllStatuses retrieves all notifications with their current status. Used for frontend display; not optimized.
	GetStatus(ctx context.Context, notificationID string) (string, error)                     // GetStatus returns the current status of a specific notification by ID.
	CancelNotification(ctx context.Context, notificationID string) error                      // CancelNotification attempts to cancel a notification by ID.
}

// NewService constructs a new Service instance with all dependencies injected.
func NewService(logger logger.Logger, broker broker.Broker, cache cache.Cache, storage repository.Storage) Service {
	return impl.NewService(logger, broker, cache, storage)
}
