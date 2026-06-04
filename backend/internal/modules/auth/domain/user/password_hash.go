package user

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// InvalidPasswordError is a safe error
// that can be returned to client
type InvalidPasswordError struct {
	msg string
}

func (ipe *InvalidPasswordError) Error() string {
	return fmt.Sprintf("invalid password: %s", ipe.msg)
}

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
		return "", &InvalidPasswordError{
			msg: "len of password must be more than 6 and less than 32",
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", &InvalidPasswordError{
			msg: "failed to generate password hash",
		}
	}

	return PasswordHash(hash), nil
}
