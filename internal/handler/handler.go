package handler

import (
	"Chronos/internal/service"
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

func NewHandler(service service.Service) http.Handler {

	//log := zlog.Logger.With().Str("layer", "handler").Logger()

	handler := ginext.New("")

	return handler

}
