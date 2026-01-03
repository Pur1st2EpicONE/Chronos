package postgres

import (
	"Chronos/internal/models"
	"context"
	"fmt"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

// Recover retrieves notifications that need to be re-queued or retried.
// It is called during service initialization and when the broker recovers from a failure.
// The method fetches all notifications that were scheduled to be sent but could not
// be delivered while the broker was unavailable.
func (s *Storage) Recover(ctx context.Context) ([]models.Notification, error) {

	var notifications []models.Notification

	query := `
	
		WITH recipients_agg AS (
			SELECT notification_uuid, array_agg(recipient) AS send_to
			FROM recipients
			GROUP BY notification_uuid
		)

		SELECT n.uuid, n.channel, n.message, n.status, n.send_at, n.send_at_local, r.send_to, n.updated_at
		FROM notifications n
		LEFT JOIN recipients_agg r
	    ON n.uuid = r.notification_uuid
		WHERE n.status IN ($1, $2)
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
			&n.Status, &n.SendAt, &n.SendAtLocal, dbpg.Array(&n.SendTo),
			&n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, nil

}
