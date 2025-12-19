package postgres

import (
	"Chronos/internal/models"
	"context"

	"github.com/wb-go/wbf/retry"
)

func (s *Storage) CreateNotification(ctx context.Context, notification models.Notification) error {

	query := `

	INSERT INTO Notifications (uuid, channel, message, status, send_at, send_to, updated_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7);`

	_, err := s.db.ExecWithRetry(ctx, retry.Strategy{
		Attempts: s.config.QueryRetryStrategy.Attempts,
		Delay:    s.config.QueryRetryStrategy.Delay,
		Backoff:  s.config.QueryRetryStrategy.Backoff}, query,

		notification.ID,
		notification.Channel,
		notification.Message,
		notification.Status,
		notification.SendAt,
		notification.SendTo,
		notification.UpdatedAt)

	return err

}
