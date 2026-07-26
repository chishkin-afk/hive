package bookingpg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/devathh/hive/backend/internal/modules/bookings/domain/booking"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type WithTx interface {
	DB
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type bookingRepository struct {
	log        *slog.Logger
	db         WithTx
	sqlBuilder squirrel.StatementBuilderType
}

func New(log *slog.Logger, db WithTx) *bookingRepository {
	return &bookingRepository{
		log:        log,
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save is a function that first checks the number of user's bookings for this coach,
// then requests the coach's offer ID, and calls the callback function with these arguments.
// If the callback function does not return an error, it saves the user.
// Everything happens within a transaction
func (br *bookingRepository) Save(ctx context.Context, offerID uuid.UUID, startTime, endTime time.Time, cbFunc booking.CreateFunc) (*booking.Booking, error) {
	br.log.Debug("start tx")
	tx, rollback, err := br.beginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start tx: %w", err)
	}
	defer rollback()

	br.log.Debug("listing all bookings per day",
		slog.String("offer_id", offerID.String()),
	)

	bookings, err := br.listAllPerDay(ctx, tx, offerID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	booking, err := cbFunc(ctx, bookings)
	if err != nil {
		return nil, err
	}

	br.log.Debug("saving booking into db",
		slog.String("booking_id", booking.ID().String()),
	)

	if err := br.save(ctx, tx, booking); err != nil {
		return nil, fmt.Errorf("failed to save booking into db: %w",
			handleError(err),
		)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	return booking, nil
}

func (br *bookingRepository) save(ctx context.Context, db DB, booking *booking.Booking) error {
	record := bookingToRecord(booking)
	query, args, err := br.buildInsertQuery(record)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (br *bookingRepository) buildInsertQuery(record *bookingRecord) (string, []any, error) {
	return br.sqlBuilder.Insert("bookings").Columns(
		"id",
		"offer_id",
		"user_id",
		"comment",
		"start_time",
		"is_cancelled",
		"created_at",
		"updated_at",
	).Values(
		record.ID,
		record.OfferID,
		record.UserID,
		record.Comment,
		record.StartTime,
		record.IsCancelled,
		record.CreatedAt,
		record.UpdatedAt,
	).ToSql()
}

func (br *bookingRepository) listAllPerDay(
	ctx context.Context,
	db DB,
	offerID uuid.UUID,
	startTime,
	endTime time.Time,
) ([]*booking.Booking, error) {
	query, args, err := br.buildListAllPerTime(
		offerID,
		startTime,
		endTime,
	)
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

// ListByOfferID returns a list of all bookings for a specific offer.
// It is important to check whether
// the user is the author of the offer before doing so
func (br *bookingRepository) ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) ([]*booking.Booking, int64, error) {
	br.log.Debug("counting booking by offer's id",
		slog.String("offer_id", offerID.String()),
	)

	count, err := br.countByOfferID(ctx, br.db, offerID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count bookings: %w",
			handleError(err),
		)
	}

	bookings, err := br.listByOfferID(ctx, br.db, offerID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list bookings by offer's id: %w",
			handleError(err),
		)
	}

	return bookings, count, nil
}

func (br *bookingRepository) listByOfferID(ctx context.Context, db DB, offerID uuid.UUID, page, limit uint32) ([]*booking.Booking, error) {
	query, args, err := br.buildListByOfferIDQuery(offerID, page, limit)
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

func (br *bookingRepository) buildListByOfferIDQuery(offerID uuid.UUID, page, limit uint32) (string, []any, error) {
	offset := limit * (page - 1)
	return br.sqlBuilder.Select(
		"id",
		"offer_id",
		"user_id",
		"comment",
		"start_time",
		"is_cancelled",
		"created_at",
		"updated_at",
	).From("bookings").Limit(uint64(limit)).Offset(uint64(offset)).Where("offer_id = ?", offerID).ToSql()
}

func (br *bookingRepository) countByOfferID(ctx context.Context, db DB, offerID uuid.UUID) (int64, error) {
	query, args, err := br.buildCountByOfferIDQuery(offerID)
	if err != nil {
		return 0, nil
	}

	return mapCountRow(db.QueryRowContext(ctx, query, args...))
}

func (br *bookingRepository) buildCountByOfferIDQuery(offerID uuid.UUID) (string, []any, error) {
	return br.sqlBuilder.Select("count(*)").
		From("bookings").
		Where("offer_id = ?", offerID).
		ToSql()
}

// ListByUserID is a function that returns
// a list of all bookings for a specific user.
// The UserID should be taken from the token
func (br *bookingRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*booking.Booking, int64, error) {
	br.log.Debug("counting bookings by user's id",
		slog.String("user_id", userID.String()),
	)

	count, err := br.countByUserID(ctx, br.db, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count bookings by user's id: %w",
			handleError(err),
		)
	}

	bookings, err := br.listByUserID(ctx, br.db, userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list bookings by user's id: %w",
			handleError(err),
		)
	}

	return bookings, count, nil
}

func (br *bookingRepository) listByUserID(ctx context.Context, db DB, userID uuid.UUID, page, limit uint32) ([]*booking.Booking, error) {
	query, args, err := br.buildListByUserIDQuery(userID, page, limit)
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

func (br *bookingRepository) buildListByUserIDQuery(userID uuid.UUID, page, limit uint32) (string, []any, error) {
	offset := limit * (page - 1)
	return br.sqlBuilder.Select(
		"id",
		"offer_id",
		"user_id",
		"comment",
		"start_time",
		"is_cancelled",
		"created_at",
		"updated_at",
	).From("bookings").Limit(uint64(limit)).Offset(uint64(offset)).Where("user_id = ?", userID).ToSql()
}

func (br *bookingRepository) countByUserID(ctx context.Context, db DB, userID uuid.UUID) (int64, error) {
	query, args, err := br.buildCountByUserIDQuery(userID)
	if err != nil {
		return 0, err
	}

	return mapCountRow(db.QueryRowContext(ctx, query, args...))
}

func (br *bookingRepository) buildCountByUserIDQuery(userID uuid.UUID) (string, []any, error) {
	return br.sqlBuilder.Select("count(*)").
		From("bookings").
		Where("user_id = ?", userID).
		ToSql()
}

// ListAllPerTime is a function that returns
// a list of all bookings (with a limit)
// in a specified period, otherwise an error
func (br *bookingRepository) ListAllPerTime(ctx context.Context, offerID uuid.UUID, startTime, endTime time.Time) ([]*booking.Booking, error) {
	br.log.Debug("listing all bookings per time",
		slog.String("offer_id", offerID.String()),
	)

	bookings, err := br.listAllPerTime(ctx, br.db, offerID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookings per time: %w",
			handleError(err),
		)
	}

	return bookings, nil
}

func (br *bookingRepository) listAllPerTime(ctx context.Context, db DB, offerID uuid.UUID, startTime, endTime time.Time) ([]*booking.Booking, error) {
	query, args, err := br.buildListAllPerTime(offerID, startTime, endTime)
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

func (br *bookingRepository) buildListAllPerTime(offerID uuid.UUID, startTime, endTime time.Time) (string, []any, error) {
	return br.sqlBuilder.Select(
		"id",
		"offer_id",
		"user_id",
		"comment",
		"start_time",
		"is_cancelled",
		"created_at",
		"updated_at",
	).From("bookings").Limit(1000).Where("offer_id = ? AND start_time >= ? AND start_time < ?", offerID, startTime, endTime).ToSql()
}

func (br *bookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*booking.Booking, error) {
	br.log.Debug("getting booking by id",
		slog.String("booking_id", id.String()),
	)

	booking, err := br.getByID(ctx, br.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking by id: %w",
			handleError(err),
		)
	}

	return booking, nil
}

func (br *bookingRepository) getByID(ctx context.Context, db DB, id uuid.UUID) (*booking.Booking, error) {
	query, args, err := br.buildSelectByIDQuery(id)
	if err != nil {
		return nil, err
	}

	return mapRow(db.QueryRowContext(ctx, query, args...))
}

func (br *bookingRepository) buildSelectByIDQuery(id uuid.UUID) (string, []any, error) {
	return br.sqlBuilder.Select(
		"id",
		"offer_id",
		"user_id",
		"comment",
		"start_time",
		"is_cancelled",
		"created_at",
		"updated_at",
	).From("bookings").Where("id = ?", id).ToSql()
}

// Update is a function that updates a record
// (for example, the isCancelled field).
// It receives a booking's ID and
// calls updFunc, then saves the changes
func (br *bookingRepository) Update(ctx context.Context, id uuid.UUID, updFunc booking.UpdateFunc) (*booking.Booking, error) {
	br.log.Debug("start tx")
	tx, rollback, err := br.beginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	br.log.Debug("getting booking for update",
		slog.String("booking_id", id.String()),
	)

	booking, err := br.getByIDForUpdate(ctx, tx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking by id: %w",
			handleError(err),
		)
	}

	if err := updFunc(ctx, booking); err != nil {
		return nil, err
	}

	if err := br.update(ctx, tx, booking); err != nil {
		return nil, fmt.Errorf("failed to update booking: %w",
			handleError(err),
		)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	return booking, nil
}

func (br *bookingRepository) update(ctx context.Context, db DB, updBooking *booking.Booking) error {
	record := bookingToRecord(updBooking)
	query, args, err := br.buildUpdateQuery(record)
	if err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return booking.ErrBookingNotFound
	}

	return nil
}

