// Package repository provides an abstraction layer over the database for
// creating, updating, retrieving, and deleting notifications.
package repository

import (
	"Chronos/internal/config"
	"Chronos/internal/logger"
	"Chronos/internal/models"
	"Chronos/internal/repository/postgres"
	"context"
	"fmt"

	"github.com/wb-go/wbf/dbpg"
)

// Storage defines the interface for interacting with the notifications storage.
// It abstracts database operations such as creating notifications, retrieving their status,
// marking late notifications, and performing cleanup.
type Storage interface {
	CreateNotification(ctx context.Context, notification models.Notification) error // CreateNotification inserts a new notification into the storage.
	DeleteNotification(ctx context.Context, notificationID string) error            // DeleteNotification removes a notification by its ID.
	GetStatus(ctx context.Context, notificationID string) (string, error)           // GetStatus returns the current status of a notification by its ID.
	GetAllStatuses(ctx context.Context) ([]models.Notification, error)              // GetAllStatuses returns all notifications and their statuses.
	SetStatus(ctx context.Context, notificationID string, status string) error      // SetStatus updates the status of a notification.
	MarkLates(ctx context.Context) ([]string, error)                                // MarkLates marks notifications that are late in the database and returns their IDs.
	Recover(ctx context.Context) ([]models.Notification, error)                     // Recover returns pending or late notifications for re-queuing.
	Cleanup(ctx context.Context)                                                    // Cleanup performs periodic cleanup tasks, such as removing expired notifications.
	Close()                                                                         // Close closes the storage connection.
}

// NewStorage creates a new Storage instance backed by Postgres.
func NewStorage(logger logger.Logger, config config.Storage, db *dbpg.DB) Storage {
	return postgres.NewStorage(logger, config, db)
}

// ConnectDB establishes a connection to the Postgres database using the provided configuration.
// It returns a dbpg.DB instance ready for queries.
func ConnectDB(config config.Storage) (*dbpg.DB, error) {

	options := &dbpg.Options{
		MaxOpenConns:    config.MaxOpenConns,
		MaxIdleConns:    config.MaxIdleConns,
		ConnMaxLifetime: config.ConnMaxLifetime,
	}

	db, err := dbpg.New(fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.DBName, config.SSLMode), nil, options)
	if err != nil {
		return nil, fmt.Errorf("database driver not found or DSN invalid: %w", err)
	}

	if err := db.Master.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return db, nil

}
