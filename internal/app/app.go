package app

import (
	"Chronos/internal/broker"
	"Chronos/internal/config"
	"Chronos/internal/handler"
	"Chronos/internal/logger"
	"Chronos/internal/repository"
	"Chronos/internal/server"
	"Chronos/internal/service"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"context"
	"errors"
	"net/http"

	"github.com/wb-go/wbf/dbpg"
)

type App struct {
	logger  logger.Logger
	logFile *os.File
	broker  broker.Broker
	server  server.Server
	ctx     context.Context
	cancel  context.CancelFunc
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

	app, err := newApp(db, logger, logFile, config)
	if err != nil {
		logger.LogFatal("app — failed to initialize", err, "layer", "app")
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

func newApp(db *dbpg.DB, logger logger.Logger, logFile *os.File, config config.Config) (*App, error) {

	ctx, cancel := newContext(logger)

	storage := repository.NewStorage(logger, db)
	broker, err := broker.NewBroker(logger, config.Broker, storage)
	if err != nil {
		return nil, err
	}

	service := service.NewService(broker, storage)
	handler := handler.NewHandler(service)
	server := server.NewServer(logger, config.Server, handler)

	return &App{
		logger:  logger,
		logFile: logFile,
		broker:  broker,
		server:  server,
		ctx:     ctx,
		cancel:  cancel,
		storage: storage,
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
		if err := a.server.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
	a.storage.Close()
	if a.logFile != nil {
		_ = a.logFile.Close()
	}
}
