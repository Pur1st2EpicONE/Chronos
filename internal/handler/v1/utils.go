package v1

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/wb-go/wbf/ginext"
)

// func parseQuery(userID string, eventDate string) (int, time.Time, error) {

// 	if userID == "" || eventDate == "" {
// 		return 0, time.Time{}, errs.ErrMissingParams
// 	}

// 	id, err := strconv.Atoi(userID)
// 	if err != nil {
// 		return 0, time.Time{}, errs.ErrInvalidUserID
// 	}

// 	date, err := parseDate(eventDate)
// 	if err != nil {
// 		return 0, time.Time{}, err
// 	}

// 	return id, date, nil

// }

func parseTime(timeStr string) (time.Time, error) {

	if timeStr == "" {
		return time.Time{}, errors.New("empty time")
	}

	if validTime, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return validTime, nil
	}

	if validTime, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
		return validTime, nil
	}

	return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)

}

func respondOK(c *ginext.Context, response any) {
	c.JSON(http.StatusOK, ginext.H{"result": response})
}

func respondError(c *ginext.Context, err error) {
	if err != nil {
		// status, msg := mapErrorToStatus(err)
		c.AbortWithStatusJSON(400, ginext.H{"error": err.Error()})
	}
}

// func mapErrorToStatus(err error) (int, string) {

// 	switch {

// 	case errors.Is(err, errs.ErrInvalidJSON),
// 		errors.Is(err, errs.ErrInvalidUserID),
// 		errors.Is(err, errs.ErrInvalidEventID),
// 		errors.Is(err, errs.ErrInvalidDateFormat),
// 		errors.Is(err, errs.ErrEmptyEventText),
// 		errors.Is(err, errs.ErrEventTextTooLong),
// 		errors.Is(err, errs.ErrMissingEventID),
// 		errors.Is(err, errs.ErrMissingParams),
// 		errors.Is(err, errs.ErrMissingDate):
// 		return http.StatusBadRequest, err.Error()

// 	case errors.Is(err, errs.ErrMaxEvents),
// 		errors.Is(err, errs.ErrEventNotFound),
// 		errors.Is(err, errs.ErrNothingToUpdate),
// 		errors.Is(err, errs.ErrEventInPast),
// 		errors.Is(err, errs.ErrEventTooFar),
// 		errors.Is(err, errs.ErrUnauthorized):
// 		return http.StatusServiceUnavailable, err.Error()

// 	default:
// 		return http.StatusInternalServerError, errs.ErrInternal.Error()

// 	}

// }
