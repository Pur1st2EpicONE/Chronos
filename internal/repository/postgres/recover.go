package postgres

import (
	"Chronos/internal/models"
	"context"
	"fmt"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

func (s *Storage) Recover(ctx context.Context) ([]models.Notification, error) {

	var notifications []models.Notification

	query := `
	
	SELECT n.id, n.channel, n.message, n.status, n.send_at, array_agg(r.recipient) AS send_to, n.updated_at
	FROM Notifications n
	JOIN Recipients r 
	ON n.uuid = r.notification_uuid
	WHERE n.status = $1 OR n.status = $2
	GROUP BY n.id
	ORDER BY n.send_at ASC
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
			&n.Status, &n.SendAt, dbpg.Array(&n.SendTo),
			&n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, nil

}
