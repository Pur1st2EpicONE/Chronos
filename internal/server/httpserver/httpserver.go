// Package httpserver provides a wrapper around the standard net/http server
// with logging and graceful shutdown support.
package httpserver

import (
	"Chronos/internal/config"
	"Chronos/internal/logger"
	"context"
	"errors"
	"net/http"
	"time"
)

// HttpServer wraps http.Server and adds logging and graceful shutdown.
type HttpServer struct {
	srv             *http.Server  // underlying HTTP server instance
	shutdownTimeout time.Duration // timeout duration for graceful shutdown
	logger          logger.Logger // logger used for info and error messages
}

// NewServer creates a new HttpServer with the provided configuration, logger, and handler.
// The shutdownTimeout from config is used for graceful server shutdown.
func NewServer(logger logger.Logger, config config.Server, handler http.Handler) *HttpServer {

	server := &HttpServer{
		shutdownTimeout: config.ShutdownTimeout,
		logger:          logger,
	}

	server.srv = &http.Server{
		Addr:           ":" + config.Port,
		Handler:        handler,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}

	return server

}

// Run starts the HTTP server and begins handling incoming requests.
// Logs the start, and returns any unexpected error (except ErrServerClosed).
func (s *HttpServer) Run() error {
	s.logger.LogInfo("server — receiving requests", "layer", "server")
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully stops the HTTP server using the configured shutdown timeout.
// Logs success or failure of the shutdown.
func (s *HttpServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.LogError("server — failed to shutdown gracefully", err, "layer", "server")
	} else {
		s.logger.LogInfo("server — shutdown complete", "layer", "server")
	}
}
