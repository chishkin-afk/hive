package registrationpg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/devathh/hive/backend/internal/modules/auth/domain/registration"
	"github.com/lib/pq"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type registrationPersistenceRepository struct {
	sqlBuiler squirrel.StatementBuilderType
	log       *slog.Logger
	db        DB
}

func New(log *slog.Logger, db DB) *registrationPersistenceRepository {
	return &registrationPersistenceRepository{
		sqlBuiler: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		log:       log,
		db:        db,
	}
}

func (rpr *registrationPersistenceRepository) Save(ctx context.Context, registration *registration.Registration) (*registration.Registration, error) {
	rpr.log.Debug("building the insert query of registration",
		slog.String("registration_id", registration.ID().String()),
	)

	record := registrationToRecord(registration)
	query, args, err := rpr.buildInsertQuery(record)
	if err != nil {
		return nil, fmt.Errorf("failed to build insert query: %w", err)
	}

	rpr.log.Debug("exec the insert query of registration",
		slog.String("query", query),
	)

	if _, err := rpr.db.ExecContext(ctx, query, args...); err != nil {
		if safeErr, ok := rpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to save registration into db: %w", err)
	}

	return recordToRegistration(record), nil
}

func (rpr *registrationPersistenceRepository) buildInsertQuery(record *registrationRecord) (string, []any, error) {
	return rpr.sqlBuiler.Insert("registrations").Columns(
		"id",
		"user_id",
		"fingerprint",
		"created_at",
	).Values(
		record.ID,
		record.UserID,
		record.Fingerprint,
		record.CreatedAt,
	).ToSql()
}

func (rpr *registrationPersistenceRepository) Count(ctx context.Context, fingerprint string) (int64, error) {
	rpr.log.Debug("building the count query of registrations",
		slog.String("fingerprint", fingerprint),
	)

	query, args, err := rpr.buildCountQuery(fingerprint)
	if err != nil {
		return 0, fmt.Errorf("failed to build count query: %w", err)
	}

	rpr.log.Debug("exec the count query of registrations",
		slog.String("query", query),
	)

	count, err := rpr.mapCountRow(rpr.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		if safeErr, ok := rpr.handleError(err); ok {
			return 0, safeErr
		}

		return 0, fmt.Errorf("failed to count all the registrations: %w", err)
	}

	return count, nil
}

func (rpr *registrationPersistenceRepository) buildCountQuery(fingerprint string) (string, []any, error) {
	return rpr.sqlBuiler.Select("count(*)").From("registrations").Where("fingerprint = ?", fingerprint).ToSql()
}

func (rpr *registrationPersistenceRepository) mapCountRow(row *sql.Row) (int64, error) {
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	if err := row.Err(); err != nil {
		return 0, err
	}

	return count, nil
}

func (rpr *registrationPersistenceRepository) handleError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return err, true
	case errors.Is(err, sql.ErrNoRows):
		return registration.ErrNotFound, true
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23503":
			return registration.ErrUserNotFound, true
		}
	}

	return err, false
}
