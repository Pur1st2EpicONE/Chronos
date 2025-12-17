package v1

import (
	"Chronos/internal/errs"
	"errors"
	"net/http"
	"time"

	"github.com/wb-go/wbf/ginext"
)

func parseTime(timeStr string) (time.Time, error) {

	if timeStr == "" {
		return time.Time{}, errs.ErrMissingSendAt
	}

	validTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}, errs.ErrInvalidSendAt
	}

	return validTime.UTC(), nil

}

func respondOK(c *ginext.Context, response any) {
	c.JSON(http.StatusOK, ginext.H{"result": response})
}

func respondError(c *ginext.Context, err error) {
	if err != nil {
		status, msg := mapErrorToStatus(err)
		c.AbortWithStatusJSON(status, ginext.H{"error": msg})
	}
}

func mapErrorToStatus(err error) (int, string) {

	switch {

	case errors.Is(err, errs.ErrInvalidJSON),
		errors.Is(err, errs.ErrInvalidNotificationID),
		errors.Is(err, errs.ErrMissingChannel),
		errors.Is(err, errs.ErrUnsupportedChannel),
		errors.Is(err, errs.ErrMessageTooLong),
		errors.Is(err, errs.ErrMissingSendAt),
		errors.Is(err, errs.ErrInvalidSendAt),
		errors.Is(err, errs.ErrSendAtInPast),
		errors.Is(err, errs.ErrSendAtTooFar),
		errors.Is(err, errs.ErrMissingSendTo),
		errors.Is(err, errs.ErrInvalidEmailFormat),
		errors.Is(err, errs.ErrCannotCancel),
		errors.Is(err, errs.ErrAlreadyCanceled),
		errors.Is(err, errs.ErrRecipientTooLong):
		return http.StatusBadRequest, err.Error()

	case errors.Is(err, errs.ErrNotificationNotFound):
		return http.StatusNotFound, err.Error()

	default:
		if errors.Is(err, errs.ErrUrgentDeliveryFailed) {
			return http.StatusInternalServerError, err.Error()
		}
		return http.StatusInternalServerError, errs.ErrInternal.Error()
	}

}
