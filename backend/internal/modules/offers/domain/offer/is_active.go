package offer

import "errors"

var (
	ErrOfferAlreadyDisabled = errors.New("offer is already disabled")
	ErrOfferAlreadyEnabled  = errors.New("offer is already enabled")
)

type IsActive bool

func (a IsActive) Enable() (IsActive, error) {
	if a {
		return a, ErrOfferAlreadyEnabled
	}

	return IsActive(true), nil
}

func (a IsActive) Disable() (IsActive, error) {
	if !a {
		return a, ErrOfferAlreadyDisabled
	}

	return IsActive(false), nil
}

func (a IsActive) Bool() bool {
	return bool(a)
}
