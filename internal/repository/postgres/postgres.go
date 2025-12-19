package postgres

import (
	"Chronos/internal/config"
	"Chronos/internal/logger"

	"github.com/wb-go/wbf/dbpg"
)

type Storage struct {
	db     *dbpg.DB
	logger logger.Logger
	config config.Storage
}

func NewStorage(logger logger.Logger, config config.Storage, db *dbpg.DB) *Storage {
	return &Storage{db: db, logger: logger, config: config}
}

func (s *Storage) Close() {
	if err := s.db.Master.Close(); err != nil {
		s.logger.LogError("postgres — failed to close properly", err, "layer", "repository.postgres")
	} else {
		s.logger.LogInfo("postgres — database closed", "layer", "repository.postgres")
	}
}

// docker exec -it postgres psql -U Neo -d chronos-db
