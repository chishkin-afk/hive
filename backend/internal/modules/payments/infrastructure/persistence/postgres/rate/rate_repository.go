package ratepg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/devathh/hive/backend/internal/modules/payments/domain/rate"
	"github.com/google/uuid"
)

type DB interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type rateRepository struct {
	log        *slog.Logger
	db         DB
	sqlBuilder squirrel.StatementBuilderType
}

func New(log *slog.Logger, db DB) *rateRepository {
	return &rateRepository{
		log:        log,
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// GetByID is a method that retrieves a rate
// by its ID by executing a SELECT query,
// or returns an error otherwise.
func (rr *rateRepository) GetByID(ctx context.Context, id uuid.UUID) (*rate.Rate, error) {
	rr.log.Debug("getting rate by id",
		slog.String("rate_id", id.String()),
	)

	rate, err := rr.getByID(ctx, rr.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate by id: %w",
			handleError(err),
		)
	}

	return rate, nil
}

func (rr *rateRepository) getByID(ctx context.Context, db DB, id uuid.UUID) (*rate.Rate, error) {
	query, args, err := rr.buildSelectQuery(id)
	if err != nil {
		return nil, err
	}

	return mapRow(db.QueryRowContext(ctx, query, args...))
}

func (rr *rateRepository) buildSelectQuery(id uuid.UUID) (string, []any, error) {
	return rr.sqlBuilder.Select(
		"id",
		"name",
		"amount_penny",
		"paid_days",
	).From("rates").Where("id = ?", id).ToSql()
}

// ListAll is a method that executes a SELECT query
// to retrieve rates (subject to a limit)
// or returns an error.
func (rr *rateRepository) ListAll(ctx context.Context) ([]*rate.Rate, error) {
	rr.log.Debug("listing all rates")

	list, err := rr.listAll(ctx, rr.db)
	if err != nil {
		return nil, fmt.Errorf("failed to list all rates: %w",
			handleError(err),
		)
	}

	return list, nil
}

func (rr *rateRepository) listAll(ctx context.Context, db DB) ([]*rate.Rate, error) {
	query, args, err := rr.buildListQuery()
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return mapRows(rows)
}

func (rr *rateRepository) buildListQuery() (string, []any, error) {
	return rr.sqlBuilder.Select(
		"id",
		"name",
		"amount_penny",
		"paid_days",
	).From("rates").Limit(100).ToSql()
}

func mapRows(rows *sql.Rows) ([]*rate.Rate, error) {
	var rates []*rate.Rate
	for rows.Next() {
		var record rateRecord
		if err := rows.Scan(
			&record.ID,
			&record.Name,
			&record.AmountPenny,
			&record.PaidDays,
		); err != nil {
			return nil, err
		}

		rates = append(rates, recordToRate(&record))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rates, nil
}

func mapRow(row *sql.Row) (*rate.Rate, error) {
	var record rateRecord
	if err := row.Scan(
		&record.ID,
		&record.Name,
		&record.AmountPenny,
		&record.PaidDays,
	); err != nil {
		return nil, err
	}

	if err := row.Err(); err != nil {
		return nil, err
	}

	return recordToRate(&record), nil
}

func handleError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return rate.ErrRateNotFound
	}

	return err
}
