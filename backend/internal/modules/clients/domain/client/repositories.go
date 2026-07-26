package client

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type UpdateFunc func(*Client) error

type ClientPersistenceRepository interface {
	Save(ctx context.Context, client *Client) error
	Update(ctx context.Context, updFunc UpdateFunc, id uuid.UUID) (*Client, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Client, error)
	GetByPairIDs(ctx context.Context, userID, offerID uuid.UUID) (*Client, error)
	ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) ([]*Client, int64, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*Client, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ClientCacheRepository interface {
	Save(ctx context.Context, client *Client) error
	GetByID(ctx context.Context, id uuid.UUID) (*Client, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

var (
	ErrClientNotFound      = errors.New("client not found")
	ErrClientAlreadyExists = errors.New("client already exists")
	ErrOfferNotFound       = errors.New("offer not found")
)
