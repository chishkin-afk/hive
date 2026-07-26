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
}

type WithTx interface {
	DB
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type clientRepository struct {
	sqlBuilder squirrel.StatementBuilderType
	log        *slog.Logger
	db         WithTx
}

func New(log *slog.Logger, db WithTx) *clientRepository {
	return &clientRepository{
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		log:        log,
		db:         db,
	}
}

// Save is a method used to save an empty client
// (with one visit) to the database,
// or otherwise return an error.
func (cr *clientRepository) Save(ctx context.Context, client *client.Client) error {
	cr.log.Debug("saving client into db",
		slog.String("client_id", client.ID().String()),
	)

	if err := cr.save(ctx, cr.db, client); err != nil {
		return fmt.Errorf("failed to save client: %w",
			handleError(err),
		)
	}

	return nil
}

func (cr *clientRepository) save(ctx context.Context, db DB, client *client.Client) error {
	record := clientToRecord(client)

	query, args, err := cr.buildInsertQuery(record)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (cr *clientRepository) buildInsertQuery(record *clientRecord) (string, []any, error) {
	return cr.sqlBuilder.Insert("clients").Columns(
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

// Update is a method that updates a user's fields
// by calling an update function
// otherwise, it returns an error.
func (cr *clientRepository) Update(ctx context.Context, updFunc client.UpdateFunc, id uuid.UUID) (*client.Client, error) {
	cr.log.Debug("start tx")
	tx, rollback, err := cr.beginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	updClient, err := cr.getForUpdate(ctx, tx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for update: %w",
			handleError(err),
		)
	}

	if err := updFunc(updClient); err != nil {
		return nil, fmt.Errorf("failed to call update func: %w", err)
	}

	if err := cr.update(ctx, tx, updClient); err != nil {
		return nil, fmt.Errorf("failed to update client in db: %w",
			handleError(err),
		)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit update tx: %w", err)
	}

	return updClient, nil
}

func (cr *clientRepository) update(ctx context.Context, db DB, updClient *client.Client) error {
	record := clientToRecord(updClient)

	query, args, err := cr.buildUpdateQuery(record)
	if err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return client.ErrClientNotFound
	}

	return nil
}

func (cr *clientRepository) buildUpdateQuery(updRecord *clientRecord) (string, []any, error) {
	return cr.sqlBuilder.Update("clients").SetMap(map[string]any{
		"visits":     updRecord.Visits,
		"last_visit": updRecord.LastVisit,
	}).Where("id = ?", updRecord.ID).ToSql()
}

func (cr *clientRepository) getForUpdate(ctx context.Context, db DB, id uuid.UUID) (*client.Client, error) {
	query, args, err := cr.buildSelectForUpdateQuery(id)
	if err != nil {
		return nil, err
	}

	return mapRow(db.QueryRowContext(ctx, query, args...))
}

func (cr *clientRepository) buildSelectForUpdateQuery(id uuid.UUID) (string, []any, error) {
	return cr.sqlBuilder.Select(
		"id",
		"user_id",
		"offer_id",
		"visits",
		"last_visit",
		"created_at",
	).From("clients").Where("id = ?", id).Suffix("FOR UPDATE").ToSql()
}

// GetByID is a method that returns a client
// by its ID, or an error otherwise.
func (cr *clientRepository) GetByID(ctx context.Context, id uuid.UUID) (*client.Client, error) {
	cr.log.Debug("getting client by id",
		slog.String("client_id", id.String()),
	)

	client, err := cr.get(ctx, cr.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get client by id: %w",
			handleError(err),
		)
	}

	return client, nil
}

func (cr *clientRepository) get(ctx context.Context, db DB, id uuid.UUID) (*client.Client, error) {
	query, args, err := cr.buildSelectByIDQuery(id)
	if err != nil {
		return nil, err
	}

	return mapRow(db.QueryRowContext(ctx, query, args...))
}

func (cr *clientRepository) buildSelectByIDQuery(id uuid.UUID) (string, []any, error) {
	return cr.sqlBuilder.Select(
		"id",
		"user_id",
		"offer_id",
		"visits",
		"last_visit",
		"created_at",
	).From("clients").Where("id = ?", id).ToSql()
}

// GetByPairIDs is a method that returns a client
// based on their ID pair (userID + offerID),
// or an error otherwise.
func (cr *clientRepository) GetByPairIDs(ctx context.Context, userID, offerID uuid.UUID) (*client.Client, error) {
	cr.log.Debug("getting client by pair's of ids",
		slog.String("user_id", userID.String()),
		slog.String("offer_id", offerID.String()),
	)

	client, err := cr.getByPair(ctx, cr.db, userID, offerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client by pair's of ids: %w",
			handleError(err),
		)
	}

	return client, nil
}

func (cr *clientRepository) getByPair(ctx context.Context, db DB, userID, offerID uuid.UUID) (*client.Client, error) {
	query, args, err := cr.buildSelectByPairQuery(userID, offerID)
	if err != nil {
		return nil, err
	}

	return mapRow(db.QueryRowContext(ctx, query, args...))
}

