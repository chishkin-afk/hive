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
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type bookingPersistenceRepository struct {
	log        *slog.Logger
	db         DB
	sqlBuilder squirrel.StatementBuilderType
}

func New(log *slog.Logger, db DB) *bookingPersistenceRepository {
	return &bookingPersistenceRepository{
		log:        log,
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save is a function that first checks the number of user's bookings for this coach,
// then requests the coach's offer ID, and calls the callback function with these arguments.
// If the callback function does not return an error, it saves the user.
// Everything happens within a transaction
func (bpr *bookingPersistenceRepository) Save(ctx context.Context, offerID uuid.UUID, startTime, endTime time.Time, cbFunc booking.CreateFunc) (*booking.Booking, error) {
	bpr.log.Debug("start the transaction to save booking",
		slog.String("offer_id", offerID.String()),
	)

	tx, err := bpr.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start tx: %w", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				bpr.log.Error("failed to rollback tx", slog.String("error", err.Error()))
			}
		}
	}()

	records, err := bpr.getListAllPerDay(ctx, tx, offerID, startTime, endTime)
	if err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to count bookings per day: %w", err)
	}

	// Call the main validation function (with business-logic)
	booking, err := cbFunc(ctx, bpr.getBookings(records))
	if err != nil {
		return nil, err
	}

	record, err := bpr.insertTx(ctx, tx, bookingToRecord(booking))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	return recordToBooking(record), nil
}

func (bpr *bookingPersistenceRepository) getBookings(records []*bookingRecord) []*booking.Booking {
	bookings := make([]*booking.Booking, len(records))

	for i, record := range records {
		bookings[i] = recordToBooking(record)
	}

	return bookings
}

func (bpr *bookingPersistenceRepository) insertTx(ctx context.Context, tx *sql.Tx, record *bookingRecord) (*bookingRecord, error) {
	bpr.log.Debug("build the insert query", slog.String("booking_id", record.ID.String()))
	insertQuery, insertArgs, err := bpr.buildInsertQuery(record)
	if err != nil {
		return nil, fmt.Errorf("failed to build insert query: %w", err)
	}

	bpr.log.Debug("exec the insert query", slog.String("query", insertQuery))
	if _, err := tx.ExecContext(ctx, insertQuery, insertArgs...); err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to save booking into db: %w", err)
	}

	return record, nil
}

func (bpr *bookingPersistenceRepository) getListAllPerDay(ctx context.Context, tx *sql.Tx, offerID uuid.UUID, startTime, endTime time.Time) ([]*bookingRecord, error) {
	bpr.log.Debug("build count query per day")
	query, args, err := bpr.buildListAllPerDay(offerID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to build list all per day query: %w", err)
	}

	bpr.log.Debug("exec the list all per day query", slog.String("query", query))
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return bpr.mapRows(rows)
}

func (bpr *bookingPersistenceRepository) buildInsertQuery(record *bookingRecord) (string, []any, error) {
	return bpr.sqlBuilder.Insert("bookings").Columns(
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

// ListByOfferID returns a list of all bookings for a specific offer.
// It is important to check whether the user is the author of the offer before doing so
func (bpr *bookingPersistenceRepository) ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) ([]*booking.Booking, int64, error) {
	bpr.log.Debug("build the count query by offer's id", slog.String("offer_id", offerID.String()))
	countQuery, countArgs, err := bpr.buildCountQueryByOfferID(offerID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build count query: %w", err)
	}

	bpr.log.Debug("exec the count query", slog.String("query", countQuery))
	count, err := bpr.mapCountRow(bpr.db.QueryRowContext(ctx, countQuery, countArgs...))
	if err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, 0, safeErr
		}

		return nil, 0, fmt.Errorf("failed to count all bookings: %w", err)
	}

	bpr.log.Debug("build the list query", slog.String("offer_id", offerID.String()))
	listQuery, listArgs, err := bpr.buildListByOfferIDQuery(offerID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build list query: %w", err)
	}

	bpr.log.Debug("exec the list query", slog.String("query", listQuery))
	rows, err := bpr.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, 0, safeErr
		}

		return nil, 0, fmt.Errorf("failed to get list of bookings: %w", err)
	}
	defer rows.Close()

	records, err := bpr.mapRows(rows)
	if err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, 0, safeErr
		}

		return nil, 0, fmt.Errorf("failed to map rows: %w", err)
	}

	bookings := make([]*booking.Booking, len(records))
	for i, record := range records {
		bookings[i] = recordToBooking(record)
	}

	return bookings, int64(count), nil
}

