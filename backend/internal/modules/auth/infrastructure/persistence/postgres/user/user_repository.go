package userpg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/devathh/hive/backend/internal/modules/auth/domain/user"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type userPersistenceRepository struct {
	sqlBuiler squirrel.StatementBuilderType
	log       *slog.Logger
	db        DB
}

func New(log *slog.Logger, db DB) *userPersistenceRepository {
	return &userPersistenceRepository{
		sqlBuiler: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		log:       log,
		db:        db,
	}
}

// Save is a persistence method
// that used to create user's record in db
func (upr *userPersistenceRepository) Save(ctx context.Context, user *user.User) (*user.User, error) {
	upr.log.Debug("start to convert domain into record", slog.String("user_id", user.ID().String()))
	record, err := toRecord(user)
	if err != nil {
		return nil, fmt.Errorf("failed to convert domain into record: %w", err)
	}

	upr.log.Debug("build the sql insert query", slog.String("user_id", user.ID().String()))
	query, args, err := upr.buildInsertQuery(record)
	if err != nil {
		return nil, fmt.Errorf("failed to build insert query: %w", err)
	}

	upr.log.Debug("executing the sql query", slog.String("user_id", user.ID().String()))
	if _, err := upr.db.ExecContext(ctx, query, args...); err != nil {
		if safeErr, ok := upr.normalizeError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to exec sql insert query: %w", err)
	}

	return toDomain(record)
}

// GetByID is a method
// that used to get user's record by id from db
func (upr *userPersistenceRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	upr.log.Debug("build the sql select query by user's id", slog.String("user_id", id.String()))
	query, args, err := upr.buildSelectQueryByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to build sql select query: %w", err)
	}

	upr.log.Debug("executing the sql query", slog.String("user_id", id.String()))
	row := upr.db.QueryRowContext(ctx, query, args...)

	upr.log.Debug("mapping the sql row", slog.String("user_id", id.String()))
	record, err := upr.mapRow(row)
	if err != nil {
		if safeErr, ok := upr.normalizeError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to map row: %w", err)
	}

	return toDomain(record)
}

// GetByEmail is a method
// that used to get user's record by email from db
func (upr *userPersistenceRepository) GetByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	upr.log.Debug("build the sql select query by user's email", slog.String("user_email", email.String()))
	query, args, err := upr.buildSelectQueryByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to build sql select query: %w", err)
	}

	upr.log.Debug("executing the sql query", slog.String("user_email", email.String()))
	row := upr.db.QueryRowContext(ctx, query, args...)

	upr.log.Debug("mapping the sql row", slog.String("user_email", email.String()))
	record, err := upr.mapRow(row)
	if err != nil {
		if safeErr, ok := upr.normalizeError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to map row: %w", err)
	}

	return toDomain(record)
}

// Update is a method
// that can be used to update the fields of user
// by id in db
func (upr *userPersistenceRepository) Update(ctx context.Context, updUser *user.User) (*user.User, error) {
	upr.log.Debug("converting domain into record", slog.String("user_id", updUser.ID().String()))
	record, err := toRecord(updUser)
	if err != nil {
		return nil, fmt.Errorf("failed to convert domain into record: %w", err)
	}

	upr.log.Debug("build the sql update query by user's id", slog.String("user_id", updUser.ID().String()))
	query, args, err := upr.buildUpdateQuery(record)
	if err != nil {
		return nil, fmt.Errorf("failed to build sql update query: %w", err)
	}

	upr.log.Debug("executing the sql query", slog.String("user_id", updUser.ID().String()))
	result, err := upr.db.ExecContext(ctx, query, args...)
	if err != nil {
		if safeErr, ok := upr.normalizeError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to exec sql update query: %w", err)
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return nil, ErrUserNotFound
	}

	return toDomain(record)
}

func (upr *userPersistenceRepository) mapRow(row *sql.Row) (*userRecord, error) {
	var record userRecord
	if err := row.Scan(
		&record.ID,
		&record.Firstname,
		&record.Lastname,
		&record.ConfirmedEmail,
		&record.PasswordHash,
		&record.PaidUntil,
		&record.OnVacation,
		&record.TgChatID,
		&record.Notification,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &record, nil
}

func (upr *userPersistenceRepository) normalizeError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return err, true
	case errors.Is(err, sql.ErrNoRows):
		return ErrUserNotFound, true
	}

	var pqError *pq.Error
	if errors.As(err, &pqError) {
		switch pqError.Code {
		case "23505":
			return ErrUserAlreadyExists, true
		}
	}

	return err, false
}

func (upr *userPersistenceRepository) buildUpdateQuery(updUser *userRecord) (string, []any, error) {
	return squirrel.Update("users").SetMap(map[string]any{
		"firstname":     updUser.Firstname,
		"lastname":      updUser.Lastname,
		"password_hash": updUser.PasswordHash,
		"on_vacation":   updUser.OnVacation,
		"tg_chat_id":    updUser.TgChatID,
		"notification":  updUser.Notification,
		"updated_at":    updUser.UpdatedAt,
	}).Where("id = ?", updUser.ID).ToSql()
}

func (upr *userPersistenceRepository) buildSelectQueryByID(id uuid.UUID) (string, []any, error) {
	return upr.buildSelectQuery().Where("id = ?", id).ToSql()
}

func (upr *userPersistenceRepository) buildSelectQueryByEmail(email user.Email) (string, []any, error) {
	return upr.buildSelectQuery().Where("confirmed_email = ?", email.String()).ToSql()
}

func (upr *userPersistenceRepository) buildSelectQuery() squirrel.SelectBuilder {
	return squirrel.Select(
		"id",
		"firstname",
		"lastname",
		"confirmed_email",
		"password_hash",
		"paid_until",
		"on_vacation",
		"tg_chat_id",
		"notification",
		"created_at",
		"updated_at",
	).From("users")
}

func (upr *userPersistenceRepository) buildInsertQuery(record *userRecord) (string, []any, error) {
	return squirrel.Insert("users").Columns(
		"id",
		"firstname",
		"lastname",
		"confirmed_email",
		"password_hash",
		"paid_until",
		"on_vacation",
		"tg_chat_id",
		"notification",
		"created_at",
		"updated_at",
	).Values(
		record.ID,
		record.Firstname,
		record.Lastname,
		record.ConfirmedEmail,
		record.PasswordHash,
		record.PaidUntil,
		record.OnVacation,
		record.TgChatID,
		record.Notification,
		record.CreatedAt,
		record.UpdatedAt,
	).ToSql()
}
