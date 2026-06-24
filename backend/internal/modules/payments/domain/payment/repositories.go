package payment

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type (
	UpdateFunc    func(ctx context.Context, paidUntil time.Time, payment *Payment) (time.Time, error)
	TakeTrialFunc func(ctx context.Context, paidUntil time.Time, trialUsed bool) (time.Time, error)
)

type PaymentPersistenceRepository interface {
	Save(ctx context.Context, payment *Payment) (*Payment, error)
	GetByID(ctx context.Context, paymentID uuid.UUID) (*Payment, error)
	GetByPaymentID(ctx context.Context, paymentID uuid.UUID) (*Payment, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*Payment, int64, error)
	Update(ctx context.Context, paymentID, userID uuid.UUID, updFunc UpdateFunc) (*Payment, error)
	TakeTrial(ctx context.Context, userID uuid.UUID, updFunc TakeTrialFunc) error
}

var (
	ErrPaymentNotFound     = errors.New("payment not found")
	ErrPaymentNotSucceeded = errors.New("payment isn't succeeded")
	ErrUserNotFound        = errors.New("user not found")
	ErrTrialAlreadyUsed    = errors.New("trial is already used")
)
