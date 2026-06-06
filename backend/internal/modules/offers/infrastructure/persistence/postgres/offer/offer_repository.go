package offerpg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

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

type offerPersistenceRepository struct {
	log        *slog.Logger
	db         DB
	sqlBuilder squirrel.StatementBuilderType
}

func New(log *slog.Logger, db DB) *offerPersistenceRepository {
	return &offerPersistenceRepository{
		log:        log,
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (opr *offerPersistenceRepository) Save(ctx context.Context, offer *offer.Offer) (*offer.Offer, error) {
	opr.log.Debug("converting domain into record", slog.String("offer_id", offer.ID().String()))
	record := domainToRecord(offer)

	opr.log.Debug("build the sql insert query")
	query, args, err := opr.buildInsertQuery(record)
	if err != nil {
		return nil, fmt.Errorf("failed to build sql insert query: %w", err)
	}

	opr.log.Debug("exec the sql insert query", slog.String("offer_id", offer.ID().String()))
	if _, err := opr.db.ExecContext(ctx, query, args...); err != nil {
		if safeErr, ok := opr.normalizeError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to save offer into db: %w", err)
	}

	return recordToDomain(record), nil
}

func (opr *offerPersistenceRepository) buildInsertQuery(record *offerRecord) (string, []any, error) {
	return opr.sqlBuilder.Insert("offers").Columns(
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

func (opr *offerPersistenceRepository) Update(ctx context.Context, updOffer *offer.Offer) (*offer.Offer, error) {
	opr.log.Debug("converting upd domain into record", slog.String("offer_id", updOffer.ID().String()))
	record := domainToRecord(updOffer)

	opr.log.Debug("build the sql update query")
	query, args, err := opr.buildUpdateQuery(record)
	if err != nil {
		return nil, fmt.Errorf("failed to build sql update query: %w", err)
	}

	opr.log.Debug("exec the sql update query", slog.String("offer_id", updOffer.ID().String()))
	result, err := opr.db.ExecContext(ctx, query, args...)
	if err != nil {
		if safeErr, ok := opr.normalizeError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to update offer in db: %w", err)
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return nil, offer.ErrOfferNotFound
	}

	return recordToDomain(record), nil
}

func (opr *offerPersistenceRepository) buildUpdateQuery(record *offerRecord) (string, []any, error) {
	return opr.sqlBuilder.Update("offers").SetMap(map[string]any{
		"title":         record.Title,
		"description":   record.Description,
		"price":         record.Price,
		"slot_duration": record.SlotDurationMinutes,
		"slot_interval": record.SlotIntervalMinutes,
		"is_active":     record.IsActive,
		"updated_at":    record.UpdatedAt,
	}).Where("id = ?", record.ID).ToSql()
}

func (opr *offerPersistenceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	opr.log.Debug("build the sql delete query", slog.String("offer_id", id.String()))
	query, args, err := opr.buildDeleteQuery(id)
	if err != nil {
		return fmt.Errorf("failed to build sql delete query: %w", err)
	}

	opr.log.Debug("exec the sql delete query", slog.String("offer_id", id.String()))
	result, err := opr.db.ExecContext(ctx, query, args...)
	if err != nil {
		if safeErr, ok := opr.normalizeError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to delete offer: %w", err)
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return offer.ErrOfferNotFound
	}

	return nil
}

func (opr *offerPersistenceRepository) buildDeleteQuery(id uuid.UUID) (string, []any, error) {
	return opr.sqlBuilder.Delete("offers").Where("id = ?", id).ToSql()
}

func (opr *offerPersistenceRepository) GetByID(ctx context.Context, id uuid.UUID) (*offer.Offer, error) {
	opr.log.Debug("build the sql select query by id", slog.String("offer_id", id.String()))
	query, args, err := opr.buildSelectQuery(id)
	if err != nil {
		return nil, fmt.Errorf("failed to build sql select query: %w", err)
	}

	opr.log.Debug("exec the sql select query by id", slog.String("offer_id", id.String()))
	row := opr.db.QueryRowContext(ctx, query, args...)
	if err := row.Err(); err != nil {
		if safeErr, ok := opr.normalizeError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get offer by id: %w", err)
	}

	record, err := opr.mapRow(row)
	if err != nil {
		return nil, fmt.Errorf("failed to map row: %w", err)
	}

	return recordToDomain(record), nil
}

func (opr *offerPersistenceRepository) buildSelectQuery(id uuid.UUID) (string, []any, error) {
	return opr.sqlBuilder.Select(
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
	).From("offers").Where("id = ?", id).ToSql()
}

func (opr *offerPersistenceRepository) ListByAuthorID(ctx context.Context, authorID uuid.UUID, page, limit uint32) ([]*offer.Offer, uint64, error) {
	opr.log.Debug("build the sql count query", slog.String("author_id", authorID.String()))
	countQuery, countArgs, err := opr.buildCountQuery(authorID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build count query: %w", err)
	}

	totalCount, err := opr.count(ctx, countQuery, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count of offers: %w", err)
	}

	opr.log.Debug("build the sql list query", slog.String("author_id", authorID.String()))
	listQuery, listArgs, err := opr.buildListQuery(authorID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build sql list query: %w", err)
	}

	opr.log.Debug("exec the sql list query", slog.String("author_id", authorID.String()))
	rows, err := opr.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		if safeErr, ok := opr.normalizeError(err); ok {
			return nil, 0, safeErr
		}

		return nil, 0, fmt.Errorf("failed to get list of offers: %w", err)
	}

	records, err := opr.mapRows(rows)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to map rows: %w", err)
	}

	result := make([]*offer.Offer, len(records))
	for idx, record := range records {
		result[idx] = recordToDomain(record)
	}

	return result, totalCount, nil
}

func (opr *offerPersistenceRepository) count(ctx context.Context, query string, args ...any) (uint64, error) {
	var count uint64
	if err := opr.db.QueryRowContext(ctx, query, args...).Scan(
		&count,
	); err != nil {
		return 0, err
	}

	return count, nil
}

func (opr *offerPersistenceRepository) buildCountQuery(authorID uuid.UUID) (string, []any, error) {
	return opr.sqlBuilder.Select("count(*)").From("offers").Where("author_id = ?", authorID).ToSql()
}

func (opr *offerPersistenceRepository) buildListQuery(authorID uuid.UUID, page, limit uint32) (string, []any, error) {
	offset := limit * (page - 1)
	return opr.sqlBuilder.Select(
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
	).From("offers").Offset(uint64(offset)).Limit(uint64(limit)).
		Where("author_id = ?", authorID).ToSql()
}

func (opr *offerPersistenceRepository) mapRow(row *sql.Row) (*offerRecord, error) {
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

	return &record, nil
}

func (opr *offerPersistenceRepository) mapRows(rows *sql.Rows) ([]*offerRecord, error) {
	var records []*offerRecord
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

		records = append(records, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (opr *offerPersistenceRepository) normalizeError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return err, true
	case errors.Is(err, sql.ErrNoRows):
		return offer.ErrOfferNotFound, true
	}

	var pqError *pq.Error
	if errors.As(err, &pqError) {
		switch pqError.Code {
		case "23503":
			return offer.ErrAuthorNotFound, true
		}
	}

	return err, false
}
