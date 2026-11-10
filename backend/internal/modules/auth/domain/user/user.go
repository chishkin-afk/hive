package user

import (
	"errors"
	"strings"
	"time"
	"uuid"

	"github.com/chishkin-afk/hive/backend/internal/modules/auth/domain/notification"
)

var (
	ErrInvalidFirstname = errors.New("len of firstname must be more than 2 and less than 64")
	ErrInvalidLastname  = errors.New("len of lastname must be more than 2 and less than 64")
	ErrInvalidSurname   = errors.New("len of surname must be more than 2 and less than 64")
)

type User struct {
	id           uuid.UUID
	notification *notification.Notification
	email        Email
	firstname    string
	lastname     string
	surname      string
	createdAt    time.Time
	updatedAt    time.Time
}

func New(
	email Email,
	firstname string,
	lastname string,
	surname string,
) (*User, error) {
	email = email.Norm()
	if err := email.Validate(); err != nil {
		return nil, err
	}

	firstname = strings.TrimSpace(firstname)
	if !checkLen(firstname, 2, 64) {
		return nil, ErrInvalidFirstname
	}

	lastname = strings.TrimSpace(lastname)
	if !checkLen(lastname, 2, 64) {
		return nil, ErrInvalidLastname
	}

	surname = strings.TrimSpace(surname)
	if surname != "" && !checkLen(surname, 2, 64) {
		return nil, ErrInvalidSurname
	}

	now := time.Now().UTC()
	return &User{
		id:           uuid.New(),
		notification: notification.New(),
		email:        email,
		firstname:    firstname,
		lastname:     lastname,
		surname:      surname,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func (u *User) ChangeFirstname(firstname string) error {
	firstname = strings.TrimSpace(firstname)
	if !checkLen(firstname, 2, 64) {
		return ErrInvalidFirstname
	}

	u.firstname = firstname
	u.updatedAt = time.Now().UTC()

	return nil
}

func (u *User) ChangeLastname(lastname string) error {
	lastname = strings.TrimSpace(lastname)
	if !checkLen(lastname, 2, 64) {
		return ErrInvalidLastname
	}

	u.lastname = lastname
	u.updatedAt = time.Now().UTC()

	return nil
}

func (u *User) ChangeSurname(surname string) error {
	surname = strings.TrimSpace(surname)
	if surname != "" && !checkLen(surname, 2, 64) {
		return ErrInvalidSurname
	}

	u.surname = surname
	u.updatedAt = time.Now().UTC()

	return nil
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Notification() *notification.Notification {
	return u.notification
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) Firstname() string {
	return u.firstname
}

func (u *User) Lastname() string {
	return u.lastname
}

func (u *User) Surname() string {
	return u.surname
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func checkLen(s string, min, max int) bool {
	n := len([]rune(s))
	return n < max && n >= min
}
