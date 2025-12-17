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

type Storage interface {
	CreateNotification(ctx context.Context, notification models.Notification) error
	DeleteNotification(ctx context.Context, notificationID string) error
	GetStatus(ctx context.Context, notificationID string) (string, error)
	SetStatus(ctx context.Context, notificationID string, status string) error
	MarkLates(ctx context.Context) ([]string, error)
	Recover(ctx context.Context) ([]models.Notification, error)
	Close()
}

func NewStorage(logger logger.Logger, db *dbpg.DB) Storage {
	return postgres.NewStorage(logger, db)
}

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
