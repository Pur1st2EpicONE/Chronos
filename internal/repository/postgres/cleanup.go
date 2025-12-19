package postgres

import (
	"Chronos/internal/models"
	"context"

	"github.com/wb-go/wbf/retry"
)

func (s *Storage) Cleanup(ctx context.Context) {

	query := `
	
        DELETE FROM Notifications 
        WHERE (status = $1 AND updated_at < NOW() - $2 * INTERVAL '1 second')
        OR (status = $3 AND updated_at < NOW() - $4 * INTERVAL '1 second')
        OR (status IN ($5, $6) AND updated_at < NOW() - $7 * INTERVAL '1 second');`

	_, err := s.db.ExecWithRetry(ctx, retry.Strategy{
		Attempts: s.config.QueryRetryStrategy.Attempts,
		Delay:    s.config.QueryRetryStrategy.Delay,
		Backoff:  s.config.QueryRetryStrategy.Backoff}, query,

		models.StatusCanceled, int(s.config.RetentionStrategy.Canceled.Seconds()),
		models.StatusSent, int(s.config.RetentionStrategy.Completed.Seconds()),
		models.StatusFailed, models.StatusFailedToSendInTime, int(s.config.RetentionStrategy.Failed.Seconds()),
	)

	if err != nil {
		s.logger.LogError("postgres — failed to delete old notifications", err, "layer", "repository.postgres")
	} else {
		s.logger.Debug("postgres — old notifications cleaned", "layer", "repository.postgres")
	}

}
