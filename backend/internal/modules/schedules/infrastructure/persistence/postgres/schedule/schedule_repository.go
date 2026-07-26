package schedulepg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/devathh/hive/backend/internal/modules/schedules/domain/schedule"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type scheduleRepository struct {
	log        *slog.Logger
	db         DB
	sqlBuilder squirrel.StatementBuilderType
}

func New(log *slog.Logger, db DB) *scheduleRepository {
	return &scheduleRepository{
		log:        log,
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save is a method, which internally executes an INSERT query
// to save the schedule to the database
// otherwise, it returns an error.
func (sr *scheduleRepository) Save(ctx context.Context, schedule *schedule.Schedule) error {
	sr.log.Debug("saving schedule into db",
		slog.String("schedule_id", schedule.ID().String()),
	)

	if err := sr.save(ctx, sr.db, schedule); err != nil {
		return fmt.Errorf("failed to save schedule into db: %w",
			handleError(err),
		)
	}

	return nil
}

func (sr *scheduleRepository) save(ctx context.Context, db DB, schedule *schedule.Schedule) error {
	record, err := scheduleToRecord(schedule)
	if err != nil {
		return err
	}

	query, args, err := sr.buildInsertQuery(record)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (sr *scheduleRepository) buildInsertQuery(record *scheduleRecord) (string, []any, error) {
	return sr.sqlBuilder.Insert("schedules").Columns(
		"id",
		"user_id",
		"day_of_week",
		"start_time",
		"end_time",
		"breaks",
	).Values(
		record.ID,
		record.UserID,
		record.DayOfWeek,
		record.StartTime,
		record.EndTime,
		record.Breaks,
	).ToSql()
}

// ListByUserID is a method that retrieves
// all schedules based on a user ID
// otherwise, it returns an error.
func (sr *scheduleRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*schedule.Schedule, error) {
	sr.log.Debug("listing schedules by user's id",
		slog.String("user_id", userID.String()),
	)

	list, err := sr.listByUserID(ctx, sr.db, userID)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (sr *scheduleRepository) listByUserID(ctx context.Context, db DB, userID uuid.UUID) ([]*schedule.Schedule, error) {
	query, args, err := sr.buildListQuery(userID)
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

func (sr *scheduleRepository) buildListQuery(userID uuid.UUID) (string, []any, error) {
	return sr.sqlBuilder.Select(
		"id",
		"user_id",
		"day_of_week",
		"start_time",
		"end_time",
		"breaks",
	).From("schedules").Limit(7).Where("user_id = ?", userID).ToSql()
}

// GetByUserID is a method that returns a schedule
// based on the user ID and day of the week,
// or an error otherwise.
func (sr *scheduleRepository) GetByUserID(ctx context.Context, userID uuid.UUID, dayOfWeek uint8) (*schedule.Schedule, error) {
	sr.log.Debug("getting schedule by user's id",
		slog.String("user_id", userID.String()),
		slog.Int("day_of_week", int(dayOfWeek)),
	)

	schedule, err := sr.getByUserID(ctx, sr.db, userID, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule by user's id: %w",
			handleError(err),
		)
	}

	return schedule, nil
}

func (sr *scheduleRepository) getByUserID(ctx context.Context, db DB, userID uuid.UUID, dayOfWeek uint8) (*schedule.Schedule, error) {
	query, args, err := sr.buildSelectByUserIDQuery(userID, dayOfWeek)
	if err != nil {
		return nil, err
	}

	return mapRow(db.QueryRowContext(ctx, query, args...))
}

func (sr *scheduleRepository) buildSelectByUserIDQuery(userID uuid.UUID, dayOfWeek uint8) (string, []any, error) {
	return sr.sqlBuilder.Select(
		"id",
		"user_id",
		"day_of_week",
		"start_time",
		"end_time",
		"breaks",
	).From("schedules").Where("user_id = ? AND day_of_week = ?", userID, dayOfWeek).ToSql()
}

// GetByID is a method that retrieves
// a schedule from the database using its ID,
// or an error otherwise
func (sr *scheduleRepository) GetByID(ctx context.Context, id uuid.UUID) (*schedule.Schedule, error) {
	sr.log.Debug("getting schedule by id",
		slog.String("schedule_id", id.String()),
	)

	schedule, err := sr.getByID(ctx, sr.db, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule by id: %w",
			handleError(err),
		)
	}

	return schedule, nil
}

func (sr *scheduleRepository) getByID(ctx context.Context, db DB, id uuid.UUID) (*schedule.Schedule, error) {
	query, args, err := sr.buildSelectByIDQuery(id)
	if err != nil {
		return nil, err
	}

	return mapRow(db.QueryRowContext(ctx, query, args...))
}

func (sr *scheduleRepository) buildSelectByIDQuery(id uuid.UUID) (string, []any, error) {
	return sr.sqlBuilder.Select(
		"id",
		"user_id",
		"day_of_week",
		"start_time",
		"end_time",
		"breaks",
	).From("schedules").Where("id = ?", id).ToSql()
}

// Update persists changes to an existing schedule.
//
// The schedule is matched by its ID, and its mutable fields are updated in the
// database. An error is returned if the update operation fails.
func (sr *scheduleRepository) Update(ctx context.Context, updSchedule *schedule.Schedule) error {
	sr.log.Debug("updating the fields of schedules by id",
		slog.String("schedule_id", updSchedule.ID().String()),
	)

	if err := sr.update(ctx, sr.db, updSchedule); err != nil {
		return fmt.Errorf("failed to update schedule: %w",
			handleError(err),
		)
	}

	return nil
}

func (sr *scheduleRepository) update(ctx context.Context, db DB, updSchedule *schedule.Schedule) error {
	record, err := scheduleToRecord(updSchedule)
	if err != nil {
		return err
	}

	query, args, err := sr.buildUpdateQuery(record)
	if err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return schedule.ErrScheduleNotFound
	}

	return nil
}

func (sr *scheduleRepository) buildUpdateQuery(updRecord *scheduleRecord) (string, []any, error) {
	return sr.sqlBuilder.Update("schedules").SetMap(map[string]any{
		"start_time": updRecord.StartTime,
		"end_time":   updRecord.EndTime,
		"breaks":     updRecord.Breaks,
	}).Where("id = ?", updRecord.ID).ToSql()
}

// Delete permanently removes the schedule with the specified ID.
//
// If no schedule with the given ID exists, the behavior depends on the
// underlying database implementation. An error is returned only if the
// delete operation fails.
func (sr *scheduleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	sr.log.Debug("deleting schedule by id",
		slog.String("schedule_id", id.String()),
	)

	if err := sr.delete(ctx, sr.db, id); err != nil {
		return fmt.Errorf("failed to delete schedule by id: %w",
			handleError(err),
		)
	}

	return nil
}

func (sr *scheduleRepository) delete(ctx context.Context, db DB, id uuid.UUID) error {
	query, args, err := sr.buildDeleteQuery(id)
	if err != nil {
		return err
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return schedule.ErrScheduleNotFound
	}

	return nil
}

func (sr *scheduleRepository) buildDeleteQuery(id uuid.UUID) (string, []any, error) {
	return sr.sqlBuilder.Delete("schedules").Where("id = ?", id).ToSql()
}

func handleError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return schedule.ErrScheduleNotFound
	}

	var pqError *pq.Error
	if errors.As(err, &pqError) {
		switch pqError.Code {
		case "23505":
			return schedule.ErrScheduleAlreadyExists
		case "23503":
			return schedule.ErrUserNotFound
		}
	}

	return err
}

func mapRows(rows *sql.Rows) ([]*schedule.Schedule, error) {
	var schedules []*schedule.Schedule
	for rows.Next() {
		var record scheduleRecord
		if err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.DayOfWeek,
			&record.StartTime,
			&record.EndTime,
			&record.Breaks,
		); err != nil {
			return nil, err
		}

		schedule, err := recordToSchedule(&record)
		if err != nil {
			return nil, err
		}

		schedules = append(schedules, schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return schedules, nil
}

func mapRow(row *sql.Row) (*schedule.Schedule, error) {
	var record scheduleRecord
	if err := row.Scan(
		&record.ID,
		&record.UserID,
		&record.DayOfWeek,
		&record.StartTime,
		&record.EndTime,
		&record.Breaks,
	); err != nil {
		return nil, err
	}

	if err := row.Err(); err != nil {
		return nil, err
	}

	return recordToSchedule(&record)
}
