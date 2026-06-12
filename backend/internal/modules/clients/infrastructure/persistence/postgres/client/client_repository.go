package clientpg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/devathh/hive/backend/internal/modules/clients/domain/client"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type clientPersistenceRepository struct {
	sqlBuilder squirrel.StatementBuilderType
	log        *slog.Logger
	db         DB
}

func New(log *slog.Logger, db DB) *clientPersistenceRepository {
	return &clientPersistenceRepository{
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		log:        log,
		db:         db,
	}
}

func (cpr *clientPersistenceRepository) Save(ctx context.Context, client *client.Client) (*client.Client, error) {
	cpr.log.Debug("converting domain into record", slog.String("client_id", client.ID().String()))
	record := clientToRecord(client)

	cpr.log.Debug("build the sql insert query", slog.String("client_id", client.ID().String()))
	query, args, err := cpr.buildInsertQuery(record)
	if err != nil {
		return nil, fmt.Errorf("failed to build insert query: %w", err)
	}

	cpr.log.Debug("exec the sql insert query", slog.String("query", query))
	if _, err := cpr.db.ExecContext(ctx, query, args...); err != nil {
		if safeErr, ok := cpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to save client into db: %w", err)
	}

	return recordToClient(record), nil
}

func (cpr *clientPersistenceRepository) buildInsertQuery(record *clientRecord) (string, []any, error) {
	return cpr.sqlBuilder.Insert("clients").Columns(
		"id",
		"user_id",
		"offer_id",
		"visits",
		"last_visit",
		"created_at",
	).Values(
		record.ID,
		record.UserID,
		record.OfferID,
		record.Visits,
		record.LastVisit,
		record.CreatedAt,
	).ToSql()
}

func (cpr *clientPersistenceRepository) Update(ctx context.Context, updFunc client.UpdateFunc, id uuid.UUID) (*client.Client, error) {
	cpr.log.Debug("start transaction", slog.String("client_id", id.String()))
	tx, err := cpr.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start tx: %w", err)
	}

	defer func() {
		if err != nil {
			cpr.log.Debug("rollback transaction")
			if err := tx.Rollback(); err != nil {
				cpr.log.Error("failed to rollback tx", slog.String("error", err.Error()))
			}
		}
	}()

	record, err := cpr.txSelectForUpdate(ctx, tx, id)
	if err != nil {
		if safeErr, ok := cpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to select client for update: %w", err)
	}

	cpr.log.Debug("call the update function", slog.String("client_id", record.ID.String()))
	client := recordToClient(record)
	if err := updFunc(client); err != nil {
		return nil, fmt.Errorf("failed to call update func: %w", err)
	}

	updatedRecord, err := cpr.txUpdateClient(ctx, tx, clientToRecord(client))
	if err != nil {
		if safeErr, ok := cpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to update client: %w", err)
	}

	if err := tx.Commit(); err != nil {
		if safeErr, ok := cpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	return recordToClient(updatedRecord), nil
}

func (cpr *clientPersistenceRepository) txUpdateClient(ctx context.Context, tx *sql.Tx, updRecord *clientRecord) (*clientRecord, error) {
	cpr.log.Debug("start to build sql update query", slog.String("client_id", updRecord.ID.String()))
	updateQuery, updateArgs, err := cpr.buildUpdateQuery(updRecord)
	if err != nil {
		return nil, err
	}

	cpr.log.Debug("exec the update sql query", slog.String("query", updateQuery), slog.String("client_id", updRecord.ID.String()))
	result, err := tx.ExecContext(ctx, updateQuery, updateArgs...)
	if err != nil {
		return nil, err
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return nil, client.ErrClientNotFound
	}

	return updRecord, nil
}

func (cpr *clientPersistenceRepository) txSelectForUpdate(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*clientRecord, error) {
	cpr.log.Debug("getting the client for update", slog.String("client_id", id.String()))
	selectQuery, selectArgs, err := cpr.buildSelectQueryForUpdate(id)
	if err != nil {
		return nil, err
	}

	cpr.log.Debug("exec the select query", slog.String("client_id", id.String()))
	row := tx.QueryRowContext(ctx, selectQuery, selectArgs...)

	return cpr.mapRow(row)
}

func (cpr *clientPersistenceRepository) buildSelectQueryForUpdate(id uuid.UUID) (string, []any, error) {
	return cpr.sqlBuilder.Select(
		"id",
		"user_id",
		"offer_id",
		"visits",
		"last_visit",
		"created_at",
	).From("clients").Where("id = ?", id).Suffix("FOR UPDATE").ToSql()
}

