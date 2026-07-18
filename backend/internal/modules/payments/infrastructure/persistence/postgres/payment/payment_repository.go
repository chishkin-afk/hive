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
}

type WithTx interface {
	DB
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type paymentRepository struct {
	log        *slog.Logger
	db         WithTx
	sqlBuilder squirrel.StatementBuilderType
}

func New(log *slog.Logger, db WithTx) *paymentRepository {
	return &paymentRepository{
		log:        log,
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save is a method that saves
// an unconfirmed payment
// to the database.
func (pr *paymentRepository) Save(ctx context.Context, payment *payment.Payment) error {
	pr.log.Debug("saving payment into db",
		slog.String("payment_id", payment.ID().String()),
	)

	if err := pr.save(ctx, pr.db, payment); err != nil {
		return fmt.Errorf("failed to save payment: %w",
			handleError(err),
		)
	}

	return nil
}

func (pr *paymentRepository) save(ctx context.Context, db DB, payment *payment.Payment) error {
	record := paymentToRecord(payment)

	query, args, err := pr.buildInsertQuery(record)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (pr *paymentRepository) buildInsertQuery(record *paymentRecord) (string, []any, error) {
	return pr.sqlBuilder.Insert("payments_history").Columns(
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

// GetByID executes a SELECT query
// against the database to
// retrieve a single payment by its ID.
func (pr *paymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*payment.Payment, error) {
	pr.log.Debug("getting payment by id",
		slog.String("payment_id", id.String()),
	)

	payment, err := pr.getByID(ctx, pr.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment by id: %w",
			handleError(err),
		)
	}

	return payment, nil
}

func (pr *paymentRepository) getByID(ctx context.Context, db DB, id uuid.UUID) (*payment.Payment, error) {
	query, args, err := pr.buildSelectByIDQuery(id)
	if err != nil {
		return nil, err
	}

	return mapRow(db.QueryRowContext(ctx, query, args...))
}

func (pr *paymentRepository) buildSelectByIDQuery(id uuid.UUID) (string, []any, error) {
	return pr.sqlBuilder.Select(
		"id",
		"payment_id",
		"user_id",
		"amount_penny",
		"status_int",
		"created_at",
		"updated_at",
	).From("payments_history").Where("id = ?", id).ToSql()
}

// GetByPaymentID executes a SELECT query
// against the database to retrieve
// a single payment record using the YooKassa payment ID.
func (pr *paymentRepository) GetByPaymentID(ctx context.Context, paymentID uuid.UUID) (*payment.Payment, error) {
	pr.log.Debug("getting payment by payment's id",
		slog.String("payment_id", paymentID.String()),
	)

	payment, err := pr.getByPaymentID(ctx, pr.db, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment by payment id: %w",
			handleError(err),
		)
	}

	return payment, nil
}

func (pr *paymentRepository) getByPaymentID(ctx context.Context, db DB, paymentID uuid.UUID) (*payment.Payment, error) {
	query, args, err := pr.buildSelectByPaymentIDQuery(paymentID)
	if err != nil {
		return nil, err
	}

	return mapRow(db.QueryRowContext(ctx, query, args...))
}

func (pr *paymentRepository) buildSelectByPaymentIDQuery(id uuid.UUID) (string, []any, error) {
	return pr.sqlBuilder.Select(
		"id",
		"payment_id",
		"user_id",
		"amount_penny",
		"status_int",
		"created_at",
		"updated_at",
	).From("payments_history").Where("payment_id = ?", id).ToSql()
}

// ListByUserID executes a SELECT query
// against the database to retrieve
// multiple payments based on the user (author) ID.
func (pr *paymentRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*payment.Payment, int64, error) {
	pr.log.Debug("counting payments by user's id",
		slog.String("user_id", userID.String()),
	)

	count, err := pr.countByUserID(ctx, pr.db, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count payments by user's id: %w",
			handleError(err),
		)
	}

	pr.log.Debug("listing payments by user's id",
		slog.String("user_id", userID.String()),
	)

	payments, err := pr.listByUserID(ctx, pr.db, userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list payments by user's id: %w",
			handleError(err),
		)
	}

	return payments, count, nil
}

func (pr *paymentRepository) listByUserID(ctx context.Context, db DB, userID uuid.UUID, page, limit uint32) ([]*payment.Payment, error) {
	query, args, err := pr.buildListQuery(userID, page, limit)
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

func (pr *paymentRepository) buildListQuery(userID uuid.UUID, page, limit uint32) (string, []any, error) {
	offset := limit * (page - 1)
	return pr.sqlBuilder.Select(
		"id",
		"payment_id",
		"user_id",
		"amount_penny",
		"status_int",
		"created_at",
		"updated_at",
	).From("payments_history").Limit(uint64(limit)).Offset(uint64(offset)).Where("user_id = ?", userID).OrderBy("created_at DESC").ToSql()
}

func (pr *paymentRepository) countByUserID(ctx context.Context, db DB, userID uuid.UUID) (int64, error) {
	query, args, err := pr.buildCountByUserIDQuery(userID)
	if err != nil {
		return 0, err
	}

	return mapCountRow(db.QueryRowContext(ctx, query, args...))
}

func (pr *paymentRepository) buildCountByUserIDQuery(userID uuid.UUID) (string, []any, error) {
	return pr.sqlBuilder.Select("count(*)").From("payments_history").Where("user_id = ?", userID).ToSql()
}

// Update is a method that, under the hood,
// triggers an update request to
// modify payment fields in the database.
func (pr *paymentRepository) Update(ctx context.Context, paymentID, userID uuid.UUID, updFunc payment.UpdateFunc) (*payment.Payment, error) {
	pr.log.Debug("start tx")
	tx, rollback, err := pr.beginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start tx: %w", err)
	}
	defer rollback()

	paidUntil, err := pr.getPaidUntil(ctx, tx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get paid until: %w",
			handleError(err),
		)
	}

	updPayment, err := pr.getForUpdateByPaymentID(ctx, tx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment for update: %w",
			handleError(err),
		)
	}

	newPaidUntil, err := updFunc(ctx, paidUntil, updPayment)
	if err != nil {
		return nil, err
	}

	if paidUntil.Compare(newPaidUntil) != 0 {
		if err := pr.updatePaidUntil(ctx, tx, newPaidUntil, userID); err != nil {
			return nil, fmt.Errorf("failed to update paid until: %w",
				handleError(err),
			)
		}
	}

	if err := pr.update(ctx, tx, updPayment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w",
			handleError(err),
		)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	return updPayment, nil
}

