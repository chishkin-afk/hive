package errs

import "errors"

var (
	ErrInternalServer        = errors.New("internal server error")
	ErrInvalidCode           = errors.New("invalid code")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidToken          = errors.New("invalid token")
	ErrNotEnoughRights       = errors.New("not enough rights")
	ErrInvalidPage           = errors.New("invalid page")
	ErrInvalidLimit          = errors.New("invalid limit")
	ErrUserNotPaid           = errors.New("user not paid")
	ErrInvalidTimeForBooking = errors.New("invalid time for booking")
	ErrInvalidAmount         = errors.New("invalid amount of payment")
)
