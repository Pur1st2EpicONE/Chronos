package models

import "time"

type Notification struct {
	ID        string    `json:"id"`
	Channel   string    `json:"channel"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	SendAt    time.Time `json:"send_at"`
	SendTo    string    `json:"send_to"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	StatusPending            = "pending"
	StatusCanceled           = "canceled"
	StatusFailedToSendInTime = "failed to send in time"
	StatusFailed             = "failed to send"
	StatusLate               = "running late"
	StatusSent               = "sent"
)
