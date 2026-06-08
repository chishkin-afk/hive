package errs

import "errors"

var (
	ErrInternalServer        = errors.New("internal server error")
	ErrInvalidCode           = errors.New("invalid code")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidToken          = errors.New("invalid token")
	ErrUserAlreadyRegistered = errors.New("user already registered")
	ErrNotEnoughRights       = errors.New("not enough rights")
	ErrInvalidPage           = errors.New("invalid page")
	ErrInvalidLimit          = errors.New("invalid limit")
)
