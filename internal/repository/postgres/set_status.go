package postgres

import (
	"Chronos/internal/errs"
	"Chronos/internal/models"
	"context"

	"github.com/wb-go/wbf/retry"
)

func (s *Storage) SetStatus(ctx context.Context, notificationID string, status string) error {

	query := `
    
	UPDATE Notifications
    SET status = $1, updated_at = NOW()
    WHERE uuid = $2;`

	var args []any
	args = append(args, status, notificationID)

	if status == models.StatusCanceled {
		query = query[:len(query)-1] + " AND status IN ($3, $4);"
		args = append(args, models.StatusPending, models.StatusLate)
	}

	res, err := s.db.ExecWithRetry(ctx, retry.Strategy{Attempts: 3, Delay: 10, Backoff: 3}, query, args...)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		if status == models.StatusCanceled {
			return errs.ErrCannotCancel
		}
		return errs.ErrNotificationNotFound
	}

	return nil

}