func (bpr *bookingPersistenceRepository) buildCountQueryByOfferID(offerID uuid.UUID) (string, []any, error) {
	return bpr.buildRawCountQuery().Where("offer_id = ?", offerID).ToSql()
}

func (bpr *bookingPersistenceRepository) buildListByOfferIDQuery(offerID uuid.UUID, page, limit uint32) (string, []any, error) {
	offset := limit * (page - 1)
	return bpr.sqlBuilder.Select(
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

// ListByUserID is a function that returns a list of all bookings for a specific user.
// The UserID should be taken from the token
func (bpr *bookingPersistenceRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*booking.Booking, int64, error) {
	bpr.log.Debug("build the count query", slog.String("user_id", userID.String()))
	countQuery, countArgs, err := bpr.buildCountQueryByUserID(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build count query: %w", err)
	}

	bpr.log.Debug("exec the count query", slog.String("query", countQuery))
	count, err := bpr.mapCountRow(bpr.db.QueryRowContext(ctx, countQuery, countArgs...))
	if err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, 0, safeErr
		}

		return nil, 0, fmt.Errorf("failed to count bookings by user's id: %w", err)
	}

	bpr.log.Debug("build the list query by user id", slog.String("user_id", userID.String()))
	listQuery, listArgs, err := bpr.buildListByUserIDQuery(userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build list query: %w", err)
	}

	bpr.log.Debug("exec the list query", slog.String("query", listQuery))
	rows, err := bpr.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, 0, safeErr
		}

		return nil, 0, fmt.Errorf("failed to get list of bookings: %w", err)
	}
	defer rows.Close()

	records, err := bpr.mapRows(rows)
	if err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, 0, safeErr
		}

		return nil, 0, fmt.Errorf("failed to map rows: %w", err)
	}

	bookings := make([]*booking.Booking, len(records))
	for i, record := range records {
		bookings[i] = recordToBooking(record)
	}

	return bookings, int64(count), nil
}

func (bpr *bookingPersistenceRepository) buildCountQueryByUserID(userID uuid.UUID) (string, []any, error) {
	return bpr.buildRawCountQuery().Where("user_id = ?", userID).ToSql()
}

func (bpr *bookingPersistenceRepository) buildListByUserIDQuery(userID uuid.UUID, page, limit uint32) (string, []any, error) {
	offset := limit * (page - 1)
	return bpr.sqlBuilder.Select(
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

// ListAllByOfferID returns all future bookings for the offer
func (bpr *bookingPersistenceRepository) ListAllByOfferID(ctx context.Context, offerID uuid.UUID) ([]*booking.Booking, error) {
	bpr.log.Debug("build the list all of bookings by offer id query", slog.String("offer_id", offerID.String()))
	query, args, err := bpr.buildListAllQuery(offerID)
	if err != nil {
		return nil, fmt.Errorf("failed to build list all query: %w", err)
	}

	bpr.log.Debug("exec the list all query", slog.String("query", query))
	rows, err := bpr.db.QueryContext(ctx, query, args...)
	if err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get list of all future bookings: %w", err)
	}
	defer rows.Close()

	records, err := bpr.mapRows(rows)
	if err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to map rows: %w", err)
	}

	bookings := make([]*booking.Booking, len(records))
	for i, record := range records {
		bookings[i] = recordToBooking(record)
	}

	return bookings, nil
}

func (bpr *bookingPersistenceRepository) buildListAllQuery(offerID uuid.UUID) (string, []any, error) {
	return bpr.sqlBuilder.Select(
		"id",
		"offer_id",
		"user_id",
		"comment",
		"start_time",
		"is_cancelled",
		"created_at",
		"updated_at",
	).From("bookings").Limit(100000).Where("start_time >= NOW() AND offer_id = ?", offerID).ToSql()
}

func (bpr *bookingPersistenceRepository) ListAllPerTime(ctx context.Context, offerID uuid.UUID, startTime, endTime time.Time) ([]*booking.Booking, error) {
	bpr.log.Debug("build the list all query per time")
	query, args, err := bpr.buildListAllPerDay(offerID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to build list all query: %w", err)
	}

	bpr.log.Debug("exec the query", slog.String("query", query))
	rows, err := bpr.db.QueryContext(ctx, query, args...)
	if err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get list of all bookings: %w", err)
	}
	defer rows.Close()

	records, err := bpr.mapRows(rows)
	if err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to map rows: %w", err)
	}

	result := make([]*booking.Booking, len(records))
	for i, record := range records {
		result[i] = recordToBooking(record)
	}

	return result, nil
}

