package user

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword    = errors.New("len of password must be more than 6 and less than 32")
	ErrFailedGenerateHash = errors.New("failed to generate password hash")
)

// PasswordHash is a value object of User model
type PasswordHash string

func (ph PasswordHash) String() string {
	return string(ph)
}

func (ph PasswordHash) IsEqual(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(ph), []byte(password)) == nil
}

func NewPasswordHash(password string) (PasswordHash, error) {
	password = strings.TrimSpace(password)
	n := len([]rune(password))
	if n < 6 || n > 32 {
		return "", ErrInvalidPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", ErrFailedGenerateHash
	}

	return PasswordHash(hash), nil
}
