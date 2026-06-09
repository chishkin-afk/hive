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
		lastVisit: now,
		createdAt: now,
	}, nil
}

func (c *Client) AddVisit() {
	c.visits++
	c.lastVisit = time.Now().UTC()
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
