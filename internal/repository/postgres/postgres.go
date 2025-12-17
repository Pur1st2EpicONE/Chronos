package postgres

import (
	"Chronos/internal/logger"

	"github.com/wb-go/wbf/dbpg"
)

type Storage struct {
	db     *dbpg.DB
	logger logger.Logger
}

func NewStorage(logger logger.Logger, db *dbpg.DB) *Storage {
	return &Storage{db: db, logger: logger}
}

func (s *Storage) Close() {
	if err := s.db.Master.Close(); err != nil {
		s.logger.LogError("postgres — failed to close properly", err, "layer", "repository.postgres")
	} else {
		s.logger.LogInfo("postgres — database closed", "layer", "repository.postgres")
	}
}

// docker exec -it postgres psql -U Neo -d chronos-db