func (cpr *clientPersistenceRepository) buildUpdateQuery(updRecord *clientRecord) (string, []any, error) {
	return cpr.sqlBuilder.Update("clients").SetMap(map[string]any{
		"visits":     updRecord.Visits,
		"last_visit": updRecord.LastVisit,
	}).Where("id = ?", updRecord.ID).ToSql()
}

func (cpr *clientPersistenceRepository) GetByID(ctx context.Context, id uuid.UUID) (*client.Client, error) {
	cpr.log.Debug("build the sql query by id", slog.String("client_id", id.String()))
	query, args, err := cpr.buildSelectQueryByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	cpr.log.Debug("exec the select query by id", slog.String("query", query))
	row := cpr.db.QueryRowContext(ctx, query, args...)

	cpr.log.Debug("map the row of client", slog.String("client_id", id.String()))
	record, err := cpr.mapRow(row)
	if err != nil {
		if safeErr, ok := cpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to map row: %w", err)
	}

	return recordToClient(record), nil
}

func (cpr *clientPersistenceRepository) buildSelectQueryByID(id uuid.UUID) (string, []any, error) {
	return cpr.sqlBuilder.Select(
		"id",
		"user_id",
		"offer_id",
		"visits",
		"last_visit",
		"created_at",
	).From("clients").Where("id = ?", id).ToSql()
}

func (cpr *clientPersistenceRepository) GetByPairIDs(ctx context.Context, userID, offerID uuid.UUID) (*client.Client, error) {
	cpr.log.Debug("build select query by pair of ids",
		slog.String("offer_id", offerID.String()),
		slog.String("user_id", userID.String()))
	query, args, err := cpr.buildSelectQueryByPair(userID, offerID)
	if err != nil {
		return nil, fmt.Errorf("failed to build select query by pair of ids: %w", err)
	}

	cpr.log.Debug("exec the sql select query", slog.String("query", query))
	row := cpr.db.QueryRowContext(ctx, query, args...)

	cpr.log.Debug("map the row of client")
	record, err := cpr.mapRow(row)
	if err != nil {
		if safeErr, ok := cpr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to map row: %w", err)
	}

	return recordToClient(record), nil
}

func (cpr *clientPersistenceRepository) buildSelectQueryByPair(userID, offerID uuid.UUID) (string, []any, error) {
	return cpr.sqlBuilder.Select(
		"id",
		"user_id",
		"offer_id",
		"visits",
		"last_visit",
		"created_at",
	).From("clients").Where("user_id = ? AND offer_id = ?", userID, offerID).ToSql()
}

func (cpr *clientPersistenceRepository) ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) ([]*client.Client, int64, error) {
	cpr.log.Debug("build the count query", slog.String("offer_id", offerID.String()))
	countQuery, countArgs, err := cpr.buildCountQueryByOfferID(offerID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build count query: %w", err)
	}

	cpr.log.Debug("exec the count query", slog.String("query", countQuery))
	count, err := cpr.count(ctx, countQuery, countArgs...)
	if err != nil {
		if safeErr, ok := cpr.handleError(err); ok {
			return nil, 0, safeErr
		}

		return nil, 0, fmt.Errorf("failed to count all rows: %w", err)
	}

	cpr.log.Debug("build the sql list query of clients", slog.String("offer_id", offerID.String()))
	listQuery, listArgs, err := cpr.buildListQueryByOfferID(offerID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build list query: %w", err)
	}

	cpr.log.Debug("exec the list query", slog.String("query", listQuery))
	rows, err := cpr.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		if safeErr, ok := cpr.handleError(err); ok {
			return nil, 0, safeErr
		}

		return nil, 0, fmt.Errorf("failed to get list of clients by offer id: %w", err)
	}
	defer rows.Close()

	records, err := cpr.mapRows(rows)
	if err != nil {
		if safeErr, ok := cpr.handleError(err); ok {
			return nil, 0, safeErr
		}

		return nil, 0, fmt.Errorf("failed to map rows: %w", err)
	}

	clients := make([]*client.Client, len(records))
	for idx, record := range records {
		clients[idx] = recordToClient(record)
	}

	return clients, int64(count), nil
}

func (cpr *clientPersistenceRepository) buildListQueryByOfferID(offerID uuid.UUID, page, limit uint32) (string, []any, error) {
	offset := limit * (page - 1)
	return cpr.sqlBuilder.Select(
		"id",
		"user_id",
		"offer_id",
		"visits",
		"last_visit",
		"created_at",
	).From("clients").OrderBy("created_at DESC").
		Offset(uint64(offset)).Limit(uint64(limit)).
		Where("offer_id = ?", offerID).ToSql()
}

