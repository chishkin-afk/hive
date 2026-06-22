package rate

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ApplyFunc func(ctx context.Context, paidUntil time.Time) (time.Time, error)

type RatePersistenceRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Rate, error)
	ListAll(ctx context.Context) ([]*Rate, error)
}

var (
	ErrRateNotFound = errors.New("rate not found")
)
