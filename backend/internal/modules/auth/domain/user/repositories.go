package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type UserPersistenceRepository interface {
	Save(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email Email) (*User, error)
	Update(ctx context.Context, updUser *User) (*User, error)
}

type UserCacheRepository interface {
	Set(ctx context.Context, user *User) error
	Get(ctx context.Context, id uuid.UUID) (*User, error)
	Del(ctx context.Context, id uuid.UUID) error
	Prepare(ctx context.Context, user *User, generatedPassword string, code int) error
	GetPrepared(ctx context.Context, code int, email Email) (*User, string, int, error)
}

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)
