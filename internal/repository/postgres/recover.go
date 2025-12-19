package postgres

import (
	"Chronos/internal/models"
	"context"
	"fmt"
)

func (s *Storage) Recover(ctx context.Context) ([]models.Notification, error) {

	var notifications []models.Notification

	query := `
	
	SELECT uuid, channel, message, status, send_at, send_to, created_at, updated_at
	FROM Notifications
	WHERE status = $1 OR status = $2
	ORDER BY send_at ASC
    LIMIT 1000;`

	rows, err := s.db.QueryContext(ctx, query, models.StatusPending, models.StatusLate)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(
			&n.ID,
			&n.Channel,
			&n.Message,
			&n.Status,
			&n.SendAt,
			&n.SendTo,
			&n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, nil

}