func (cr *clientRepository) buildSelectByPairQuery(userID, offerID uuid.UUID) (string, []any, error) {
	return cr.sqlBuilder.Select(
		"id",
		"user_id",
		"offer_id",
		"visits",
		"last_visit",
		"created_at",
	).From("clients").Where("user_id = ? AND offer_id = ?", userID, offerID).ToSql()
}

// ListByOfferID is a method that returns
// a list of clients by offer ID,
// or an error otherwise.
func (cr *clientRepository) ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) ([]*client.Client, int64, error) {
	cr.log.Debug("counting all the clients by offer's id",
		slog.String("offer_id", offerID.String()),
	)

	count, err := cr.countByOfferID(ctx, cr.db, offerID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list clients by offer id: %w",
			handleError(err),
		)
	}

	cr.log.Debug("listing all the clients by offer's id",
		slog.String("offer_id", offerID.String()),
	)

	list, err := cr.listByOfferID(ctx, cr.db, offerID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list clients by offer's id: %w",
			handleError(err),
		)
	}

	return list, count, nil
}

func (cr *clientRepository) listByOfferID(ctx context.Context, db DB, offerID uuid.UUID, page, limit uint32) ([]*client.Client, error) {
	query, args, err := cr.buildListQueryByOfferID(offerID, page, limit)
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

func (cr *clientRepository) buildListQueryByOfferID(offerID uuid.UUID, page, limit uint32) (string, []any, error) {
	offset := limit * (page - 1)
	return cr.sqlBuilder.Select(
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

func (cr *clientRepository) countByOfferID(ctx context.Context, db DB, offerID uuid.UUID) (int64, error) {
	query, args, err := cr.buildCountByOfferIDQuery(offerID)
	if err != nil {
		return 0, err
	}

	return mapCountRow(db.QueryRowContext(ctx, query, args...))
}

func (cr *clientRepository) buildCountByOfferIDQuery(offerID uuid.UUID) (string, []any, error) {
	return cr.sqlBuilder.Select("count(*)").From("clients").Where("offer_id = ?", offerID).ToSql()
}

// ListByUserID is a method that returns
// a list of clients by user ID,
// or an error otherwise.
func (cr *clientRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*client.Client, int64, error) {
	cr.log.Debug("counting all the clients by user's id",
		slog.String("user_id", userID.String()),
	)

	count, err := cr.countByUserID(ctx, cr.db, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count all the clients by user's id: %w",
			handleError(err),
		)
	}

	cr.log.Debug("listing all the clients by user's id",
		slog.String("user_id", userID.String()),
	)

	clients, err := cr.listByUserID(ctx, cr.db, userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list clients by user's id: %w",
			handleError(err),
		)
	}

	return clients, count, nil
}

func (cr *clientRepository) listByUserID(ctx context.Context, db DB, userID uuid.UUID, page, limit uint32) ([]*client.Client, error) {
	query, args, err := cr.buildListByUserIDQuery(userID, page, limit)
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

func (cr *clientRepository) buildListByUserIDQuery(userID uuid.UUID, page, limit uint32) (string, []any, error) {
	offset := limit * (page - 1)
	return cr.sqlBuilder.Select(
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

func (cr *clientRepository) countByUserID(ctx context.Context, db DB, userID uuid.UUID) (int64, error) {
	query, args, err := cr.buildCountByUserIDQuery(userID)
	if err != nil {
		return 0, err
	}

	return mapCountRow(db.QueryRowContext(ctx, query, args...))
}

func (cr *clientRepository) buildCountByUserIDQuery(userID uuid.UUID) (string, []any, error) {
	return cr.sqlBuilder.Select("count(*)").From("clients").Where("user_id = ?", userID).ToSql()
}

// Delete`is a method that removes a client
// from the database by their ID
// otherwise, it returns an error.
func (cr *clientRepository) Delete(ctx context.Context, id uuid.UUID) error {
	cr.log.Debug("deleting client by id",
		slog.String("client_id", id.String()),
	)

	if err := cr.delete(ctx, cr.db, id); err != nil {
		return fmt.Errorf("failed to delete client from db: %w",
			handleError(err),
		)
	}

	return nil
}

func (cr *clientRepository) delete(ctx context.Context, db DB, id uuid.UUID) error {
	query, args, err := cr.buildDeleteQuery(id)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (cr *clientRepository) buildDeleteQuery(id uuid.UUID) (string, []any, error) {
	return cr.sqlBuilder.Delete("clients").Where("id = ?", id).ToSql()
}

func (cr *clientRepository) beginTx(ctx context.Context) (*sql.Tx, func(), error) {
	tx, err := cr.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, nil, err
	}

	return tx, func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			cr.log.Error("failed to rollback tx",
				slog.String("error", err.Error()),
			)
		}
	}, nil
}

func mapRow(row *sql.Row) (*client.Client, error) {
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

	return recordToClient(&record), nil
}

func mapCountRow(row *sql.Row) (int64, error) {
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func mapRows(rows *sql.Rows) ([]*client.Client, error) {
	var clients []*client.Client
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

		clients = append(clients, recordToClient(&record))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clients, nil
}

func handleError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return client.ErrClientNotFound
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return client.ErrClientAlreadyExists
		case "23503":
			return client.ErrInvalidOfferID
		}
	}

	return err
}
