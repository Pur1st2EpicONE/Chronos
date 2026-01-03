// Package server provides an abstraction for running and managing the HTTP server.
package server

import (
	"Chronos/internal/config"
	"Chronos/internal/logger"
	"Chronos/internal/server/httpserver"
	"net/http"
)

// Server defines the interface for running and shutting down the HTTP server.
type Server interface {
	Run() error // Run starts the HTTP server and begins handling incoming requests.
	Shutdown()  // Shutdown gracefully stops the HTTP server.
}

// NewServer creates a new HTTP server using the provided logger, configuration, and handler.
func NewServer(logger logger.Logger, config config.Server, handler http.Handler) Server {
	return httpserver.NewServer(logger, config, handler)
}
