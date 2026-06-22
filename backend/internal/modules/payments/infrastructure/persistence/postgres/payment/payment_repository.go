package paymentpg

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/devathh/hive/backend/internal/modules/payments/domain/payment"
	"github.com/google/uuid"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type paymentPersistenceRepository struct {
	log *slog.Logger
	db  DB
}

func New(log *slog.Logger, db DB) *paymentPersistenceRepository {
	return &paymentPersistenceRepository{
		log: log,
		db:  db,
	}
}

func (ppr *paymentPersistenceRepository) Save(ctx context.Context, payment *payment.Payment) (*payment.Payment, error) {
	panic("not impl")
}

func (ppr *paymentPersistenceRepository) GetByID(ctx context.Context, paymentID uuid.UUID) (*payment.Payment, error) {
	panic("not impl")
}

func (ppr *paymentPersistenceRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*payment.Payment, int64, error) {
	panic("not impl")
}

func (ppr *paymentPersistenceRepository) Update(ctx context.Context, paymentID, userID uuid.UUID, updFunc payment.UpdateFunc) (*payment.Payment, error) {
	panic("not impl")
}
