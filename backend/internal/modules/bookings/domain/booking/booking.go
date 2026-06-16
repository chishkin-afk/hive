package booking

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyOfferID = errors.New("offer's id is empty")
	ErrEmptyUserID  = errors.New("user's id is empty")
	ErrCancelled    = errors.New("booking is cancelled")
)

type Booking struct {
	id          uuid.UUID
	offerID     uuid.UUID
	userID      uuid.UUID
	comment     string
	startTime   DayTime
	isCancelled bool
	createdAt   time.Time
	updatedAt   time.Time
}

func New(
	offerID, userID uuid.UUID,
	comment string,
	startTime DayTime,
) (*Booking, error) {
	if offerID == uuid.Nil {
		return nil, ErrEmptyOfferID
	}

	if userID == uuid.Nil {
		return nil, ErrEmptyUserID
	}

	now := time.Now().UTC()
	return &Booking{
		id:          uuid.New(),
		offerID:     offerID,
		userID:      userID,
		comment:     comment,
		startTime:   startTime,
		isCancelled: false,
		createdAt:   now,
		updatedAt:   now,
	}, nil
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

func (b *Booking) ChangeStartTime(newStartTime DayTime) error {
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

func (b *Booking) StartTime() DayTime {
	return b.startTime
}

func (b *Booking) IsCancelled() bool {
	return b.isCancelled
}

func (b *Booking) CreatedAt() time.Time {
	return b.createdAt
}

func (b *Booking) UpdatedAt() time.Time {
	return b.updatedAt
}
