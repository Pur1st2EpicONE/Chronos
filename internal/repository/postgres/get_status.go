package postgres

import (
	"Chronos/internal/errs"
	"context"

	"github.com/wb-go/wbf/retry"
)

func (s *Storage) GetStatus(ctx context.Context, notificationID string) (string, error) {

	query := `

    SELECT status
    FROM Notifications
    WHERE uuid = $1;`

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