func (bpr *bookingPersistenceRepository) buildListAllPerDay(offerID uuid.UUID, startTime, endTime time.Time) (string, []any, error) {
	return bpr.sqlBuilder.Select(
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

// GetByID returns a single booking by its ID.
// It is important to check whether this record belongs to a userID
// or whether the offer owns a userID
func (bpr *bookingPersistenceRepository) GetByID(ctx context.Context, id uuid.UUID) (*booking.Booking, error) {
	bpr.log.Debug("build the select query", slog.String("booking_id", id.String()))
	query, args, err := bpr.buildSelectByIDQuery(id)
	if err != nil {
		return nil, fmt.Errorf("failed to build select by id query: %w", err)
	}

	bpr.log.Debug("exec the select by id query", slog.String("query", query))
	record, err := bpr.mapRow(bpr.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get booking by id: %w", err)
	}

	return recordToBooking(record), nil
}

func (bpr *bookingPersistenceRepository) buildSelectByIDQuery(id uuid.UUID) (string, []any, error) {
	return bpr.sqlBuilder.Select(
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

// Update is a function that updates a record (for example, the isCancelled field).
// It receives a booking's ID and calls updFunc, then saves the changes
func (bpr *bookingPersistenceRepository) Update(ctx context.Context, id uuid.UUID, updFunc booking.UpdateFunc) (*booking.Booking, error) {
	tx, err := bpr.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start tx: %w", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				bpr.log.Error("failed to rollback tx", slog.String("error", err.Error()))
			}
		}
	}()

	bpr.log.Debug("build the select by id query for update")
	selectQuery, selectArgs, err := bpr.buildSelectByIDQueryForUpdate(id)
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	bpr.log.Debug("exec the select query", slog.String("query", selectQuery))
	record, err := bpr.mapRow(tx.QueryRow(selectQuery, selectArgs...))
	if err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get booking for update: %w", err)
	}

	bpr.log.Debug("call the upd function", slog.String("booking_id", record.ID.String()))
	updBooking := recordToBooking(record)
	if err := updFunc(ctx, updBooking); err != nil {
		return nil, err
	}

	bpr.log.Debug("build update query", slog.String("booking_id", updBooking.ID().String()))
	updateQuery, updateArgs, err := bpr.buildUpdateQuery(bookingToRecord(updBooking))
	if err != nil {
		return nil, fmt.Errorf("failed to build update query: %w", err)
	}

	bpr.log.Debug("exec the update query", slog.String("query", updateQuery))
	result, err := tx.ExecContext(ctx, updateQuery, updateArgs...)
	if err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to update booking: %w", err)
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return nil, booking.ErrBookingNotFound
	}

	if err := tx.Commit(); err != nil {
		if safeErr, ok := bpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	return updBooking, nil
}

func (bpr *bookingPersistenceRepository) buildUpdateQuery(record *bookingRecord) (string, []any, error) {
	return bpr.sqlBuilder.Update("bookings").SetMap(map[string]any{
		"comment":      record.Comment,
		"start_time":   record.StartTime,
		"is_cancelled": record.IsCancelled,
		"updated_at":   record.UpdatedAt,
	}).Where("id = ?", record.ID).ToSql()
}

func (bpr *bookingPersistenceRepository) buildSelectByIDQueryForUpdate(id uuid.UUID) (string, []any, error) {
	return bpr.sqlBuilder.Select(
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

func (bpr *bookingPersistenceRepository) mapRows(rows *sql.Rows) ([]*bookingRecord, error) {
	var records []*bookingRecord
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

		records = append(records, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (bpr *bookingPersistenceRepository) mapRow(row *sql.Row) (*bookingRecord, error) {
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

	return &record, nil
}

func (bpr *bookingPersistenceRepository) mapCountRow(row *sql.Row) (int, error) {
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (bpr *bookingPersistenceRepository) buildRawCountQuery() squirrel.SelectBuilder {
	return bpr.sqlBuilder.Select("count(*)").From("bookings")
}

func (bpr *bookingPersistenceRepository) handleError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return err, true
	case errors.Is(err, sql.ErrNoRows):
		return booking.ErrBookingNotFound, true
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return booking.ErrBookingAlreadyExists, true
		case "23503":
			return booking.ErrInvalidOfferID, true
		}
	}

	return err, false
}
