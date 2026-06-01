package user

import (
	"fmt"
	"net/mail"
	"strings"
)

// InvalidEmailError is a safe error
// that can be returned to user
type InvalidEmailError struct {
	email    string
	original error
}

func (iee *InvalidEmailError) Error() string {
	return fmt.Sprintf("invalid email: %s", iee.email)
}

func (iee *InvalidEmailError) Unwrap() error {
	return iee.original
}

// Email is a value object of User model
type Email string

func (e Email) String() string {
	return string(e)
}

func (e Email) Validate() error {
	if _, err := mail.ParseAddress(e.String()); err != nil {
		return &InvalidEmailError{
			email:    e.String(),
			original: err,
		}
	}

	return nil
}

func (e Email) Norm() Email {
	normalized := strings.TrimSpace(e.String())
	return Email(strings.ToLower(normalized))
}
