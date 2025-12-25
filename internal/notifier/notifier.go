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

type Notifier interface {
	Notify(notification models.Notification) error
}

func NewNotifier(config config.Notifier) Notifier {
	return newSender(config)
}

type Sender struct {
	telegramToken    string
	telegramReceiver string
	emailSender      string
	emailPassword    string
	emailSMTP        string
	emailSMTPAddr    string
}

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

func (s *Sender) sendEmail(sendTo []string, subject string, body string) error {
	auth := smtp.PlainAuth("", s.emailSender, s.emailPassword, s.emailSMTP)
	message := []byte("Subject: " + subject + "\n" + body)
	if err := smtp.SendMail(s.emailSMTPAddr, auth, s.emailSender, sendTo, message); err != nil {
		return fmt.Errorf("failed to send email to %v via SMTP server %s: %w", sendTo, s.emailSMTPAddr, err)
	}
	return nil
}

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
