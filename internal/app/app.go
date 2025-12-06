package app

import (
	"Chronos/internal/config"
	"Chronos/internal/handler"
	"Chronos/internal/repository"
	"Chronos/internal/server"
	"Chronos/internal/service"

	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type App struct {
	server  server.Server
	storage repository.Storage
	ctx     context.Context
	cancel  context.CancelFunc
	wg      *sync.WaitGroup
	log     zlog.Zerolog
}

func Boot() *App {

	zlog.InitConsole()
	zlog.SetLevel("trace")
	log := zlog.Logger.With().Str("layer", "app").Logger()

	config, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("app — failed to load configs")
	}

	db, err := repository.ConnectDB(config.Storage)
	if err != nil {
		log.Fatal().Err(err).Msg("app — failed to connect to database")
	}
	log.Info().Msg("app — connected to database")

	server, storage := wireApp(db, config)

	ctx, cancel := newContext(log)
	wg := new(sync.WaitGroup)

	return &App{
		server:  server,
		storage: storage,
		ctx:     ctx,
		cancel:  cancel,
		log:     log,
		wg:      wg,
	}

}

func wireApp(db *dbpg.DB, config config.App) (server.Server, repository.Storage) {
	storage := repository.NewStorage(db, config.Storage)
	service := service.NewService(config.Service, storage)
	handler := handler.NewHandler(service)
	server := server.NewServer(config.Server, handler)
	return server, storage
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

func (a *App) Run() {

	a.wg.Go(func() {
		if err := a.server.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.log.Fatal().Err(err).Msg("server run failed")
		}
	})

	<-a.ctx.Done()

	a.log.Info().Msg("app — shutting down...")
	a.Stop()

	a.wg.Wait()

}

func (a *App) Stop() {
	a.server.Shutdown()
	a.storage.Close()
}
