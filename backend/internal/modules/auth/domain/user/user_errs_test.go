package user

import (
	"errors"
	"testing"
)

func TestInvalidEmailError(t *testing.T) {
	origErr := errors.New("parse error")
	err := &InvalidEmailError{
		email:    "bad@email",
		original: origErr,
	}

	if got := err.Error(); got != "invalid email: bad@email" {
		t.Errorf("InvalidEmailError.Error() = %v, want %v", got, "invalid email: bad@email")
	}

	if unwrapped := err.Unwrap(); unwrapped != origErr {
		t.Errorf("InvalidEmailError.Unwrap() = %v, want %v", unwrapped, origErr)
	}
}

func TestInvalidFirstnameError(t *testing.T) {
	err := &InvalidFirstnameError{
		firstname: "Jo",
	}
	expected := "len of firstname must be more than 3 and less than 32"
	if got := err.Error(); got != expected {
		t.Errorf("InvalidFirstnameError.Error() = %v, want %v", got, expected)
	}
}

func TestInvalidLastnameError(t *testing.T) {
	err := &InvalidLastnameError{
		lastname: "A",
	}
	expected := "len of lastname must be more than 3 and less than 32"
	if got := err.Error(); got != expected {
		t.Errorf("InvalidLastnameError.Error() = %v, want %v", got, expected)
	}
}

func TestInvalidTgChatID(t *testing.T) {
	err := &InvalidTgChatID{
		tgChatID: -123,
	}
	expected := "invalid tg chat's id: -123"
	if got := err.Error(); got != expected {
		t.Errorf("InvalidTgChatID.Error() = %v, want %v", got, expected)
	}
}

func TestDomainErrors(t *testing.T) {
	if ErrAlreadyOnRest == nil {
		t.Error("ErrAlreadyOnRest should not be nil")
	}
	if ErrAlreadyOffRest == nil {
		t.Error("ErrAlreadyOffRest should not be nil")
	}
	if ErrNotificationTgDisabled == nil {
		t.Error("ErrNotificationTgDisabled should not be nil")
	}
}