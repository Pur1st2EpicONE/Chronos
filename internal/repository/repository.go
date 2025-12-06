package repository

import (
	"Chronos/internal/config"

	"github.com/wb-go/wbf/dbpg"
)

type Storage interface {
	Close()
}

func NewStorage(db *dbpg.DB, config config.Storage) Storage {
	return nil
}

func ConnectDB(config config.Storage) (*dbpg.DB, error) {
	options := &dbpg.Options{
		MaxOpenConns:    config.MaxOpenConns,
		MaxIdleConns:    config.MaxIdleConns,
		ConnMaxLifetime: config.ConnMaxLifetime,
	}
	db, err := dbpg.New(config.MasterDSN, nil, options)
	if err != nil {
		return nil, err
	}
	return db, nil
}
