package errs

import "errors"

var (
	ErrInternalServer        = errors.New("internal server error")
	ErrInvalidCode           = errors.New("invalid code")
	ErrInvalidExternalCode   = errors.New("invalid external code")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidToken          = errors.New("invalid token")
	ErrNotEnoughRights       = errors.New("not enough rights")
	ErrInvalidPage           = errors.New("invalid page")
	ErrInvalidLimit          = errors.New("invalid limit")
	ErrUserNotPaid           = errors.New("user not paid")
	ErrInvalidTimeForBooking = errors.New("invalid time for booking")
	ErrInvalidAmount         = errors.New("invalid amount of payment")
	ErrInvalidPaymentStatus  = errors.New("invalid status of payment")
	ErrInvalidFingerprint    = errors.New("invalid fingerprint")
	ErrTooMuchRegistrations  = errors.New("too much registrations")
	ErrTooMuchResents        = errors.New("too much resents")
)
