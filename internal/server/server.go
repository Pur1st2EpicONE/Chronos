package server

import (
	"Chronos/internal/config"
	"Chronos/internal/server/httpserver"
	"net/http"
)

type Server interface {
	Run() error
	Shutdown()
}

func NewServer(config config.Server, handler http.Handler) Server {
	return httpserver.NewServer(config, handler)
}
