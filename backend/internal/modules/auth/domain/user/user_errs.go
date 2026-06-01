package user

import (
	"errors"
	"fmt"
)

// InvalidFirstnameError is a safe error
// that can be returned to client
type InvalidFirstnameError struct {
	firstname string
}

func (ife *InvalidFirstnameError) Error() string {
	return fmt.Sprintf("len of firstname must be more than %d and less than %d", minNameLen, maxNameLen)
}

// InvalidLastnameError is a safe error
// that can be returned to client
type InvalidLastnameError struct {
	lastname string
}

func (ile *InvalidLastnameError) Error() string {
	return fmt.Sprintf("len of lastname must be more than %d and less than %d", minNameLen, maxNameLen)
}

// InvalidTgChatID is a safe error
// that can be returned to client
type InvalidTgChatID struct {
	tgChatID int64
}

func (ite *InvalidTgChatID) Error() string {
	return fmt.Sprintf("invalid tg chat's id: %d", ite.tgChatID)
}

var (
	ErrAlreadyOnRest          = errors.New("user already is on rest")
	ErrAlreadyOffRest         = errors.New("user already isn't on rest")
	ErrNotificationTgDisabled = errors.New("tg chat's id is empty, you can't enable telegram")
)
