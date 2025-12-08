package v1

import (
	"Chronos/internal/models"
	"Chronos/internal/service"

	"github.com/wb-go/wbf/ginext"
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
		respondError(c, err)
		return
	}

	sendAt, err := parseTime(request.SendAt)
	if err != nil {
		respondError(c, err)
		return
	}

	notification := models.Notification{
		Channel: request.Channel,
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
