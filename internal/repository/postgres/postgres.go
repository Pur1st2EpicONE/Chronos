package postgres

import (
	"Chronos/internal/errs"
	"Chronos/internal/logger"
	"Chronos/internal/models"
	"context"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type Storage struct {
	db     *dbpg.DB
	logger logger.Logger
}

func NewStorage(logger logger.Logger, db *dbpg.DB) *Storage {
	return &Storage{db: db, logger: logger}
}

func (s *Storage) CreateNotification(ctx context.Context, notification models.Notification) (int64, error) {

	query := `

	INSERT INTO Notifications (channel, message, send_at, send_to) 
	VALUES ($1, $2, $3, $4)
	RETURNING id;`

	row, err := s.db.QueryRowWithRetry(
		ctx,
		retry.Strategy{Attempts: 3, Delay: 10, Backoff: 3},
		query,
		notification.Channel,
		notification.Message,
		notification.SendAt,
		notification.SendTo,
	)
	if err != nil {
		return 0, err
	}

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil

}

func (s *Storage) GetStatus(ctx context.Context, notificationID int64) (string, error) {

	query := `

    SELECT status
    FROM Notifications
    WHERE id = $1;`

	row, err := s.db.QueryRowWithRetry(ctx, retry.Strategy{Attempts: 3, Delay: 10, Backoff: 3}, query, notificationID)
	if err != nil {
		return "", err
	}

	var status string
	if err := row.Scan(&status); err != nil {
		return "", errs.ErrNotificationNotFound
	}

	return status, nil

}

func (s *Storage) SetStatus(ctx context.Context, notificationID int64, status string) error {

	query := `
    
	UPDATE Notifications
    SET status = $1, updated_at = NOW()
    WHERE id = $2;`

	res, err := s.db.ExecWithRetry(ctx, retry.Strategy{Attempts: 3, Delay: 10, Backoff: 3}, query, status, notificationID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errs.ErrNotificationNotFound
	}

	return nil

}

func (s *Storage) Close() {
	if err := s.db.Master.Close(); err != nil {
		s.logger.LogError("postgres — failed to close properly", err, "layer", "repository.postgres")
	} else {
		s.logger.LogInfo("postgres — stopped", "layer", "repository.postgres")
	}
}

// docker exec -it postgres psql -U Neo -d chronos-db
