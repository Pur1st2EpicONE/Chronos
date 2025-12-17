package postgres

import (
	"Chronos/internal/models"
	"context"

	"github.com/wb-go/wbf/retry"
)

func (s *Storage) MarkLates(ctx context.Context) ([]string, error) {

	query := `

        UPDATE Notifications
        SET status = $1, updated_at = NOW()
        WHERE status = $2 AND send_at < NOW()
        RETURNING uuid;`

	rows, err := s.db.QueryWithRetry(ctx, retry.Strategy{Attempts: 3, Delay: 10, Backoff: 3}, query, models.StatusLate, models.StatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uuids []string

	for rows.Next() {
		var uuid string
		if err := rows.Scan(&uuid); err != nil {
			return nil, err
		}
		uuids = append(uuids, uuid)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return uuids, nil

}
