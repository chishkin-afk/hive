package offer

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type OfferPersistenceRepository interface {
	Save(ctx context.Context, offer *Offer) (*Offer, error)
	Update(ctx context.Context, updOffer *Offer) (*Offer, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Offer, error)
	ListByAuthorID(ctx context.Context, authorID uuid.UUID, page, limit uint32) ([]*Offer, uint64, error)
}

type OfferCacheRepository interface {
	Set(ctx context.Context, offer *Offer) error
	Get(ctx context.Context, id uuid.UUID) (*Offer, error)
	Del(ctx context.Context, id uuid.UUID) error
}

var (
	ErrOfferNotFound  = errors.New("offer not found")
	ErrAuthorNotFound = errors.New("author of offer not found")
)