func (cpr *clientPersistenceRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*client.Client, int64, error) {
	cpr.log.Debug("build the count query", slog.String("user_id", userID.String()))
	countQuery, countArgs, err := cpr.buildCountQueryByUserID(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build count query: %w", err)
	}

	cpr.log.Debug("exec the count query", slog.String("query", countQuery))
	count, err := cpr.count(ctx, countQuery, countArgs...)
	if err != nil {
		if safeErr, ok := cpr.handleError(err); ok {
			return nil, 0, safeErr
		}

		return nil, 0, fmt.Errorf("failed to count all rows: %w", err)
	}

	cpr.log.Debug("build the sql list query of clients", slog.String("user_id", userID.String()))
	listQuery, listArgs, err := cpr.buildListQueryByUserID(userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build list query: %w", err)
	}

	cpr.log.Debug("exec the list query", slog.String("query", listQuery))
	rows, err := cpr.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		if safeErr, ok := cpr.handleError(err); ok {
			return nil, 0, safeErr
		}

		return nil, 0, fmt.Errorf("failed to get list of clients by user id: %w", err)
	}
	defer rows.Close()

	records, err := cpr.mapRows(rows)
	if err != nil {
		if safeErr, ok := cpr.handleError(err); ok {
			return nil, 0, safeErr
		}

		return nil, 0, fmt.Errorf("failed to map rows: %w", err)
	}

	clients := make([]*client.Client, len(records))
	for idx, record := range records {
		clients[idx] = recordToClient(record)
	}

	return clients, int64(count), nil
}

func (cpr *clientPersistenceRepository) buildListQueryByUserID(userID uuid.UUID, page, limit uint32) (string, []any, error) {
	offset := limit * (page - 1)
	return cpr.sqlBuilder.Select(
		"id",
		"user_id",
		"offer_id",
		"visits",
		"last_visit",
		"created_at",
	).From("clients").OrderBy("created_at DESC").
		Offset(uint64(offset)).Limit(uint64(limit)).
		Where("user_id = ?", userID).ToSql()
}

func (cpr *clientPersistenceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	cpr.log.Debug("build the sql delete query", slog.String("client_id", id.String()))
	query, args, err := cpr.buildDeleteQuery(id)
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	cpr.log.Debug("exec the sql delete query", slog.String("query", query))
	result, err := cpr.db.ExecContext(ctx, query, args...)
	if err != nil {
		if safeErr, ok := cpr.handleError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to delete client from db: %w", err)
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return client.ErrClientNotFound
	}

	return nil
}

func (cpr *clientPersistenceRepository) buildDeleteQuery(id uuid.UUID) (string, []any, error) {
	return cpr.sqlBuilder.Delete("clients").Where("id = ?", id).ToSql()
}

func (cpr *clientPersistenceRepository) count(ctx context.Context, query string, args ...any) (uint64, error) {
	var count uint64
	if err := cpr.db.QueryRowContext(ctx, query, args...).Scan(
		&count,
	); err != nil {
		return 0, err
	}

	return count, nil
}

func (cpr *clientPersistenceRepository) buildCountQueryByOfferID(offerID uuid.UUID) (string, []any, error) {
	return cpr.sqlBuilder.Select("count(*)").From("clients").Where("offer_id = ?", offerID).ToSql()
}

func (cpr *clientPersistenceRepository) buildCountQueryByUserID(userID uuid.UUID) (string, []any, error) {
	return cpr.sqlBuilder.Select("count(*)").From("clients").Where("user_id = ?", userID).ToSql()
}

func (cpr *clientPersistenceRepository) handleError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return err, true
	case errors.Is(err, sql.ErrNoRows):
		return client.ErrClientNotFound, true
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return client.ErrClientAlreadyExists, true
		case "23503":
			return client.ErrInvalidOfferID, true
		}
	}

	return err, false
}

func (cpr *clientPersistenceRepository) mapRow(row *sql.Row) (*clientRecord, error) {
	var record clientRecord
	if err := row.Scan(
		&record.ID,
		&record.UserID,
		&record.OfferID,
		&record.Visits,
		&record.LastVisit,
		&record.CreatedAt,
	); err != nil {
		return nil, err
	}

	if err := row.Err(); err != nil {
		return nil, err
	}

	return &record, nil
}

func (cpr *clientPersistenceRepository) mapRows(rows *sql.Rows) ([]*clientRecord, error) {
	var records []*clientRecord
	for rows.Next() {
		var record clientRecord
		if err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.OfferID,
			&record.Visits,
			&record.LastVisit,
			&record.CreatedAt,
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
