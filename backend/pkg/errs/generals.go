package errs

import "errors"

var (
	ErrInternalServer        = errors.New("internal server error")
	ErrInvalidCode           = errors.New("invalid code")
	ErrInvalidExternalCode   = errors.New("invalid external code")
	ErrEmptyMsg              = errors.New("empty msg for sending")
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
	ErrInvalidLenIDs         = errors.New("len of ids must be more than 1 and less than 200")
	ErrInvalidDays           = errors.New("difference between from to must be more than 1 and less than 14 days")
	ErrInvalidVisit          = errors.New("invalid time for visit")
	ErrCoachOnVacation       = errors.New("coach on vacation")
)
