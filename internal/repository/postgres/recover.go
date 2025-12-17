package postgres

import (
	"Chronos/internal/models"
	"context"
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
	defer rows.Close()

	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(
			&n.ID,
			&n.Channel,
			&n.Message,
			&n.Status,
			&n.SendAt,
			&n.SendTo,
			&n.CreatedAt,
			&n.UpdatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil

}
