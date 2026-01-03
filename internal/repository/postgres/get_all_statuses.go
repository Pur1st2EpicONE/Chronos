package postgres

import (
	"Chronos/internal/models"
	"context"
	"fmt"

	"github.com/wb-go/wbf/retry"
)

// GetAllStatuses returns all notifications with their scheduling timestamps
// and current status, ordered by send time.
// Note: This method is intended only for the web frontend and is not optimized
// for API usage or large datasets.
func (s *Storage) GetAllStatuses(ctx context.Context) ([]models.Notification, error) {

	var notifications []models.Notification

	query := `
	
	SELECT uuid, send_at, send_at_local, status 
	FROM Notifications
	ORDER BY send_at ASC;`

	rows, err := s.db.QueryWithRetry(ctx, retry.Strategy{
		Attempts: s.config.QueryRetryStrategy.Attempts,
		Delay:    s.config.QueryRetryStrategy.Delay,
		Backoff:  s.config.QueryRetryStrategy.Backoff}, query)

	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(&n.ID, &n.SendAt, &n.SendAtLocal, &n.Status); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, nil

}
