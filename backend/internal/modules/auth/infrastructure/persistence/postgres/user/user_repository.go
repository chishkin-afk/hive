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

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type userRepository struct {
	log        *slog.Logger
	db         DB
	sqlBuilder squirrel.StatementBuilderType
}

func New(log *slog.Logger, db DB) *userRepository {
	return &userRepository{
		log:        log,
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save This is a function that saves the user to the database.
// It takes a domain model as input and returns an error
// for example, if the user already exists.
func (ur *userRepository) Save(ctx context.Context, user *user.User) error {
	ur.log.Debug("saving user",
		slog.String("user_id", user.ID().String()),
	)

	if err := ur.save(ctx, ur.db, user); err != nil {
		return fmt.Errorf("failed to save user: %w",
			handleError(err),
		)
	}

	return nil
}

func (ur *userRepository) save(ctx context.Context, db DB, user *user.User) error {
	record, err := toRecord(user)
	if err != nil {
		return err
	}

	query, args, err := ur.buildInsertQuery(record)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (ur *userRepository) buildInsertQuery(record *userRecord) (string, []any, error) {
	return ur.sqlBuilder.Insert("users").Columns(
		"id",
		"firstname",
		"lastname",
		"confirmed_email",
		"phone_number",
		"password_hash",
		"paid_until",
		"on_vacation",
		"tg_chat_id",
		"notification",
		"trial_used",
		"created_at",
		"updated_at",
	).Values(
		record.ID,
		record.Firstname,
		record.Lastname,
		record.ConfirmedEmail,
		record.PhoneNumber,
		record.PasswordHash,
		record.PaidUntil,
		record.OnVacation,
		record.TgChatID,
		record.Notification,
		record.TrialUsed,
		record.CreatedAt,
		record.UpdatedAt,
	).ToSql()
}

// GetByID is a function that returns a user
// from the database based on their ID, or
// an error otherwise.
func (ur *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	ur.log.Debug("getting user by id",
		slog.String("user_id", id.String()),
	)

	user, err := ur.getByID(ctx, ur.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w",
			handleError(err),
		)
	}

	return user, nil
}

func (ur *userRepository) getByID(ctx context.Context, db DB, id uuid.UUID) (*user.User, error) {
	query, args, err := ur.buildSelectQueryByID(id)
	if err != nil {
		return nil, err
	}

	return ur.mapRow(db.QueryRowContext(ctx, query, args...))
}

func (ur *userRepository) buildSelectQueryByID(id uuid.UUID) (string, []any, error) {
	return ur.buildSelectQuery().Where("id = ?", id).ToSql()
}

// GetByEmail is a function that retrieves a user
// from the database by their email;
// otherwise, it returns an error.
func (ur *userRepository) GetByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	ur.log.Debug("getting user by email",
		slog.String("email", email.String()),
	)

	user, err := ur.getByEmail(ctx, ur.db, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w",
			handleError(err),
		)
	}

	return user, nil
}

func (ur *userRepository) getByEmail(ctx context.Context, db DB, email user.Email) (*user.User, error) {
	query, args, err := ur.buildSelectQueryByEmail(email)
	if err != nil {
		return nil, err
	}

	return ur.mapRow(db.QueryRowContext(ctx, query, args...))
}

func (ur *userRepository) buildSelectQueryByEmail(email user.Email) (string, []any, error) {
	return ur.buildSelectQuery().Where("confirmed_email = ?", email.String()).ToSql()
}

func (ur *userRepository) buildSelectQuery() squirrel.SelectBuilder {
	return ur.sqlBuilder.Select(
		"id",
		"firstname",
		"lastname",
		"confirmed_email",
		"phone_number",
		"password_hash",
		"paid_until",
		"on_vacation",
		"tg_chat_id",
		"notification",
		"trial_used",
		"created_at",
		"updated_at",
	).From("users")
}

// Update is a function that implements
// the updating of user fields based
// on the specified domain model.
func (ur *userRepository) Update(ctx context.Context, updUser *user.User) error {
	ur.log.Debug("updating user by id",
		slog.String("id", updUser.ID().String()),
	)

	if err := ur.update(ctx, ur.db, updUser); err != nil {
		return fmt.Errorf("failed to update user: %w",
			handleError(err),
		)
	}

	return nil
}

func (ur *userRepository) update(ctx context.Context, db DB, updUser *user.User) error {
	record, err := toRecord(updUser)
	if err != nil {
		return err
	}

	query, args, err := ur.buildUpdateQuery(record)
	if err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

func (ur *userRepository) buildUpdateQuery(updUser *userRecord) (string, []any, error) {
	return ur.sqlBuilder.Update("users").SetMap(map[string]any{
		"firstname":     updUser.Firstname,
		"lastname":      updUser.Lastname,
		"phone_number":  updUser.PhoneNumber,
		"password_hash": updUser.PasswordHash,
		"on_vacation":   updUser.OnVacation,
		"tg_chat_id":    updUser.TgChatID,
		"notification":  updUser.Notification,
		"updated_at":    updUser.UpdatedAt,
	}).Where("id = ?", updUser.ID).ToSql()
}

// Lookup is a function that returns a list of users
// based on their IDs (subject to a limit),
// with a restricted set of fields.
func (ur *userRepository) Lookup(ctx context.Context, ids []uuid.UUID) ([]*user.User, error) {
	ur.log.Debug("lookuping users by ids",
		slog.Int("len_of_ids", len(ids)),
	)

	users, err := ur.lookup(ctx, ur.db, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup users: %w",
			handleError(err),
		)
	}

	return users, nil
}

func (ur *userRepository) lookup(ctx context.Context, db DB, ids []uuid.UUID) ([]*user.User, error) {
	query, args, err := ur.buildLookupUsers(ids)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return ur.mapLookupRows(rows)
}

func (ur *userRepository) buildLookupUsers(ids []uuid.UUID) (string, []any, error) {
	return ur.sqlBuilder.Select(
		"id",
		"firstname",
		"lastname",
		"confirmed_email",
	).From("users").Limit(1000).Where(squirrel.Eq{
		"id": ids,
	}).ToSql()
}

func (ur *userRepository) mapRow(row *sql.Row) (*user.User, error) {
	var record userRecord
	if err := row.Scan(
		&record.ID,
		&record.Firstname,
		&record.Lastname,
		&record.ConfirmedEmail,
		&record.PhoneNumber,
		&record.PasswordHash,
		&record.PaidUntil,
		&record.OnVacation,
		&record.TgChatID,
		&record.Notification,
		&record.TrialUsed,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return toDomain(&record)
}

func (ur *userRepository) mapLookupRows(rows *sql.Rows) ([]*user.User, error) {
	var users []*user.User
	for rows.Next() {
		var record userRecord
		if err := rows.Scan(
			&record.ID,
			&record.Firstname,
			&record.Lastname,
			&record.ConfirmedEmail,
		); err != nil {
			return nil, err
		}

		users = append(users, lookupToDomain(&record))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func handleError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return user.ErrUserNotFound
	}

	var pqError *pq.Error
	if errors.As(err, &pqError) {
		switch pqError.Code.Name() {
		case "unique_violation":
			return user.ErrUserAlreadyExists
		}
	}

	return err
}
