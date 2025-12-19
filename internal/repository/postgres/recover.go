package postgres

import (
	"Chronos/internal/models"
	"context"
	"fmt"

	"github.com/wb-go/wbf/retry"
)

func (s *Storage) Recover(ctx context.Context) ([]models.Notification, error) {

	var notifications []models.Notification

	query := `
	
	SELECT uuid, channel, message, status, send_at, send_to, updated_at
	FROM Notifications
	WHERE status = $1 OR status = $2
	ORDER BY send_at ASC
    LIMIT $3;`

	rows, err := s.db.QueryWithRetry(ctx, retry.Strategy{
		Attempts: s.config.QueryRetryStrategy.Attempts,
		Delay:    s.config.QueryRetryStrategy.Delay,
		Backoff:  s.config.QueryRetryStrategy.Backoff}, query,

		models.StatusPending, models.StatusLate,
		s.config.RecoverLimit)

	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(
			&n.ID, &n.Channel, &n.Message,
			&n.Status, &n.SendAt, &n.SendTo,
			&n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, nil

}
