package payment

import (
	"context"

	"github.com/google/uuid"
)

type UpdateFunc func(ctx context.Context, payment *Payment) error

type PaymentPersistenceRepository interface {
	Save(ctx context.Context, payment *Payment) (*Payment, error)
	GetByID(ctx context.Context, paymentID uuid.UUID) (*Payment, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*Payment, error)
	Update(ctx context.Context, paymentID uuid.UUID, updFunc UpdateFunc) (*Payment, error)
}
