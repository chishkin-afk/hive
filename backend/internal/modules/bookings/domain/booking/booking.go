package booking

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type (
	CreateFunc func(ctx context.Context, bookings []*Booking) (*Booking, error)
	UpdateFunc func(ctx context.Context, booking *Booking) error
)

var (
	ErrInvalidComment = errors.New("len of comment must be less than 128")
	ErrEmptyOfferID   = errors.New("offer's id is empty")
	ErrEmptyUserID    = errors.New("user's id is empty")
	ErrCancelled      = errors.New("booking is cancelled")
)

type Booking struct {
	id          uuid.UUID
	offerID     uuid.UUID
	userID      uuid.UUID
	comment     string
	startTime   time.Time
	isCancelled bool
	createdAt   time.Time
	updatedAt   time.Time

	offerDurationMinutes uint32
	offerIntervalMinutes uint32
}

func New(
	offerID, userID uuid.UUID,
	comment string,
	startTime time.Time,
) (*Booking, error) {
	if offerID == uuid.Nil {
		return nil, ErrEmptyOfferID
	}

	if userID == uuid.Nil {
		return nil, ErrEmptyUserID
	}

	comment = strings.TrimSpace(comment)
	if len([]rune(comment)) > 128 {
		return nil, ErrInvalidComment
	}

	now := time.Now().UTC()
	return &Booking{
		id:          uuid.New(),
		offerID:     offerID,
		userID:      userID,
		comment:     comment,
		startTime:   startTime.UTC(),
		isCancelled: false,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func From(
	id uuid.UUID,
	offerID, userID uuid.UUID,
	comment string,
	startTime time.Time,
	isCancelled bool,
	createdAt time.Time,
	updatedAt time.Time,
) *Booking {
	return &Booking{
		id:          id,
		offerID:     offerID,
		userID:      userID,
		comment:     comment,
		startTime:   startTime.UTC(),
		isCancelled: isCancelled,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (b *Booking) SetOfferTiming(durationMinutes, intervalMinutes uint32) {
	b.offerDurationMinutes = durationMinutes
	b.offerIntervalMinutes = intervalMinutes
}

func (b *Booking) Cancel() error {
	if b.isCancelled {
		return ErrCancelled
	}

	b.isCancelled = true
	b.updatedAt = time.Now().UTC()

	return nil
}

func (b *Booking) ChangeComment(newComment string) error {
	if b.isCancelled {
		return ErrCancelled
	}

	b.comment = strings.TrimSpace(newComment)
	b.updatedAt = time.Now().UTC()

	return nil
}

func (b *Booking) ChangeStartTime(newStartTime time.Time) error {
	if b.isCancelled {
		return ErrCancelled
	}

	b.startTime = newStartTime
	b.updatedAt = time.Now().UTC()

	return nil
}

func (b *Booking) ID() uuid.UUID {
	return b.id
}

func (b *Booking) OfferID() uuid.UUID {
	return b.offerID
}

func (b *Booking) UserID() uuid.UUID {
	return b.userID
}

func (b *Booking) Comment() string {
	return b.comment
}

func (b *Booking) StartTime() time.Time {
	return b.startTime
}

func (b *Booking) IsCancelled() bool {
	return b.isCancelled
}

func (b *Booking) OfferDurationMinutes() uint32 {
	return b.offerDurationMinutes
}

func (b *Booking) OfferIntervalMinutes() uint32 {
	return b.offerIntervalMinutes
}

func (b *Booking) CreatedAt() time.Time {
	return b.createdAt
}

func (b *Booking) UpdatedAt() time.Time {
	return b.updatedAt
}
