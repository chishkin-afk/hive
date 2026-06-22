package payment

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type UpdateFunc func(ctx context.Context, paidUntil time.Time, payment *Payment) (time.Time, error)

type PaymentPersistenceRepository interface {
	Save(ctx context.Context, payment *Payment) (*Payment, error)
	GetByID(ctx context.Context, paymentID uuid.UUID) (*Payment, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*Payment, int64, error)
	Update(ctx context.Context, paymentID, userID uuid.UUID, updFunc UpdateFunc) (*Payment, error)
}

var (
	ErrPaymentNotFound     = errors.New("payment not found")
	ErrPaymentNotSucceeded = errors.New("payment isn't succeeded")
	ErrUserNotFound        = errors.New("user not found")
)
