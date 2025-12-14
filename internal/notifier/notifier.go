package notifier

import (
	"Chronos/internal/config"
	"Chronos/internal/models"
	"fmt"
	"net/http"
	"net/url"
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
}

func newSender(config config.Notifier) *Sender {
	return &Sender{
		telegramToken:    config.TelegramToken,
		telegramReceiver: config.TelegramReceiver,
	}
}

func (s *Sender) Notify(notification models.Notification) error {

	switch notification.Channel {
	case "telegram":
		if err := s.sendTelegram(notification.Message); err != nil {
			return err
		}
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
		fmt.Println("err post form", err)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("err status code", err)
		return fmt.Errorf("telegram API returned status %s", resp.Status)
	}

	return nil

}
