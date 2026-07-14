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

type registrationRepository struct {
	sqlBuilder squirrel.StatementBuilderType
	log        *slog.Logger
	db         DB
}

func New(log *slog.Logger, db DB) *registrationRepository {
	return &registrationRepository{
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		log:        log,
		db:         db,
	}
}

// Save is a function that saves a registration record
// to the database; otherwise, it returns an error.
func (rr *registrationRepository) Save(ctx context.Context, registration *registration.Registration) error {
	rr.log.Debug("saving registration in db",
		slog.String("registration_id", registration.ID().String()),
	)

	if err := rr.save(ctx, rr.db, registration); err != nil {
		return fmt.Errorf("failed to save registration: %w",
			handleError(err),
		)
	}

	return nil
}

func (rr *registrationRepository) save(ctx context.Context, db DB, registration *registration.Registration) error {
	query, args, err := rr.buildInsertQuery(registrationToRecord(registration))
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (rr *registrationRepository) buildInsertQuery(record *registrationRecord) (string, []any, error) {
	return rr.sqlBuilder.Insert("registrations").Columns(
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

// Count is a function that calculates
// the total number of user registrations
// based on their fingerprint.
func (rr *registrationRepository) Count(ctx context.Context, fingerprint string) (int64, error) {
	rr.log.Debug("counting registrations of user")

	count, err := rr.count(ctx, rr.db, fingerprint)
	if err != nil {
		return 0, fmt.Errorf("failed to count registrations: %w",
			handleError(err),
		)
	}

	return count, nil
}

func (rr *registrationRepository) count(ctx context.Context, db DB, fingerprint string) (int64, error) {
	query, args, err := rr.buildCountQuery(fingerprint)
	if err != nil {
		return 0, err
	}

	return rr.mapCountRow(db.QueryRowContext(ctx, query, args...))
}

func (rr *registrationRepository) buildCountQuery(fingerprint string) (string, []any, error) {
	return rr.sqlBuilder.Select("count(*)").
		From("registrations").
		Where("fingerprint = ?", fingerprint).
		ToSql()
}

func (rr *registrationRepository) mapCountRow(row *sql.Row) (int64, error) {
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	if err := row.Err(); err != nil {
		return 0, err
	}

	return count, nil
}

func handleError(err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return registration.ErrNotFound
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23503":
			return registration.ErrUserNotFound
		}
	}

	return err
}
