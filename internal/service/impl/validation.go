package impl

import (
	"Chronos/internal/errs"
	"Chronos/internal/models"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

func validateCreate(notification *models.Notification) error {

	if err := validateChannel(notification.Channel); err != nil {
		return err
	}

	if err := validateMessage(&notification.Message); err != nil {
		return err
	}

	if err := validateSendAt(notification.SendAt); err != nil {
		return err
	}

	if notification.Channel == models.Email {
		if err := validateEmails(notification.SendTo, notification.Subject); err != nil {
			return err
		}
	}

	return nil

}

func validateChannel(channel string) error {

	if channel == "" {
		return errs.ErrMissingChannel
	}

	ch := strings.ToLower(channel)

	if ch != models.Telegram && ch != models.Email && ch != models.Stdout {
		return errs.ErrUnsupportedChannel
	}

	return nil

}

func validateMessage(message *string) error {

	if utf8.RuneCountInString(*message) > models.MaxMessageLength {
		return errs.ErrMessageTooLong
	}

	if len(*message) == 0 {
		*message = "ㅤ"
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
func validateEmails(recipients []string, subject string) error {

	if len(recipients) == 0 {
		return errs.ErrMissingSendTo
	}

	if subject == "" {
		return errs.ErrMissingEmailSubject
	}

	if utf8.RuneCountInString(subject) > models.MaxSubjectLength {
		return errs.ErrEmailSubjectTooLong
	}

	for _, recipient := range recipients {

		if recipient == "" {
			return errs.ErrInvalidEmailFormat
		}

		if len(recipient) > models.MaxEmailLength {
			return errs.ErrRecipientTooLong
		}

		addr, err := mail.ParseAddress(recipient)
		if err != nil {
			return errs.ErrInvalidEmailFormat
		}

		parts := strings.Split(addr.Address, "@")
		if len(parts) != 2 || !strings.Contains(parts[1], ".") {
			return errs.ErrInvalidEmailFormat
		}

	}

	return nil

}
