package app

import (
	"Chronos/internal/broker"
	"Chronos/internal/cache"
	"Chronos/internal/config"
	"Chronos/internal/handler"
	"Chronos/internal/logger"
	"Chronos/internal/notifier"
	"Chronos/internal/repository"
	"Chronos/internal/server"
	"Chronos/internal/service"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"context"

	"github.com/wb-go/wbf/dbpg"
)

type App struct {
	logger  logger.Logger
	logFile *os.File
	broker  broker.Broker
	server  server.Server
	ctx     context.Context
	cancel  context.CancelFunc
	cache   cache.Cache
	storage repository.Storage
}

func Boot() *App {

	config, err := config.Load()
	if err != nil {
		log.Fatalf("app — failed to load configs: %v", err)
	}

	logger, logFile := logger.NewLogger(config.Logger)

	db, err := connectDB(logger, config.Storage)
	if err != nil {
		logger.LogFatal("app — failed to connect to database", err, "layer", "app")
	}

	cache, err := connectCache(logger, config.Cache)
	if err != nil {
		logger.LogFatal("app — failed to connect to cache", err, "layer", "app")
	}

	app, err := wireApp(db, cache, logger, logFile, config)
	if err != nil {
		logger.LogFatal("app — failed to connect to broker", err, "layer", "app")
	}

	return app

}

func connectDB(logger logger.Logger, config config.Storage) (*dbpg.DB, error) {
	db, err := repository.ConnectDB(config)
	if err != nil {
		return nil, err
	}
	logger.LogInfo("app — connected to database", "layer", "app")
	return db, nil
}

func connectCache(logger logger.Logger, config config.Cache) (cache.Cache, error) {
	cache, err := cache.Connect(logger, config)
	if err != nil {
		return nil, err
	}
	logger.LogInfo("app — connected to cache", "layer", "app")
	return cache, nil
}

func wireApp(db *dbpg.DB, cache cache.Cache, logger logger.Logger, logFile *os.File, config config.Config) (*App, error) {

	ctx, cancel := newContext(logger)
	storge := repository.NewStorage(logger, db)
	notifier := notifier.NewNotifier(config.Notifier)
	broker, err := broker.NewBroker(logger, config.Broker, cache, storge, notifier)
	service := service.NewService(logger, broker, cache, storge)
	handler := handler.NewHandler(service)
	server := server.NewServer(logger, config.Server, handler)

	if err != nil {
		return nil, err
	}

	return &App{
		logger:  logger,
		logFile: logFile,
		broker:  broker,
		server:  server,
		ctx:     ctx,
		cancel:  cancel,
		cache:   cache,
		storage: storge,
	}, nil

}

func newContext(logger logger.Logger) (context.Context, context.CancelFunc) {

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sig := <-sigCh
		logger.LogInfo("app — received signal "+sig.String()+", initiating graceful shutdown", "layer", "app")
		cancel()
	}()

	return ctx, cancel

}

func (a *App) Run() {

	go func() {
		if err := a.server.Run(); err != nil {
			a.logger.LogFatal("server run failed", err, "layer", "app")
		}
	}()

	var wg sync.WaitGroup

	wg.Go(func() {
		if err := a.broker.Consume(a.ctx); err != nil {
			a.logger.LogFatal("consumer run failed", err, "layer", "app")
		}
	})

	<-a.ctx.Done()
	wg.Wait()

	a.Stop()

}

func (a *App) Stop() {
	a.server.Shutdown()
	a.cache.Close()
	a.storage.Close()
	if a.logFile != nil {
		_ = a.logFile.Close()
	}
}
