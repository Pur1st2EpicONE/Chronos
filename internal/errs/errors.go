package errs

import "errors"

var (
	ErrInvalidJSON           = errors.New("invalid JSON format")                                               // invalid JSON format
	ErrInvalidNotificationID = errors.New("missing or invalid notification ID")                                // invalid notification ID
	ErrMissingChannel        = errors.New("channel is required")                                               // channel is required
	ErrUnsupportedChannel    = errors.New("unsupported channel")                                               // unsupported channel
	ErrMessageTooLong        = errors.New("message exceeds maximum length")                                    // message exceeds maximum length
	ErrMissingSendAt         = errors.New("send_at is required")                                               // send_at is required
	ErrInvalidSendAt         = errors.New("invalid send_at format, expected RFC3339 or 'YYYY-MM-DD HH:MM:SS'") // invalid send_at format, expected RFC3339 or 'YYYY-MM-DD HH:MM:SS'
	ErrSendAtInPast          = errors.New("send_at cannot be in the past")                                     // send_at cannot be in the past
	ErrSendAtTooFar          = errors.New("send_at is too far in the future")                                  // send_at is too far in the future
	ErrMissingSendTo         = errors.New("send_to is required")                                               // send_to is required
	ErrInvalidEmailFormat    = errors.New("invalid email format")                                              // invalid email format
	ErrRecipientTooLong      = errors.New("recipient exceeds maximum length")                                  // recipient exceeds maximum length
	ErrNotificationNotFound  = errors.New("notification not found")                                            // notification not found
	ErrInternal              = errors.New("internal server error")                                             // internal server error
)
