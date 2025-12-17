package postgres

import (
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
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}
