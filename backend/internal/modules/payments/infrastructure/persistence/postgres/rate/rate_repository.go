package ratepg

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/devathh/hive/backend/internal/modules/payments/domain/rate"
	"github.com/google/uuid"
)

type DB interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type ratePersistenceRepository struct {
	log *slog.Logger
	db  DB
}

func New(log *slog.Logger, db DB) *ratePersistenceRepository {
	return &ratePersistenceRepository{
		log: log,
		db:  db,
	}
}

func (rpr *ratePersistenceRepository) GetByID(ctx context.Context, id uuid.UUID) (*rate.Rate, error) {
	panic("not impl")
}

func (rpr *ratePersistenceRepository) ListAll(ctx context.Context) ([]*rate.Rate, error) {
	panic("not impl")
}
