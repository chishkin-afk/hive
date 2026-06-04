package offer

import (
	"context"

	"github.com/google/uuid"
)

type OfferPersistenceRepository interface {
	Save(ctx context.Context, offer *Offer) (*Offer, error)
	Update(ctx context.Context, updOffer *Offer) (*Offer, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Offer, error)
	ListByAuthorID(ctx context.Context, authorID uuid.UUID) ([]*Offer, error)
}
