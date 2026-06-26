package registration

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyUserID      = errors.New("user's id is empty")
	ErrEmptyFingerprint = errors.New("fingerprint is empty")
)

type Registration struct {
	id          uuid.UUID
	userID      uuid.UUID
	fingerprint string
	createdAt   time.Time
}

func New(userID uuid.UUID, fingerprint string) (*Registration, error) {
	if userID == uuid.Nil {
		return nil, ErrEmptyUserID
	}

	fingerprint = strings.TrimSpace(fingerprint)
	if fingerprint == "" {
		return nil, ErrEmptyFingerprint
	}

	return &Registration{
		id:          uuid.New(),
		userID:      userID,
		fingerprint: fingerprint,
		createdAt:   time.Now().UTC(),
	}, nil
}

func From(
	id uuid.UUID,
	userID uuid.UUID,
	fingerprint string,
	createdAt time.Time,
) *Registration {
	return &Registration{
		id:          id,
		userID:      userID,
		fingerprint: fingerprint,
		createdAt:   createdAt,
	}
}

func (r *Registration) ID() uuid.UUID {
	return r.id
}

func (r *Registration) UserID() uuid.UUID {
	return r.userID
}

func (r *Registration) Fingerprint() string {
	return r.fingerprint
}

func (r *Registration) CreatedAt() time.Time {
	return r.createdAt
}
