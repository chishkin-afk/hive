package paymentpg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/devathh/hive/backend/internal/modules/payments/domain/payment"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type paymentPersistenceRepository struct {
	log        *slog.Logger
	db         DB
	sqlBuilder squirrel.StatementBuilderType
}

func New(log *slog.Logger, db DB) *paymentPersistenceRepository {
	return &paymentPersistenceRepository{
		log:        log,
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (ppr *paymentPersistenceRepository) Save(ctx context.Context, payment *payment.Payment) (*payment.Payment, error) {
	ppr.log.Debug("converting payment into record", slog.String("payment_id", payment.ID().String()))
	record := paymentToRecord(payment)

	ppr.log.Debug("build insert query", slog.String("payment_id", payment.ID().String()))
	query, args, err := ppr.buildInsertQuery(record)
	if err != nil {
		return nil, fmt.Errorf("failed to build insert query: %w", err)
	}

	ppr.log.Debug("exec the insert query of payment", slog.String("query", query))
	if _, err := ppr.db.ExecContext(ctx, query, args...); err != nil {
		if safeErr, ok := ppr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to save payment into db: %w", err)
	}

	return recordToPayment(record), nil
}

func (ppr *paymentPersistenceRepository) buildInsertQuery(record *paymentRecord) (string, []any, error) {
	return ppr.sqlBuilder.Insert("payments_history").Columns(
		"id",
		"payment_id",
		"user_id",
		"amount_penny",
		"status_int",
		"created_at",
		"updated_at",
	).Values(
		record.ID,
		record.PaymentID,
		record.UserID,
		record.AmountPenny,
		record.StatusInt,
		record.CreatedAt,
		record.UpdatedAt,
	).ToSql()
}

func (ppr *paymentPersistenceRepository) GetByID(ctx context.Context, paymentID uuid.UUID) (*payment.Payment, error) {
	ppr.log.Debug("build the select query by id", slog.String("payment_id", paymentID.String()))
	query, args, err := ppr.buildSelectQueryByID(paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	ppr.log.Debug("exec the select by id query", slog.String("query", query))
	record, err := ppr.mapRow(ppr.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		if safeErr, ok := ppr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get payment by id from db: %w", err)
	}

	return recordToPayment(record), nil
}

func (ppr *paymentPersistenceRepository) buildSelectQueryByID(id uuid.UUID) (string, []any, error) {
	return ppr.sqlBuilder.Select(
		"id",
		"payment_id",
		"user_id",
		"amount_penny",
		"status_int",
		"created_at",
		"updated_at",
	).From("payments_history").Where("id = ?", id).ToSql()
}

func (ppr *paymentPersistenceRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*payment.Payment, int64, error) {
	ppr.log.Debug("build count query", slog.String("user_id", userID.String()))
	countQuery, countArgs, err := ppr.buildCountQuery(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build count query: %w", err)
	}

	count, err := ppr.mapCountRow(ppr.db.QueryRowContext(ctx, countQuery, countArgs...))
	if err != nil {
		if safeErr, ok := ppr.handleError(err); ok {
			return nil, 0, safeErr
		}

		return nil, 0, fmt.Errorf("failed to get count of all the rows: %w", err)
	}

	ppr.log.Debug("build list query by user's id", slog.String("user_id", userID.String()))
	listQuery, listArgs, err := ppr.buildListQuery(userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build list query: %w", err)
	}

	rows, err := ppr.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		if safeErr, ok := ppr.handleError(err); ok {
			return nil, 0, safeErr
		}

		return nil, 0, fmt.Errorf("failed to get list of payments: %w", err)
	}
	defer rows.Close()

	records, err := ppr.mapRows(rows)
	if err != nil {
		if safeErr, ok := ppr.handleError(err); ok {
			return nil, 0, safeErr
		}

		return nil, 0, fmt.Errorf("failed to map rows: %w", err)
	}

	payments := make([]*payment.Payment, len(records))
	for i, record := range records {
		payments[i] = recordToPayment(record)
	}

	return payments, count, nil
}

func (ppr *paymentPersistenceRepository) buildCountQuery(userID uuid.UUID) (string, []any, error) {
	return ppr.sqlBuilder.Select("count(*)").From("payments_history").Where("user_id", userID).ToSql()
}

func (ppr *paymentPersistenceRepository) buildListQuery(userID uuid.UUID, page, limit uint32) (string, []any, error) {
	offset := limit * (page - 1)
	return ppr.sqlBuilder.Select(
		"id",
		"payment_id",
		"user_id",
		"amount_penny",
		"status_int",
		"created_at",
		"updated_at",
	).From("payments_history").Limit(uint64(limit)).Offset(uint64(offset)).Where("user_id = ?", userID).ToSql()
}

// TODO: refactor all this payment
func (ppr *paymentPersistenceRepository) Update(ctx context.Context, paymentID, userID uuid.UUID, updFunc payment.UpdateFunc) (*payment.Payment, error) {
	ppr.log.Debug("start tx for update payment", slog.String("payment_id", paymentID.String()))
	tx, err := ppr.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start tx: %w", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				ppr.log.Error("failed to rollback tx", slog.String("error", err.Error()))
			}
		}
	}()

	ppr.log.Debug("build the select paid until query", slog.String("user_id", userID.String()))
	userQuery, userArgs, err := ppr.buildSelectPaidUntilQuery(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to build select paid until query: %w", err)
	}

	ppr.log.Debug("exec the select paid until query", slog.String("query", userQuery))
	paidUntil, err := ppr.mapPaidUntilRow(tx.QueryRowContext(ctx, userQuery, userArgs...))
	if err != nil {
		if safeErr, ok := ppr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get user's paid until field: %w", err)
	}

	paymentQuery, paymentArgs, err := ppr.buildSelectQueryForUpdate(paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	payment, err := ppr.mapRow(tx.QueryRowContext(ctx, paymentQuery, paymentArgs...))
	if err != nil {
		if safeErr, ok := ppr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get payment for update: %w", err)
	}

	updPayment := recordToPayment(payment)
	newPaidUntil, err := updFunc(ctx, paidUntil, updPayment)
	if err != nil {
		return nil, err
	}

	if paidUntil.Compare(newPaidUntil) != 0 {
		updateUserQuery, updateUserArgs, err := ppr.buildUpdatePaidUntilQuery(userID, newPaidUntil)
		if err != nil {
			return nil, fmt.Errorf("failed to build update user's paid until query: %w", err)
		}

		if _, err := tx.ExecContext(ctx, updateUserQuery, updateUserArgs...); err != nil {
			if safeErr, ok := ppr.handleError(err); ok {
				return nil, safeErr
			}

			return nil, fmt.Errorf("failed to update user's paid until: %w", err)
		}
	}

	updatePaymentQuery, updatePaymentArgs, err := ppr.buildUpdatePaymentQuery(paymentToRecord(updPayment))
	if err != nil {
		return nil, fmt.Errorf("failed to build update payment query: %w", err)
	}

	if _, err := tx.ExecContext(ctx, updatePaymentQuery, updatePaymentArgs...); err != nil {
		if safeErr, ok := ppr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		if safeErr, ok := ppr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	return updPayment, nil
}

func (ppr *paymentPersistenceRepository) buildUpdatePaymentQuery(updRecord *paymentRecord) (string, []any, error) {
	return ppr.sqlBuilder.Update("payments_history").SetMap(map[string]any{
		"status_int": updRecord.StatusInt,
		"updated_at": updRecord.UpdatedAt,
	}).Where("id = ?", updRecord.ID).ToSql()
}

func (ppr *paymentPersistenceRepository) buildUpdatePaidUntilQuery(userID uuid.UUID, paidUntil time.Time) (string, []any, error) {
	return ppr.sqlBuilder.Update("users").Set("paid_until", paidUntil).Where("id = ?", userID).ToSql()
}

func (ppr *paymentPersistenceRepository) buildSelectQueryForUpdate(id uuid.UUID) (string, []any, error) {
	return ppr.sqlBuilder.Select(
		"id",
		"payment_id",
		"user_id",
		"amount_penny",
		"status_int",
		"created_at",
		"updated_at",
	).From("payments_history").Where("id = ?", id).Suffix("FOR UPDATE").ToSql()
}

func (ppr *paymentPersistenceRepository) buildSelectPaidUntilQuery(userID uuid.UUID) (string, []any, error) {
	return ppr.sqlBuilder.Select("paid_until").From("users").Suffix("FOR UPDATE").Where("id = ?", userID).ToSql()
}

func (ppr *paymentPersistenceRepository) mapPaidUntilRow(row *sql.Row) (time.Time, error) {
	var paidUntil time.Time
	if err := row.Scan(&paidUntil); err != nil {
		return time.Time{}, err
	}

	if err := row.Err(); err != nil {
		return time.Time{}, err
	}

	return paidUntil, nil
}

func (ppr *paymentPersistenceRepository) mapRows(rows *sql.Rows) ([]*paymentRecord, error) {
	var records []*paymentRecord
	for rows.Next() {
		var record paymentRecord
		if err := rows.Scan(
			&record.ID,
			&record.PaymentID,
			&record.UserID,
			&record.AmountPenny,
			&record.StatusInt,
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

func (ppr *paymentPersistenceRepository) mapRow(row *sql.Row) (*paymentRecord, error) {
	var record paymentRecord
	if err := row.Scan(
		&record.ID,
		&record.PaymentID,
		&record.UserID,
		&record.AmountPenny,
		&record.StatusInt,
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

func (ppr *paymentPersistenceRepository) mapCountRow(row *sql.Row) (int64, error) {
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	if err := row.Err(); err != nil {
		return 0, err
	}

	return count, nil
}

func (ppr *paymentPersistenceRepository) handleError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return err, true
	case errors.Is(err, sql.ErrNoRows):
		return payment.ErrPaymentNotFound, true
	}

	var pqError *pq.Error
	if errors.As(err, &pqError) {
		switch pqError.Code {
		case "23503":
			return payment.ErrUserNotFound, true
		}
	}

	return err, false
}
