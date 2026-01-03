// Package notifier provides functionality to send notifications via different channels.
// Supported channels are Email, Telegram, and Stdout.
package notifier

import (
	"Chronos/internal/config"
	"Chronos/internal/models"
	"fmt"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
)

// Notifier defines the interface for sending notifications.
type Notifier interface {
	Notify(notification models.Notification) error // Notify sends a notification using the specified channel.
}

// NewNotifier creates a new Notifier instance based on the provided configuration.
func NewNotifier(config config.Notifier) Notifier {
	return newSender(config)
}

// Sender implements the Notifier interface and sends notifications via Email, Telegram, or Stdout.
type Sender struct {
	telegramToken    string // Telegram bot token
	telegramReceiver string // Telegram chat ID to receive messages
	emailSender      string // Email address to send from
	emailPassword    string // Password or app-specific password for email account
	emailSMTP        string // SMTP server host
	emailSMTPAddr    string // SMTP server address (host:port)
}

// newSender creates a new Sender instance based on the configuration.
func newSender(config config.Notifier) *Sender {
	return &Sender{
		telegramToken:    config.TelegramToken,
		telegramReceiver: config.TelegramReceiver,
		emailSender:      config.EmailSender,
		emailPassword:    config.EmailPassword,
		emailSMTP:        config.EmailSMTP,
		emailSMTPAddr:    config.EmailSMTPAddr,
	}
}

// Notify sends the notification using the appropriate channel.
// Returns an error if sending fails or if the channel is unsupported.
func (s *Sender) Notify(notification models.Notification) error {

	switch strings.ToLower(notification.Channel) {
	case models.Telegram:
		if err := s.sendTelegram(notification.Message); err != nil {
			return fmt.Errorf("unable to send Telegram notification: %w", err)
		}
	case models.Email:
		if err := s.sendEmail(notification.SendTo, notification.Subject, notification.Message); err != nil {
			return fmt.Errorf("unable to send Email notification: %w", err)
		}
	case models.Stdout:
		if _, err := fmt.Println(notification.Message); err != nil {
			return fmt.Errorf("unable to print notification to stdout: %w", err)
		}
	default:
		return fmt.Errorf("unsupported notification channel: %s", notification.Channel)
	}

	return nil

}

// sendEmail sends an email to the specified recipients using SMTP.
func (s *Sender) sendEmail(sendTo []string, subject string, body string) error {
	auth := smtp.PlainAuth("", s.emailSender, s.emailPassword, s.emailSMTP)
	message := []byte("Subject: " + subject + "\n" + body)
	if err := smtp.SendMail(s.emailSMTPAddr, auth, s.emailSender, sendTo, message); err != nil {
		return fmt.Errorf("failed to send email to %v via SMTP server %s: %w", sendTo, s.emailSMTPAddr, err)
	}
	return nil
}

// sendTelegram sends a message via Telegram bot API.
func (s *Sender) sendTelegram(message string) error {

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.telegramToken)

	data := url.Values{}
	data.Set("chat_id", s.telegramReceiver)
	data.Set("text", message)

	client := new(http.Client)

	resp, err := client.PostForm(apiURL, data)
	if err != nil {
		return fmt.Errorf("failed to POST form to Telegram API %s: %w", apiURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned non-OK status %s for chat_id %s", resp.Status, s.telegramReceiver)
	}

	return nil

}
