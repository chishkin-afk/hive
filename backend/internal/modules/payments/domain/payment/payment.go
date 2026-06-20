package payment

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyPaymentID   = errors.New("payment's id is empty")
	ErrEmptyUserID      = errors.New("user's id is empty")
	ErrInvalidAmount    = errors.New("invalid amount")
	ErrAlreadyConfirmed = errors.New("payment is already confirmed")
)

type Payment struct {
	id          uuid.UUID
	paymentID   uuid.UUID
	userID      uuid.UUID
	amountPenny uint64
	isConfirmed bool
	createdAt   time.Time
	confirmedAt *time.Time
}

func New(
	paymentID uuid.UUID,
	userID uuid.UUID,
	amountPenny uint64,
) (*Payment, error) {
	if paymentID == uuid.Nil {
		return nil, ErrEmptyPaymentID
	}

	if userID == uuid.Nil {
		return nil, ErrEmptyUserID
	}

	if amountPenny < 1000 {
		return nil, ErrInvalidAmount
	}

	now := time.Now().UTC()
	return &Payment{
		id:          uuid.New(),
		paymentID:   paymentID,
		userID:      userID,
		amountPenny: amountPenny,
		isConfirmed: false,
		createdAt:   now,
	}, nil
}

func (p *Payment) Confirm() error {
	if p.isConfirmed {
		return ErrAlreadyConfirmed
	}

	now := time.Now().UTC()
	p.isConfirmed = true
	p.confirmedAt = &now

	return nil
}

func (p *Payment) ID() uuid.UUID {
	return p.id
}

func (p *Payment) PaymentID() uuid.UUID {
	return p.paymentID
}

func (p *Payment) UserID() uuid.UUID {
	return p.userID
}

func (p *Payment) AmountPenny() uint64 {
	return p.amountPenny
}

func (p *Payment) IsConfirmed() bool {
	return p.isConfirmed
}

func (p *Payment) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Payment) ConfirmedAt() *time.Time {
	return p.confirmedAt
}
