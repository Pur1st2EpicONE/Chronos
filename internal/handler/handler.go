package handler

import (
	"Chronos/internal/errs"
	"Chronos/internal/service"
	"html/template"
	"net/http"

	v1 "Chronos/internal/handler/v1"

	"github.com/wb-go/wbf/ginext"
)

const templatePath = "web/templates/index.html"

func NewHandler(service service.Service) http.Handler {

	handler := ginext.New("")

	handler.Use(ginext.Recovery())
	handler.Static("/static", "./web/static")

	apiV1 := handler.Group("/api/v1")
	handlerV1 := v1.NewHandler(service)

	apiV1.GET("/notify", handlerV1.GetNotification)
	apiV1.POST("/notify", handlerV1.CreateNotification)
	apiV1.DELETE("/notify", handlerV1.CancelNotification)

	handler.GET("/", homePage(template.Must(template.ParseFiles(templatePath)), service))

	return handler

}

func homePage(tmpl *template.Template, service service.Service) func(c *ginext.Context) {
	return func(c *ginext.Context) {
		notifications := service.GetAllStatuses(c.Request.Context())
		c.Header("Content-Type", "text/html")
		if err := tmpl.Execute(c.Writer, map[string]any{"Notifications": notifications}); err != nil {
			c.String(http.StatusInternalServerError, errs.ErrInternal.Error())
		}
	}
}
