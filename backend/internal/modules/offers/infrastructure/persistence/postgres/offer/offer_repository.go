package offerpg

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/devathh/hive/backend/internal/modules/offers/domain/offer"
	"github.com/google/uuid"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (*sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type offerPersistenceRepository struct {
	log *slog.Logger
	db  DB
}

func New(log *slog.Logger, db DB) *offerPersistenceRepository {
	return &offerPersistenceRepository{
		log: log,
		db:  db,
	}
}

func (opr *offerPersistenceRepository) Save(ctx context.Context, offer *offer.Offer) (*offer.Offer, error) {
	panic("not impl")
}

func (opr *offerPersistenceRepository) Update(ctx context.Context, updOffer *offer.Offer) (*offer.Offer, error) {
	panic("not impl")
}

func (opr *offerPersistenceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	panic("not impl")
}

func (opr *offerPersistenceRepository) GetByID(ctx context.Context, id uuid.UUID) (*offer.Offer, error) {
	panic("not impl")
}

func (opr *offerPersistenceRepository) ListByAuthorID(ctx context.Context, authorID uuid.UUID, page, limit uint32) ([]*offer.Offer, uint64, error) {
	panic("not impl")
}
