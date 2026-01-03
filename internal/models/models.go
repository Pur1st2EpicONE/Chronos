// Package models defines the core data structures used in Chronos,
// including notifications and their associated constants.
package models

import "time"

// Notification represents a scheduled notification.
// It contains information about the delivery channel, message content,
// recipients, status, and timestamps.
type Notification struct {
	ID          string    `json:"id"`            // Unique identifier for the notification
	Channel     string    `json:"channel"`       // Delivery channel
	Subject     string    `json:"subject"`       // Subject or title of the notification
	Message     string    `json:"message"`       // Main content of the notification
	Status      string    `json:"status"`        // Current status of the notification
	SendAt      time.Time `json:"send_at"`       // Scheduled UTC time for sending
	SendAtLocal string    `json:"send_at_local"` // Scheduled time in local timezone
	SendTo      []string  `json:"send_to"`       // List of recipients
	UpdatedAt   time.Time `json:"updated_at"`    // Last update timestamp
}

const (
	StatusPending            = "pending"                // Notification is created but not yet sent
	StatusCanceled           = "canceled"               // Notification has been canceled
	StatusFailedToSendInTime = "failed to send in time" // Notification failed to send before scheduled time
	StatusFailed             = "failed to send"         // Notification failed to send due to error
	StatusLate               = "running late"           // Notification delayed past its scheduled send time
	StatusSent               = "sent"                   // Notification was successfully sent
)

const (
	Email    = "email"    // Email channel
	Stdout   = "stdout"   // Standard output
	Telegram = "telegram" // Telegram bot channel
)

const (
	MaxEmailLength   = 254 // Maximum length for email addresses
	MaxSubjectLength = 254 // Maximum length for email subject
	MaxMessageLength = 254 // Maximum length for message content
)
