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

	INSERT INTO Notifications (channel, message, status, send_at, send_to, created_at, updated_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id;`

	row, err := s.db.QueryRowWithRetry(
		ctx,
		retry.Strategy{Attempts: 3, Delay: 10, Backoff: 3},
		query,
		notification.Channel,
		notification.Message,
		notification.Status,
		notification.SendAt,
		notification.SendTo,
		notification.CreatedAt,
		notification.UpdatedAt,
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

func (s *Storage) MarkLate(ctx context.Context) error {

	query := `
    
	UPDATE Notifications
    SET status = $1, updated_at = NOW()
    WHERE status = $2 AND send_at < NOW();`

	_, err := s.db.ExecWithRetry(ctx, retry.Strategy{Attempts: 3, Delay: 10, Backoff: 3}, query, models.StatusLate, models.StatusPending)
	if err != nil {
		return err
	}

	return nil

}

func (s *Storage) Recover(ctx context.Context) ([]models.Notification, error) {

	var notifications []models.Notification

	query := `
	
	SELECT * FROM Notifications
	WHERE status = $1 OR status = $2;`

	rows, err := s.db.QueryContext(ctx, query, models.StatusPending, models.StatusLate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(
			&n.ID,
			&n.Channel,
			&n.Message,
			&n.Status,
			&n.SendAt,
			&n.SendTo,
			&n.CreatedAt,
			&n.UpdatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil

}

func (s *Storage) Close() {
	if err := s.db.Master.Close(); err != nil {
		s.logger.LogError("postgres — failed to close properly", err, "layer", "repository.postgres")
	} else {
		s.logger.LogInfo("postgres — database closed", "layer", "repository.postgres")
	}
}

// docker exec -it postgres psql -U Neo -d chronos-db
