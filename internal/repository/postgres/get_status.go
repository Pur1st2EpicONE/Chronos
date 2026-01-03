package postgres

import (
	"Chronos/internal/errs"
	"context"
	"fmt"

	"github.com/wb-go/wbf/retry"
)

// GetStatus returns the current status of a notification by its ID.
func (s *Storage) GetStatus(ctx context.Context, notificationID string) (string, error) {

	query := `

    SELECT status
    FROM Notifications
    WHERE uuid = $1;`

	row, err := s.db.QueryRowWithRetry(ctx, retry.Strategy{
		Attempts: s.config.QueryRetryStrategy.Attempts,
		Delay:    s.config.QueryRetryStrategy.Delay,
		Backoff:  s.config.QueryRetryStrategy.Backoff}, query, notificationID)

	if err != nil {
		return "", fmt.Errorf("failed to execute query: %w", err)
	}

	var status string
	if err := row.Scan(&status); err != nil {
		return "", errs.ErrNotificationNotFound
	}

	return status, nil

}
