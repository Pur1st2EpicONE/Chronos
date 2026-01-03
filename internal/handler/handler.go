// Package handler provides HTTP handlers for the Chronos application.
// It sets up routes, static file serving, and HTML templates for the web interface.
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

// NewHandler creates and returns an http.Handler configured with all routes, middleware, and template rendering.
// It includes API v1 routes for notifications and a web frontend at the root path.
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

// homePage returns a handler function that renders the HTML home page for the web frontend.
// It retrieves all notification statuses from the service and injects them into the template.
// Note: The service is primarily designed for API usage, so loading all notifications from the database
// is not optimized and is included here only to make the UI more illustrative.
func homePage(tmpl *template.Template, service service.Service) func(c *ginext.Context) {
	return func(c *ginext.Context) {
		notifications := service.GetAllStatuses(c.Request.Context())
		c.Header("Content-Type", "text/html")
		if err := tmpl.Execute(c.Writer, map[string]any{"Notifications": notifications}); err != nil {
			c.String(http.StatusInternalServerError, errs.ErrInternal.Error())
		}
	}
}
