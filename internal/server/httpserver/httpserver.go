package httpserver

import (
	"Chronos/internal/config"
	"context"
	"net/http"
	"time"

	"github.com/wb-go/wbf/zlog"
)

type HttpServer struct {
	srv             *http.Server
	shutdownTimeout time.Duration
	log             zlog.Zerolog
}

func NewServer(config config.Server, handler http.Handler) *HttpServer {
	server := new(HttpServer)
	server.srv = &http.Server{
		Addr:           ":" + config.Port,
		Handler:        handler,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}
	server.shutdownTimeout = config.ShutdownTimeout
	server.log = zlog.Logger.With().Str("layer", "server").Logger()
	return server
}

func (s *HttpServer) Run() error {
	s.log.Info().Msg("server — receiving requests")
	return s.srv.ListenAndServe()
}

func (s *HttpServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		s.log.Err(err).Msg("server — failed to shutdown gracefully")
	} else {
		s.log.Info().Msg("server — shutdown complete")
	}
}
