package v1

type CreateNotificationV1 struct {
	Channel string `json:"channel" binding:"required"`
	Message string `json:"message"`
	SendAt  string `json:"send_at" binding:"required"`
	SendTo  string `json:"send_to" binding:"required"`
}
