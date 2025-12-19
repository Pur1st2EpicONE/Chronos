package postgres

import (
	"Chronos/internal/errs"
	"context"
	"fmt"

	"github.com/wb-go/wbf/retry"
)

func (s *Storage) DeleteNotification(ctx context.Context, notificationID string) error {

	query := `
	
	DELETE FROM Notifications 
	WHERE uuid = $1;`

	result, err := s.db.ExecWithRetry(ctx, retry.Strategy{Attempts: 3, Delay: 10, Backoff: 3}, query, notificationID)
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
