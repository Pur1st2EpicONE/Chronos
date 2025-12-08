package handler

import (
	v1 "Chronos/internal/handler/v1"
	"Chronos/internal/service"
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

func NewHandler(service service.Service) http.Handler {

	handler := ginext.New("")

	handler.Use(ginext.Recovery())

	apiV1 := handler.Group("/api/v1")
	handlerV1 := v1.NewHandler(service)

	apiV1.POST("/notify", handlerV1.CreateNotification)

	return handler

}
