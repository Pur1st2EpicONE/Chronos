package postgres

import (
	"Chronos/internal/models"
	"context"

	"github.com/wb-go/wbf/retry"
)

func (s *Storage) Cleanup(ctx context.Context) {

	query := `
	
	DELETE FROM Notifications 
	WHERE status = $1;`

	_, err := s.db.ExecWithRetry(ctx, retry.Strategy{Attempts: 3, Delay: 10, Backoff: 3}, query, models.StatusCanceled)
	if err != nil {
		s.logger.LogError("postgres — failed to delete canceled notifications", err, "layer", "repository.postgres")
	} else {
		s.logger.Debug("postgres — canceled notifications cleaned", "layer", "repository.postgres")
	}

}
