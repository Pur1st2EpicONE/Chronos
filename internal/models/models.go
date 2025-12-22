package models

import "time"

type Notification struct {
	ID        string    `json:"id"`
	Channel   string    `json:"channel"`
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	SendAt    time.Time `json:"send_at"`
	SendTo    []string  `json:"send_to"`
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

const (
	Email    = "email"
	Stdout   = "stdout"
	Telegram = "telegram"
)

const (
	MaxEmailLength   = 254
	MaxSubjectLength = 254
	MaxMessageLength = 254
)
