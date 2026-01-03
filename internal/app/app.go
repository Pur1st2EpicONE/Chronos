// Package app wires application components, provides lifecycle management
// and exposes the entry point for booting and running the service.
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

// App represents the application's composition root.
// It holds long-lived resources (logger, DB, cache, broker, server) and
// the context/cancel function used for graceful shutdown.
type App struct {
	logger  logger.Logger      // logger is the structured logger used across application layers.
	logFile *os.File           // logFile is the file handle where logs are written.
	broker  broker.Broker      // broker is the message broker used for consuming/producing domain messages.
	server  server.Server      // server is the HTTP server instance.
	ctx     context.Context    // ctx is the root context used to coordinate shutdown across components.
	cancel  context.CancelFunc // cancel cancels the root context when a shutdown signal is received.
	cache   cache.Cache        // cache is the cache layer used by services (e.g., redis).
	storage repository.Storage // storage is the data storage abstraction backed by the database.
}

// Boot loads configuration, initializes logger, connects to database and cache,
// wires all components and returns a fully constructed *App ready to run.
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

// connectDB establishes a database connection using repository.ConnectDB
// and logs successful connection.
func connectDB(logger logger.Logger, config config.Storage) (*dbpg.DB, error) {
	db, err := repository.ConnectDB(config)
	if err != nil {
		return nil, err
	}
	logger.LogInfo("app — connected to database", "layer", "app")
	return db, nil
}

// connectCache establishes a connection to the cache backend using cache.Connect
// and logs successful connection.
func connectCache(logger logger.Logger, config config.Cache) (cache.Cache, error) {
	cache, err := cache.Connect(logger, config)
	if err != nil {
		return nil, err
	}
	logger.LogInfo("app — connected to cache", "layer", "app")
	return cache, nil
}

// wireApp constructs application components (storage, notifier, broker, service,
// handler, server), creates a cancellable context and returns the assembled *App.
func wireApp(db *dbpg.DB, cache cache.Cache, logger logger.Logger, logFile *os.File, config config.Config) (*App, error) {

	ctx, cancel := newContext(logger)
	storge := repository.NewStorage(logger, config.Storage, db)
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

// newContext creates a context that is cancelled when the process
// receives SIGINT or SIGTERM. It also logs receipt of the signal
// and initiates graceful shutdown by calling the cancel function.
func newContext(logger logger.Logger) (context.Context, context.CancelFunc) {

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sig := <-sigCh
		sigString := sig.String()
		if sig == syscall.SIGTERM {
			sigString = "terminate" // sig.String() returns the SIGTERM string in past tense for some reason
		}
		logger.LogInfo("app — received signal "+sigString+", initiating graceful shutdown", "layer", "app")
		cancel()
	}()

	return ctx, cancel

}

// Run starts the server and broker consumers in background goroutines and blocks
// until the application's context is cancelled. After cancellation it invokes Stop.
func (a *App) Run() {

	var wg sync.WaitGroup

	wg.Go(func() {
		if err := a.server.Run(); err != nil {
			a.logger.LogFatal("server run failed", err, "layer", "app")
		}
	})

	wg.Go(func() {
		if err := a.broker.Consume(); err != nil {
			a.logger.LogFatal("consumer run failed", err, "layer", "app")
		}
	})

	<-a.ctx.Done()

	a.Stop(&wg)

}

// Stop performs an orderly shutdown of application components: it shuts down
// the server and broker, waits for background work to finish, closes cache and
// storage, and closes the log file if it is not os.Stdout.
func (a *App) Stop(wg *sync.WaitGroup) {

	a.server.Shutdown()
	a.broker.Shutdown()

	wg.Wait()

	a.cache.Close()
	a.storage.Close()

	if a.logFile != nil && a.logFile != os.Stdout {
		_ = a.logFile.Close()
	}

}
