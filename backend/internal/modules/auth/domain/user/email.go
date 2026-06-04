package user

import (
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrInvalidEmail = errors.New("invalid email of user")
)

// Email is a value object of User model
type Email string

func (e Email) String() string {
	return string(e)
}

func (e Email) Validate() error {
	if _, err := mail.ParseAddress(e.String()); err != nil {
		return ErrInvalidEmail
	}

	return nil
}

func (e Email) Norm() Email {
	normalized := strings.TrimSpace(e.String())
	return Email(strings.ToLower(normalized))
}
