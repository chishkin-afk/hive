package user

import (
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrInvalidEmail = errors.New("email is invalid")
)

type Email string

func (e Email) Validate() error {
	if _, err := mail.ParseAddress(e.String()); err != nil {
		return ErrInvalidEmail
	}

	return nil
}

func (e Email) Norm() Email {
	trimmed := strings.TrimSpace(e.String())
	return Email(strings.ToLower(trimmed))
}

func (e Email) String() string {
	return string(e)
}