func (pr *paymentRepository) update(ctx context.Context, db DB, updPayment *payment.Payment) error {
	record := paymentToRecord(updPayment)

	query, args, err := pr.buildUpdatePaymentQuery(record)
	if err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return payment.ErrPaymentNotFound
	}

	return nil
}

func (pr *paymentRepository) buildUpdatePaymentQuery(updRecord *paymentRecord) (string, []any, error) {
	return pr.sqlBuilder.Update("payments_history").SetMap(map[string]any{
		"status_int": updRecord.StatusInt,
		"updated_at": updRecord.UpdatedAt,
	}).Where("id = ?", updRecord.ID).ToSql()
}

func (pr *paymentRepository) updatePaidUntil(ctx context.Context, db DB, paidUntil time.Time, userID uuid.UUID) error {
	query, args, err := pr.buildUpdatePaidUntilQuery(userID, paidUntil)
	if err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return payment.ErrUserNotFound
	}

	return nil
}

func (pr *paymentRepository) buildUpdatePaidUntilQuery(userID uuid.UUID, paidUntil time.Time) (string, []any, error) {
	return pr.sqlBuilder.Update("users").Set("paid_until", paidUntil).Where("id = ?", userID).ToSql()
}

func (pr *paymentRepository) getForUpdateByPaymentID(ctx context.Context, db DB, id uuid.UUID) (*payment.Payment, error) {
	query, args, err := pr.buildSelectForUpdateQuery(id)
	if err != nil {
		return nil, err
	}

	return mapRow(db.QueryRowContext(ctx, query, args...))
}

func (pr *paymentRepository) buildSelectForUpdateQuery(id uuid.UUID) (string, []any, error) {
	return pr.sqlBuilder.Select(
		"id",
		"payment_id",
		"user_id",
		"amount_penny",
		"status_int",
		"created_at",
		"updated_at",
	).From("payments_history").Where("payment_id = ?", id).Suffix("FOR UPDATE").ToSql()
}

