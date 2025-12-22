package postgres

import (
	"Chronos/internal/models"
	"context"
	"database/sql"
	"fmt"

	"github.com/wb-go/wbf/retry"
)

func (s *Storage) CreateNotification(ctx context.Context, notification models.Notification) error {

	strategy := retry.Strategy{
		Attempts: s.config.QueryRetryStrategy.Attempts,
		Delay:    s.config.QueryRetryStrategy.Delay,
		Backoff:  s.config.QueryRetryStrategy.Backoff,
	}

	notificationsQuery := `

			INSERT INTO Notifications (uuid, channel, message, status, send_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6);`

	recipientsQuery := `

			INSERT INTO Recipients (notification_uuid, recipient)
			VALUES ($1, $2);`

	return s.db.WithTxWithRetry(ctx, strategy, func(tx *sql.Tx) error {

		_, err := tx.ExecContext(ctx, notificationsQuery,
			notification.ID, notification.Channel,
			notification.Message, notification.Status,
			notification.SendAt, notification.UpdatedAt)

		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}

		if notification.Channel == models.Email {

			for _, recipient := range notification.SendTo {
				if _, err := tx.ExecContext(ctx, recipientsQuery, notification.ID, recipient); err != nil {
					return fmt.Errorf("failed to execute query: %w", err)
				}
			}

		}

		return nil

	})

}
