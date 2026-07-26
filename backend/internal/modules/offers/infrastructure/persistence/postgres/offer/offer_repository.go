package offerpg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/devathh/hive/backend/internal/modules/offers/domain/offer"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type offerRepository struct {
	log        *slog.Logger
	db         DB
	sqlBuilder squirrel.StatementBuilderType
}

func New(log *slog.Logger, db DB) *offerRepository {
	return &offerRepository{
		log:        log,
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save is a method that saves the offer to the database
// otherwise, it returns a safe error.
func (or *offerRepository) Save(ctx context.Context, offer *offer.Offer) error {
	or.log.Debug("saving offer into db",
		slog.String("offer_id", offer.ID().String()),
	)

	if err := or.save(ctx, or.db, offer); err != nil {
		return fmt.Errorf("failed to save offer into db: %w",
			handleError(err),
		)
	}

	return nil
}

func (or *offerRepository) save(ctx context.Context, db DB, offer *offer.Offer) error {
	record := domainToRecord(offer)
	query, args, err := or.buildInsertQuery(record)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (or *offerRepository) buildInsertQuery(record *offerRecord) (string, []any, error) {
	return or.sqlBuilder.Insert("offers").Columns(
		"id",
		"author_id",
		"title",
		"description",
		"price",
		"slot_duration_minutes",
		"slot_interval_minutes",
		"is_active",
		"created_at",
		"updated_at",
	).Values(
		record.ID,
		record.AuthorID,
		record.Title,
		record.Description,
		record.Price,
		record.SlotDurationMinutes,
		record.SlotIntervalMinutes,
		record.IsActive,
		record.CreatedAt,
		record.UpdatedAt,
	).ToSql()
}

// Update is a method that updates
// the mutable fields of an offer in the database
// otherwise, it returns an error.
func (or *offerRepository) Update(ctx context.Context, updOffer *offer.Offer) error {
	or.log.Debug("updating offer by id",
		slog.String("offer_id", updOffer.ID().String()),
	)

	if err := or.update(ctx, or.db, updOffer); err != nil {
		return fmt.Errorf("failed to update offer by id: %w",
			handleError(err),
		)
	}

	return nil
}

func (or *offerRepository) update(ctx context.Context, db DB, updOffer *offer.Offer) error {
	record := domainToRecord(updOffer)
	query, args, err := or.buildUpdateQuery(record)
	if err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return offer.ErrOfferNotFound
	}

	return nil
}

func (or *offerRepository) buildUpdateQuery(record *offerRecord) (string, []any, error) {
	return or.sqlBuilder.Update("offers").SetMap(map[string]any{
		"title":                 record.Title,
		"description":           record.Description,
		"price":                 record.Price,
		"slot_duration_minutes": record.SlotDurationMinutes,
		"slot_interval_minutes": record.SlotIntervalMinutes,
		"is_active":             record.IsActive,
		"updated_at":            record.UpdatedAt,
	}).Where("id = ?", record.ID).ToSql()
}

// Delete is a method that
// removes an offer by ID.
func (or *offerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	or.log.Debug("deleting offer from db",
		slog.String("offer_id", id.String()),
	)

	if err := or.delete(ctx, or.db, id); err != nil {
		return fmt.Errorf("failed to delete offer from db: %w",
			handleError(err),
		)
	}

	return nil
}

func (or *offerRepository) delete(ctx context.Context, db DB, id uuid.UUID) error {
	query, args, err := or.buildDeleteQuery(id)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (or *offerRepository) buildDeleteQuery(id uuid.UUID) (string, []any, error) {
	return or.sqlBuilder.Update("offers").Set(
		"deleted_at", time.Now().UTC(),
	).Where("id = ?", id).ToSql()
}

// Get is a method that returns an offer
// from the db by its ID,
// or otherwise returns an error.
func (or *offerRepository) GetByID(ctx context.Context, id uuid.UUID) (*offer.Offer, error) {
	or.log.Debug("getting offer by id",
		slog.String("offer_id", id.String()),
	)

	offer, err := or.get(ctx, or.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get offer by id: %w",
			handleError(err),
		)
	}

	return offer, nil
}

func (or *offerRepository) get(ctx context.Context, db DB, id uuid.UUID) (*offer.Offer, error) {
	query, args, err := or.buildSelectQuery(id)
	if err != nil {
		return nil, err
	}

	return mapRow(db.QueryRowContext(ctx, query, args...))
}

func (or *offerRepository) buildSelectQuery(id uuid.UUID) (string, []any, error) {
	return or.sqlBuilder.Select(
		"id",
		"author_id",
		"title",
		"description",
		"price",
		"slot_duration_minutes",
		"slot_interval_minutes",
		"is_active",
		"created_at",
		"updated_at",
	).From("offers").Where("id = ? AND deleted_at is NULL", id).ToSql()
}

// ListByAuthorID is a method that
// returns paginated offers by author ID,
// or an error otherwise.
func (or *offerRepository) ListByAuthorID(ctx context.Context, authorID uuid.UUID, page, limit uint32) ([]*offer.Offer, int64, error) {
	or.log.Debug("counting all the offers by author id",
		slog.String("author_id", authorID.String()),
	)

	count, err := or.countByAuthorID(ctx, or.db, authorID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count by author id: %w",
			handleError(err),
		)
	}

	or.log.Debug("listing offers by author id",
		slog.String("author_id", authorID.String()),
	)

	list, err := or.listByAuthorID(ctx, or.db, authorID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list offers by author id: %w",
			handleError(err),
		)
	}

	return list, count, nil
}

