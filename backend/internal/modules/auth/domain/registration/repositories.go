package registration

import (
	"context"
	"errors"
)

type RegistrationPersistenceRepository interface {
	Save(ctx context.Context, registration *Registration) (*Registration, error)
	Count(ctx context.Context, fingerprint string) (int64, error)
}

var (
	ErrNotFound     = errors.New("registrations not found")
	ErrUserNotFound = errors.New("user not found")
)
