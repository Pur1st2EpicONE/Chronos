package v1

import (
	"Chronos/internal/errs"
	"Chronos/internal/models"
	"Chronos/internal/service"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/helpers"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{service: service}
}

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
