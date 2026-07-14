package tgbinding

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type TgBindingCacheRepository interface {
	Set(ctx context.Context, tgBinding *TgBinding) error
	Get(ctx context.Context, id uuid.UUID) (*TgBinding, error)
	Del(ctx context.Context, id uuid.UUID) error
}

var (
	ErrTgBindingNotFound = errors.New("tg binding's not found")
)
