// Package v1 provides version 1 of the Chronos API handlers for notifications.
// It includes endpoints to create, query, and cancel notifications via HTTP.
package v1

import (
	"Chronos/internal/errs"
	"Chronos/internal/models"
	"Chronos/internal/service"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/helpers"
)

// Handler is the v1 API handler for notifications.
// It wraps the service layer and provides HTTP endpoints for CRUD operations.
type Handler struct {
	service service.Service
}

// NewHandler creates a new v1 Handler with the provided service.
func NewHandler(service service.Service) *Handler {
	return &Handler{service: service}
}

// CreateNotification handles POST /notify requests.
// It parses the JSON body, validates the input, creates a new notification via the service,
// and returns the generated notification ID. Errors are returned for invalid input or service failures.
func (h *Handler) CreateNotification(c *ginext.Context) {

	var request CreateNotificationV1

	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, errs.ErrInvalidJSON)
		return
	}

	sendAt, err := parseTime(request.SendAt)
	if err != nil {
		respondError(c, err)
		return
	}

	notification := models.Notification{
		Channel: request.Channel,
		Subject: request.Subject,
		Message: request.Message,
		SendAt:  sendAt,
		SendTo:  request.SendTo,
	}

	id, err := h.service.CreateNotification(c.Request.Context(), notification)
	if err != nil {
		respondError(c, err)
		return
	}

	respondOK(c, id)

}

// GetNotification handles GET /notify?id=<id> requests.
// It validates the notification ID and returns the current status of the notification.
// Returns an error if the ID is invalid or the notification is not found.
func (h *Handler) GetNotification(c *ginext.Context) {

	notificationID := c.Query("id")
	if err := helpers.ParseUUID(notificationID); err != nil {
		respondError(c, errs.ErrInvalidNotificationID)
		return
	}

	status, err := h.service.GetStatus(c.Request.Context(), notificationID)
	if err != nil {
		respondError(c, err)
		return
	}

	respondOK(c, status)

}

// CancelNotification handles DELETE /notify?id=<id> requests.
// It validates the notification ID and cancels the notification if possible.
// Returns an error if the ID is invalid or the notification cannot be canceled.
func (h *Handler) CancelNotification(c *ginext.Context) {

	notificationID := c.Query("id")
	if err := helpers.ParseUUID(notificationID); err != nil {
		respondError(c, errs.ErrInvalidNotificationID)
		return
	}

	if err := h.service.CancelNotification(c.Request.Context(), notificationID); err != nil {
		respondError(c, err)
		return
	}

	respondOK(c, "canceled")

}
