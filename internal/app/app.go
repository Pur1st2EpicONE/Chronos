package app

import (
	"Chronos/internal/config"
	"Chronos/internal/handler"
	"Chronos/internal/repository"
	"Chronos/internal/server"
	"Chronos/internal/service"
	"os"
	"os/signal"
	"syscall"

	"context"
	"errors"
	"net/http"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type App struct {
	server  server.Server
	storage repository.Storage
	ctx     context.Context
	cancel  context.CancelFunc
	log     zlog.Zerolog
}

func Boot() *App {

	zlog.InitConsole()

	config, err := config.Load()
	if err != nil {
		zlog.Logger.Fatal().Err(err).Str("layer", "app").Msg("app — failed to load configs")
	}

	zlog.SetLevel(config.Logger.Level)
	log := zlog.Logger.With().Str("layer", "app").Logger()

	ctx, cancel := newContext(log)

	db, err := connectDB(log, config.Storage)
	if err != nil {
		log.Fatal().Err(err).Msg("app — failed to connect to database")
	}

	server, storage := wireApp(db, config)

	return &App{
		log:     log,
		ctx:     ctx,
		cancel:  cancel,
		server:  server,
		storage: storage,
	}

}

func newContext(log zlog.Zerolog) (context.Context, context.CancelFunc) {

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sig := <-sigCh
		log.Info().Msg("app — received signal " + sig.String() + ", initiating graceful shutdown")
		cancel()
	}()

	return ctx, cancel

}

func connectDB(log zlog.Zerolog, config config.Storage) (*dbpg.DB, error) {
	db, err := repository.ConnectDB(config)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("app — connected to database")
	return db, nil
}

func wireApp(db *dbpg.DB, config config.Config) (server.Server, repository.Storage) {
	storage := repository.NewStorage(db, config.Storage)
	service := service.NewService(config.Service, storage)
	handler := handler.NewHandler(service)
	server := server.NewServer(config.Server, handler)
	return server, storage
}

func (a *App) Run() {

	go func() {
		if err := a.server.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.log.Fatal().Err(err).Msg("server run failed")
		}
	}()

	<-a.ctx.Done()

	a.log.Info().Msg("app — shutting down...")
	a.Stop()

}

func (a *App) Stop() {
	a.server.Shutdown()
	a.storage.Close()
}
