package impl

import (
	"Chronos/internal/broker"
	"Chronos/internal/cache"
	"Chronos/internal/logger"
	"Chronos/internal/repository"
)

type Service struct {
	logger  logger.Logger
	broker  broker.Broker
	cache   cache.Cache
	storage repository.Storage
}

func NewService(logger logger.Logger, broker broker.Broker, cache cache.Cache, storage repository.Storage) *Service {
	return &Service{logger: logger, broker: broker, cache: cache, storage: storage}
}
