package models

import "time"

type Notification struct {
	ID        int64     `json:"id"`
	Channel   string    `json:"channel"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	SendAt    time.Time `json:"send_at"`
	SendTo    string    `json:"send_to"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
