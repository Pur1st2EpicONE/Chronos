package impl

import (
	"Chronos/internal/errs"
	"Chronos/internal/models"
	"net/mail"
	"strings"
	"time"
)

func validateCreate(notification models.Notification) error {

	if err := validateChannel(notification.Channel); err != nil {
		return err
	}

	if err := validateMessage(notification.Message); err != nil {
		return err
	}

	if err := validateSendAt(notification.SendAt); err != nil {
		return err
	}

	if err := validateSendTo(notification.Channel, notification.SendTo); err != nil {
		return err
	}

	return nil

}

func validateChannel(channel string) error {

	if channel == "" {
		return errs.ErrMissingChannel
	}

	ch := strings.ToLower(channel)

	if ch != "telegram" && ch != "email" {
		return errs.ErrUnsupportedChannel
	}

	return nil

}

func validateMessage(msg string) error {

	if len(msg) > 500 {
		return errs.ErrMessageTooLong
	}

	return nil

}

func validateSendAt(t time.Time) error {

	if t.IsZero() {
		return errs.ErrMissingSendAt
	}

	now := time.Now().UTC()

	if t.Before(now) {
		return errs.ErrSendAtInPast
	}

	if t.After(now.AddDate(1, 0, 0)) {
		return errs.ErrSendAtTooFar
	}

	return nil

}

func validateSendTo(channel string, recipient string) error {

	if recipient == "" {
		return errs.ErrMissingSendTo
	}

	if len(recipient) > 254 {
		return errs.ErrRecipientTooLong
	}

	if channel == "email" {
		addr, err := mail.ParseAddress(recipient)
		if err != nil || !strings.Contains(strings.Split(addr.Address, "@")[1], ".") {
			return errs.ErrInvalidEmailFormat
		}
	}

	if channel == "telegram" {
		//
	}

	return nil

}
