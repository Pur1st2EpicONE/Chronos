package postgres

import (
	"Chronos/internal/errs"
	"Chronos/internal/models"
	"context"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type Storage struct {
	db  *dbpg.DB
	log zlog.Zerolog
}

func NewStorage(db *dbpg.DB) *Storage {
	return &Storage{db: db, log: zlog.Logger.With().Str("layer", "repository.postgres").Logger()}
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

func (s *Storage) CancelNotification(ctx context.Context, id int64) error {

	query := `
    
	UPDATE Notifications
    SET status = 'canceled', updated_at = NOW()
    WHERE id = $1 AND status = 'pending';`

	res, err := s.db.ExecWithRetry(ctx, retry.Strategy{Attempts: 3, Delay: 10, Backoff: 3}, query, id)
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
		s.log.Err(err).Msg("postgres — failed to close properly")
	} else {
		s.log.Info().Msg("postgres — stopped")
	}
}

// docker exec -it postgres psql -U Neo -d chronos-db