func (br *bookingRepository) buildUpdateQuery(record *bookingRecord) (string, []any, error) {
	return br.sqlBuilder.Update("bookings").SetMap(map[string]any{
		"comment":      record.Comment,
		"start_time":   record.StartTime,
		"is_cancelled": record.IsCancelled,
		"updated_at":   record.UpdatedAt,
	}).Where("id = ?", record.ID).ToSql()
}

func (br *bookingRepository) getByIDForUpdate(ctx context.Context, db DB, id uuid.UUID) (*booking.Booking, error) {
	query, args, err := br.buildSelectByIDForUpdateQuery(id)
	if err != nil {
		return nil, err
	}

	return mapRow(db.QueryRowContext(ctx, query, args...))
}

func (br *bookingRepository) buildSelectByIDForUpdateQuery(id uuid.UUID) (string, []any, error) {
	return br.sqlBuilder.Select(
		"id",
		"offer_id",
		"user_id",
		"comment",
		"start_time",
		"is_cancelled",
		"created_at",
		"updated_at",
	).From("bookings").Where("id = ?", id).Suffix("FOR UPDATE").ToSql()
}

func (br *bookingRepository) beginTx(ctx context.Context) (*sql.Tx, func(), error) {
	tx, err := br.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, nil, err
	}

	return tx, func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			br.log.Error("failed to rollback tx",
				slog.String("error", err.Error()),
			)
		}
	}, nil
}

func mapRows(rows *sql.Rows) ([]*booking.Booking, error) {
	var bookings []*booking.Booking
	for rows.Next() {
		var record bookingRecord
		if err := rows.Scan(
			&record.ID,
			&record.OfferID,
			&record.UserID,
			&record.Comment,
			&record.StartTime,
			&record.IsCancelled,
			&record.CreatedAt,
			&record.UpdatedAt,
		); err != nil {
			return nil, err
		}

		bookings = append(bookings, recordToBooking(&record))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookings, nil
}

func mapRow(row *sql.Row) (*booking.Booking, error) {
	var record bookingRecord
	if err := row.Scan(
		&record.ID,
		&record.OfferID,
		&record.UserID,
		&record.Comment,
		&record.StartTime,
		&record.IsCancelled,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if err := row.Err(); err != nil {
		return nil, err
	}

	return recordToBooking(&record), nil
}

func mapCountRow(row *sql.Row) (int64, error) {
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func handleError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return booking.ErrBookingNotFound
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return booking.ErrBookingAlreadyExists
		case "23503":
			return booking.ErrInvalidOfferID
		}
	}

	return err
}
