package postgres

import (
	"Chronos/internal/errs"
	"context"
	"fmt"

	"github.com/wb-go/wbf/retry"
)

// DeleteNotification deletes a notification by its UUID.
// The deletion is attempted directly; if no rows are affected,
// it returns ErrNotificationNotFound. This avoids an extra query
// to check existence before deleting.
func (s *Storage) DeleteNotification(ctx context.Context, notificationID string) error {

	query := `
	
	DELETE FROM Notifications 
	WHERE uuid = $1;`

	result, err := s.db.ExecWithRetry(ctx, retry.Strategy{
		Attempts: s.config.QueryRetryStrategy.Attempts,
		Delay:    s.config.QueryRetryStrategy.Delay,
		Backoff:  s.config.QueryRetryStrategy.Backoff}, query, notificationID)

	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get number of affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return errs.ErrNotificationNotFound
	}

	return nil

}