func (or *offerRepository) listByAuthorID(ctx context.Context, db DB, authorID uuid.UUID, page, limit uint32) ([]*offer.Offer, error) {
	query, args, err := or.buildListQuery(authorID, page, limit)
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

func (or *offerRepository) buildListQuery(authorID uuid.UUID, page, limit uint32) (string, []any, error) {
	offset := limit * (page - 1)
	return or.sqlBuilder.Select(
		"id",
		"author_id",
		"title",
		"description",
		"price",
		"slot_duration_minutes",
		"slot_interval_minutes",
		"is_active",
		"created_at",
		"updated_at",
	).From("offers").OrderBy("created_at DESC").Offset(uint64(offset)).Limit(uint64(limit)).
		Where("author_id = ? AND deleted_at is NULL", authorID).ToSql()
}

func (or *offerRepository) countByAuthorID(ctx context.Context, db DB, authorID uuid.UUID) (int64, error) {
	query, args, err := or.buildCountQuery(authorID)
	if err != nil {
		return 0, err
	}

	return mapCountRow(db.QueryRowContext(ctx, query, args...))
}

func (or *offerRepository) buildCountQuery(authorID uuid.UUID) (string, []any, error) {
	return or.sqlBuilder.Select("count(*)").From("offers").Where("author_id = ? AND deleted_at is NULL", authorID).ToSql()
}

func mapCountRow(row *sql.Row) (int64, error) {
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func mapRow(row *sql.Row) (*offer.Offer, error) {
	var record offerRecord
	if err := row.Scan(
		&record.ID,
		&record.AuthorID,
		&record.Title,
		&record.Description,
		&record.Price,
		&record.SlotDurationMinutes,
		&record.SlotIntervalMinutes,
		&record.IsActive,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return recordToDomain(&record), nil
}

func mapRows(rows *sql.Rows) ([]*offer.Offer, error) {
	var offers []*offer.Offer
	for rows.Next() {
		var record offerRecord
		if err := rows.Scan(
			&record.ID,
			&record.AuthorID,
			&record.Title,
			&record.Description,
			&record.Price,
			&record.SlotDurationMinutes,
			&record.SlotIntervalMinutes,
			&record.IsActive,
			&record.CreatedAt,
			&record.UpdatedAt,
		); err != nil {
			return nil, err
		}

		offers = append(offers, recordToDomain(&record))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return offers, nil
}

func handleError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return offer.ErrOfferNotFound
	}

	var pqError *pq.Error
	if errors.As(err, &pqError) {
		switch pqError.Code {
		case "23503":
			return offer.ErrAuthorNotFound
		}
	}

	return err
}
