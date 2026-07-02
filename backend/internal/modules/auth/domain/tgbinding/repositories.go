package tgbinding

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type TgBindingRepository interface {
	Save(ctx context.Context, tgBinding *TgBinding) error
	GetByID(ctx context.Context, id uuid.UUID) (*TgBinding, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

var (
	ErrTgBindingNotFound = errors.New("tg binding's not found")
)
