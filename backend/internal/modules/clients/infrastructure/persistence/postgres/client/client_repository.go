package clientpg

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/devathh/hive/backend/internal/modules/clients/domain/client"
	"github.com/google/uuid"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type clientPersistenceRepository struct {
	sqlBuiler squirrel.StatementBuilderType
	log       *slog.Logger
	db        DB
}

func New(log *slog.Logger, db DB) *clientPersistenceRepository {
	return &clientPersistenceRepository{
		sqlBuiler: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		log:       log,
		db:        db,
	}
}

func (cpr *clientPersistenceRepository) Save(ctx context.Context, client *client.Client) (*client.Client, error) {
	panic("not impl")
}

func (cpr *clientPersistenceRepository) Update(ctx context.Context, updFunc client.UpdateFunc, id uuid.UUID) (*client.Client, error) {
	panic("not impl")
}

func (cpr *clientPersistenceRepository) GetByID(ctx context.Context, id uuid.UUID) (*client.Client, error) {
	panic("not impl")
}

func (cpr *clientPersistenceRepository) GetByPairIDs(ctx context.Context, userID, offerID uuid.UUID) (*client.Client, error) {
	panic("not impl")
}

func (cpr *clientPersistenceRepository) ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) ([]*client.Client, int64, error) {
	panic("not impl")
}

func (cpr *clientPersistenceRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*client.Client, int64, error) {
	panic("not impl")
}

func (cpr *clientPersistenceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	panic("not impl")
}
