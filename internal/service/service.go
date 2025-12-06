package service

import (
	"Chronos/internal/config"
	"Chronos/internal/repository"
)

type Service any

func NewService(config config.Service, storage repository.Storage) Service {
	return nil
}
