package client

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidUserID  = errors.New("invalid user's id")
	ErrInvalidOfferID = errors.New("invalid offer's id")
)

type Client struct {
	id        uuid.UUID
	userID    uuid.UUID
	offerID   uuid.UUID
	visits    uint64
	lastVisit time.Time
	createdAt time.Time
}

func New(
	userID uuid.UUID,
	offerID uuid.UUID,
	visit time.Time,
) (*Client, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	if offerID == uuid.Nil {
		return nil, ErrInvalidOfferID
	}

	now := time.Now().UTC()
	return &Client{
		id:        uuid.New(),
		userID:    userID,
		offerID:   offerID,
		visits:    1,
		lastVisit: visit.UTC(),
		createdAt: now,
	}, nil
}

func From(
	id uuid.UUID,
	userID uuid.UUID,
	offerID uuid.UUID,
	visits uint64,
	lastVisit time.Time,
	createdAt time.Time,
) *Client {
	return &Client{
		id:        id,
		userID:    userID,
		offerID:   offerID,
		visits:    visits,
		lastVisit: lastVisit,
		createdAt: createdAt,
	}
}

func (c *Client) AddVisit(visit time.Time) {
	c.visits++
	c.lastVisit = visit.UTC()
}

func (c *Client) ID() uuid.UUID {
	return c.id
}

func (c *Client) UserID() uuid.UUID {
	return c.userID
}

func (c *Client) OfferID() uuid.UUID {
	return c.offerID
}

func (c *Client) Visits() uint64 {
	return c.visits
}

func (c *Client) LastVisit() time.Time {
	return c.lastVisit
}

func (c *Client) CreatedAt() time.Time {
	return c.createdAt
}
