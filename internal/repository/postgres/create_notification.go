package postgres

import (
	"Chronos/internal/models"
	"context"

	"github.com/wb-go/wbf/retry"
)

func (s *Storage) CreateNotification(ctx context.Context, notification models.Notification) error {

	query := `

	INSERT INTO Notifications (uuid, channel, message, status, send_at, send_to, created_at, updated_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`

	_, err := s.db.ExecWithRetry(
		ctx,
		retry.Strategy{Attempts: 3, Delay: 10, Backoff: 3},
		query,
		notification.ID,
		notification.Channel,
		notification.Message,
		notification.Status,
		notification.SendAt,
		notification.SendTo,
		notification.CreatedAt,
		notification.UpdatedAt,
	)

	return err

}
