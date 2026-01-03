package v1

// CreateNotificationV1 represents the JSON payload for creating a new notification via the v1 API.
// It is used in POST /notify requests.
type CreateNotificationV1 struct {
	Channel string   `json:"channel"` // The channel to send the notification through (e.g., "email", "telegram").
	Subject string   `json:"subject"` // The subject or title of the notification (used for email, optional for other channels).
	Message string   `json:"message"` // The main content of the notification.
	SendAt  string   `json:"send_at"` // The scheduled send time in RFC3339 format.
	SendTo  []string `json:"send_to"` // The list of recipients for the notification.
}
