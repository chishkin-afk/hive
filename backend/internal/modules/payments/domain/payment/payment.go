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
	ErrAlreadyCancelled = errors.New("payment is already cancelled")
	ErrAlreadyConfirmed = errors.New("payment is already confirmed")
)

type Payment struct {
	id          uuid.UUID
	paymentID   uuid.UUID
	userID      uuid.UUID
	amountPenny uint64
	status      Status
	createdAt   time.Time
	updatedAt   time.Time
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
		status:      NewStatus(StatusPending),
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func From(
	id uuid.UUID,
	paymentID uuid.UUID,
	userID uuid.UUID,
	amountPenny uint64,
	status Status,
	createdAt time.Time,
	updatedAt time.Time,
) *Payment {
	return &Payment{
		id:          id,
		paymentID:   paymentID,
		userID:      userID,
		amountPenny: amountPenny,
		status:      status,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (p *Payment) Confirm() error {
	if p.status.Value() == StatusConfirmed {
		return ErrAlreadyConfirmed
	}

	p.status = NewStatus(StatusConfirmed)
	p.updatedAt = time.Now().UTC()

	return nil
}

func (p *Payment) Cancel() error {
	if p.status.Value() == StatusCanceled {
		return ErrAlreadyCancelled
	}

	p.status = NewStatus(StatusCanceled)
	p.updatedAt = time.Now().UTC()

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

func (p *Payment) Status() Status {
	return p.status
}

func (p *Payment) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Payment) UpdatedAt() time.Time {
	return p.updatedAt
}
