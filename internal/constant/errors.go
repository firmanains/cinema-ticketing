package constant

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrDuplicateEmail     = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenInvalid       = errors.New("token is invalid or expired")
	ErrShowtimeNotFound   = errors.New("showtime not found")
	ErrInvalidTimeRange   = errors.New("start time must be before end time")
	ErrNoFieldsToUpdate   = errors.New("no fields to update")
)
