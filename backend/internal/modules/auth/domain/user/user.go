package user

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	maxNameLen = 32
	minNameLen = 3
)

// User is a root aggregate of auth module
type User struct {
	id             uuid.UUID
	firstname      string
	lastname       string // can be empty
	confirmedEmail Email
	passwordHash   PasswordHash
	paidUntil      time.Time // by default: now
	onVacation     bool
	tgChatID       int64 // by default 0
	notification   Notification
	createdAt      time.Time
	updatedAt      time.Time
}

func New(
	firstname, lastname string,
	confirmedEmail Email,
	password string,
) (*User, error) {
	firstname = strings.TrimSpace(firstname)
	if !isValidName(firstname) {
		return nil, &InvalidFirstnameError{
			firstname: firstname,
		}
	}

	lastname = strings.TrimSpace(lastname)
	if !isValidName(lastname) {
		return nil, &InvalidLastnameError{
			lastname: lastname,
		}
	}

	if err := confirmedEmail.Validate(); err != nil {
		return nil, err
	}

	passwordHash, err := NewPasswordHash(password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &User{
		id:             uuid.New(),
		firstname:      firstname,
		lastname:       lastname,
		confirmedEmail: confirmedEmail,
		passwordHash:   passwordHash,
		paidUntil:      now,
		onVacation:     false,
		tgChatID:       0,
		notification:   NewNotification(false),
		createdAt:      now,
		updatedAt:      now,
	}, nil
}

func From(
	id uuid.UUID,
	firstname, lastname string,
	confirmedEmail Email,
	passwordHash PasswordHash,
	paidUntil time.Time,
	onVacation bool,
	tgChatID int64,
	notification Notification,
	createdAt, updatedAt time.Time,
) *User {
	return &User{
		id:             id,
		firstname:      firstname,
		lastname:       lastname,
		confirmedEmail: confirmedEmail,
		passwordHash:   passwordHash,
		paidUntil:      paidUntil,
		onVacation:     onVacation,
		tgChatID:       tgChatID,
		notification:   notification,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

// ChangeFirstname is a business-logic method
// to change the firstname of user
func (u *User) ChangeFirstname(newFirstname string) error {
	newFirstname = strings.TrimSpace(newFirstname)
	if !isValidName(newFirstname) {
		return &InvalidFirstnameError{
			firstname: newFirstname,
		}
	}

	u.firstname = newFirstname
	u.updatedAt = time.Now().UTC()

	return nil
}

// ChangeLastname is a business-logic method
// to change the lastname of user (can be empty)
func (u *User) ChangeLastname(newLastname string) error {
	newLastname = strings.TrimSpace(newLastname)
	if len([]rune(newLastname)) != 0 && !isValidName(newLastname) {
		return &InvalidLastnameError{
			lastname: newLastname,
		}
	}

	u.lastname = newLastname
	u.updatedAt = time.Now().UTC()

	return nil
}

// ChangePassword is a business-logic method
// to create the new hash of password and set
func (u *User) ChangePassword(password string) error {
	hash, err := NewPasswordHash(password)
	if err != nil {
		return err
	}

	u.passwordHash = hash
	u.updatedAt = time.Now().UTC()

	return nil
}

// OnRest is a business-logic method
// to set onVacation = true field of user
func (u *User) OnRest() error {
	if u.onVacation {
		return ErrAlreadyOnRest
	}

	u.onVacation = true
	u.updatedAt = time.Now().UTC()

	return nil
}

// OnRest is a business-logic method
// to set onVacation = false field of user
func (u *User) OffRest() error {
	if !u.onVacation {
		return ErrAlreadyOffRest
	}

	u.onVacation = false
	u.updatedAt = time.Now().UTC()

	return nil
}

// ChangeTgChatID is a business-logic method
// to change telegram chat's id of user (can be empty)
func (u *User) ChangeTgChatID(newID int64) error {
	if newID < 0 {
		return &InvalidTgChatID{
			tgChatID: newID,
		}
	}

	u.tgChatID = newID
	u.updatedAt = time.Now().UTC()

	return nil
}

// ChangeNotification is a business-logic method
// to change notification's field of user
func (u *User) ChangeNotification(newNotification Notification) error {
	if newNotification.IsTelegramEnable() && u.tgChatID == 0 {
		return ErrNotificationTgDisabled
	}

	u.notification = newNotification
	u.updatedAt = time.Now().UTC()

	return nil
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Firstname() string {
	return u.firstname
}

func (u *User) Lastname() string {
	return u.lastname
}

func (u *User) ConfirmedEmail() Email {
	return u.confirmedEmail
}

func (u *User) PasswordHash() PasswordHash {
	return u.passwordHash
}

func (u *User) PaidUntil() time.Time {
	return u.paidUntil
}

func (u *User) OnVacation() bool {
	return u.onVacation
}

func (u *User) TgChatID() int64 {
	return u.tgChatID
}

func (u *User) Notification() Notification {
	return u.notification
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func isValidName(name string) bool {
	fnLen := len([]rune(name))
	if fnLen < minNameLen || fnLen > maxNameLen {
		return false
	}

	return true
}
