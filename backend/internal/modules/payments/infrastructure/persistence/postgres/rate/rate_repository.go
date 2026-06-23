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

type ratePersistenceRepository struct {
	log        *slog.Logger
	db         DB
	sqlBuilder squirrel.StatementBuilderType
}

func New(log *slog.Logger, db DB) *ratePersistenceRepository {
	return &ratePersistenceRepository{
		log:        log,
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (rpr *ratePersistenceRepository) GetByID(ctx context.Context, id uuid.UUID) (*rate.Rate, error) {
	rpr.log.Debug("build get by id query", slog.String("id", id.String()))
	query, args, err := rpr.buildSelectQuery(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate by id: %w", err)
	}

	rpr.log.Debug("exec the select query by id", slog.String("query", query))
	record, err := rpr.mapRow(rpr.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		if safeErr, ok := rpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get rate from db: %w", err)
	}

	return recordToRate(record), nil
}

func (rpr *ratePersistenceRepository) buildSelectQuery(id uuid.UUID) (string, []any, error) {
	return rpr.sqlBuilder.Select(
		"id",
		"name",
		"amount_penny",
		"paid_days",
	).From("rates").Where("id = ?", id).ToSql()
}

func (rpr *ratePersistenceRepository) ListAll(ctx context.Context) ([]*rate.Rate, error) {
	rpr.log.Debug("build list query of rates")
	query, args, err := rpr.buildListQuery()
	if err != nil {
		return nil, fmt.Errorf("failed to list all rates: %w", err)
	}

	rpr.log.Debug("exec the list query of rates", slog.String("query", query))
	rows, err := rpr.db.QueryContext(ctx, query, args...)
	if err != nil {
		if safeErr, ok := rpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get list of rates: %w", err)
	}
	defer rows.Close()

	records, err := rpr.mapRows(rows)
	if err != nil {
		if safeErr, ok := rpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to map rows: %w", err)
	}

	rates := make([]*rate.Rate, len(records))
	for i, record := range records {
		rates[i] = recordToRate(record)
	}

	return rates, nil
}

func (rpr *ratePersistenceRepository) buildListQuery() (string, []any, error) {
	return rpr.sqlBuilder.Select(
		"id",
		"name",
		"amount_penny",
		"paid_days",
	).From("rates").Limit(100).ToSql()
}

func (rpr *ratePersistenceRepository) mapRows(rows *sql.Rows) ([]*rateRecord, error) {
	var records []*rateRecord
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

		records = append(records, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (rpr *ratePersistenceRepository) mapRow(row *sql.Row) (*rateRecord, error) {
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

	return &record, nil
}

func (rpr *ratePersistenceRepository) handleError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return err, true
	case errors.Is(err, sql.ErrNoRows):
		return rate.ErrRateNotFound, true
	}

	return err, false
}
