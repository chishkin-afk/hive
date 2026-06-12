package user

import (
	"errors"

	"github.com/dongri/phonenumber"
)

var (
	ErrInvalidPhonenumber = errors.New("invalid phonenumber")
)

type PhoneNumber string

func (pn PhoneNumber) Standart() (PhoneNumber, error) {
	if pn == "" {
		return "", nil
	}

	normalized := phonenumber.Parse(string(pn), "RU")
	if normalized == "" {
		return "", ErrInvalidPhonenumber
	}

	return PhoneNumber(normalized), nil
}

func (pn PhoneNumber) String() string {
	return string(pn)
}
