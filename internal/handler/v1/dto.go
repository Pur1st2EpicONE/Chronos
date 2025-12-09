package v1

type CreateNotificationV1 struct {
	Channel string `json:"channel"`
	Message string `json:"message"`
	SendAt  string `json:"send_at"`
	SendTo  string `json:"send_to"`
}
