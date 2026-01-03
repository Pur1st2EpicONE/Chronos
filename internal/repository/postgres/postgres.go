// Package postgres provides a PostgreSQL-backed implementation of the repository layer.
package postgres

import (
	"Chronos/internal/config"
	"Chronos/internal/logger"

	"github.com/wb-go/wbf/dbpg"
)

// Storage implements the Storage interface using PostgreSQL.
type Storage struct {
	db     *dbpg.DB       // Database connection pool
	logger logger.Logger  // Application logger
	config config.Storage // Storage-related configuration
}

// NewStorage creates a new PostgreSQL storage instance.
func NewStorage(logger logger.Logger, config config.Storage, db *dbpg.DB) *Storage {
	return &Storage{db: db, logger: logger, config: config}
}

// Close closes the database connection.
func (s *Storage) Close() {
	if err := s.db.Master.Close(); err != nil {
		s.logger.LogError("postgres — failed to close properly", err, "layer", "repository.postgres")
	} else {
		s.logger.LogInfo("postgres — database closed", "layer", "repository.postgres")
	}
}

// DB returns the underlying database instance.
// Exposed primarily for testing purposes.
func (s *Storage) DB() *dbpg.DB {
	return s.db
}

// Config returns the storage configuration.
// Exposed primarily for testing purposes.
func (s *Storage) Config() *config.Storage {
	return &s.config
}