func (pr *paymentRepository) getPaidUntil(ctx context.Context, db DB, userID uuid.UUID) (time.Time, error) {
	query, args, err := pr.buildSelectPaidUntilQuery(userID)
	if err != nil {
		return time.Time{}, err
	}

	return mapPaidUntilRow(db.QueryRowContext(ctx, query, args...))
}

func (pr *paymentRepository) buildSelectPaidUntilQuery(userID uuid.UUID) (string, []any, error) {
	return pr.sqlBuilder.Select("paid_until").From("users").Suffix("FOR UPDATE").Where("id = ?", userID).ToSql()
}

// TakeTrial is a method that updates
// the user's trial fields in the database.
func (pr *paymentRepository) TakeTrial(ctx context.Context, userID uuid.UUID, updFunc payment.TakeTrialFunc) error {
	pr.log.Debug("start tx")
	tx, rollback, err := pr.beginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start tx: %w", err)
	}
	defer rollback()

	paidUntil, trialUsed, err := pr.getTrialOfUser(ctx, tx, userID)
	if err != nil {
		return fmt.Errorf("failed to get trial of user: %w",
			handleError(err),
		)
	}

	newPaidUntil, err := updFunc(ctx, paidUntil, trialUsed)
	if err != nil {
		return err
	}

	if err := pr.takeTrial(ctx, tx, userID, newPaidUntil); err != nil {
		return fmt.Errorf("failed to take trial: %w",
			handleError(err),
		)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	return nil
}

func (pr *paymentRepository) takeTrial(ctx context.Context, db DB, userID uuid.UUID, paidUntil time.Time) error {
	query, args, err := pr.buildTakeTrialQuery(userID, paidUntil, true)
	if err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return payment.ErrUserNotFound
	}

	return nil
}

func (pr *paymentRepository) buildTakeTrialQuery(userID uuid.UUID, paidUntil time.Time, trialUsed bool) (string, []any, error) {
	return pr.sqlBuilder.Update("users").SetMap(map[string]any{
		"paid_until": paidUntil,
		"trial_used": trialUsed,
	}).Where("id = ?", userID).ToSql()
}

func (pr *paymentRepository) getTrialOfUser(ctx context.Context, db DB, userID uuid.UUID) (time.Time, bool, error) {
	query, args, err := pr.buildSelectTrialQuery(userID)
	if err != nil {
		return time.Time{}, false, err
	}

	return mapTrialRow(db.QueryRowContext(ctx, query, args...))
}

func (pr *paymentRepository) buildSelectTrialQuery(userID uuid.UUID) (string, []any, error) {
	return pr.sqlBuilder.Select("paid_until", "trial_used").From("users").Suffix("FOR UPDATE").Where("id = ?", userID).ToSql()
}

func (pr *paymentRepository) beginTx(ctx context.Context) (*sql.Tx, func(), error) {
	tx, err := pr.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, nil, err
	}

	return tx, func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			pr.log.Error("failed to rollback tx",
				slog.String("error", err.Error()),
			)
		}
	}, nil
}

func mapPaidUntilRow(row *sql.Row) (time.Time, error) {
	var paidUntil time.Time
	if err := row.Scan(&paidUntil); err != nil {
		return time.Time{}, err
	}

	if err := row.Err(); err != nil {
		return time.Time{}, err
	}

	return paidUntil, nil
}

func mapTrialRow(row *sql.Row) (time.Time, bool, error) {
	var (
		paidUntil time.Time
		trialUsed bool
	)
	if err := row.Scan(
		&paidUntil,
		&trialUsed,
	); err != nil {
		return time.Time{}, false, err
	}

	if err := row.Err(); err != nil {
		return time.Time{}, false, err
	}

	return paidUntil, trialUsed, nil
}

func mapRows(rows *sql.Rows) ([]*payment.Payment, error) {
	var payments []*payment.Payment
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

		payments = append(payments, recordToPayment(&record))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}

func mapRow(row *sql.Row) (*payment.Payment, error) {
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

	return recordToPayment(&record), nil
}

func mapCountRow(row *sql.Row) (int64, error) {
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
	if errors.Is(err, sql.ErrNoRows) {
		return payment.ErrPaymentNotFound
	}

	var pqError *pq.Error
	if errors.As(err, &pqError) {
		switch pqError.Code {
		case "23503":
			return payment.ErrUserNotFound
		}
	}

	return err
}
