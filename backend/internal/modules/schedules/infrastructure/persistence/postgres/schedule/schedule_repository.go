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

type schedulePersistenceRepository struct {
	log        *slog.Logger
	db         DB
	sqlBuilder squirrel.StatementBuilderType
}

func New(log *slog.Logger, db DB) *schedulePersistenceRepository {
	return &schedulePersistenceRepository{
		log:        log,
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (spr *schedulePersistenceRepository) Save(ctx context.Context, schedule *schedule.Schedule) (*schedule.Schedule, error) {
	spr.log.Debug("converting domain model to record", slog.String("schedule_id", schedule.ID().String()))
	record, err := scheduleToRecord(schedule)
	if err != nil {
		return nil, fmt.Errorf("failed to convert schedule: %w", err)
	}

	spr.log.Debug("build the sql insert query", slog.String("schedule_id", record.ID.String()))
	query, args, err := spr.buildInsertQuery(record)
	if err != nil {
		return nil, fmt.Errorf("failed to build insert query: %w", err)
	}

	spr.log.Debug("exec the insert query", slog.String("query", query))
	if _, err := spr.db.ExecContext(ctx, query, args); err != nil {
		if safeErr, ok := spr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to save user into db: %w", err)
	}

	return recordToSchedule(record)
}

func (spr *schedulePersistenceRepository) buildInsertQuery(record *scheduleRecord) (string, []any, error) {
	return spr.sqlBuilder.Insert("schedules").Columns(
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

func (spr *schedulePersistenceRepository) GetSchedulesByUserID(ctx context.Context, userID uuid.UUID) ([]*schedule.Schedule, error) {
	spr.log.Debug("build the list query by user's id", slog.String("user_id", userID.String()))
	query, args, err := spr.buildListQuery(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to build list query: %w", err)
	}

	spr.log.Debug("exec the list query", slog.String("query", query))
	rows, err := spr.db.QueryContext(ctx, query, args)
	if err != nil {
		if safeErr, ok := spr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get list of schedules: %w", err)
	}
	defer rows.Close()

	spr.log.Debug("map the rows of schedules", slog.String("user_id", userID.String()))
	records, err := spr.mapRows(rows)
	if err != nil {
		if safeErr, ok := spr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to map list of schedules: %w", err)
	}

	result := make([]*schedule.Schedule, len(records))
	for i, record := range records {
		schedule, err := recordToSchedule(&record)
		if err != nil {
			return nil, err
		}

		result[i] = schedule
	}

	return result, nil
}

func (spr *schedulePersistenceRepository) buildListQuery(userID uuid.UUID) (string, []any, error) {
	return spr.sqlBuilder.Select(
		"id",
		"user_id",
		"day_of_week",
		"start_time",
		"end_time",
		"breaks",
	).From("schedules").Where("user_id = ?", userID).ToSql()
}

func (spr *schedulePersistenceRepository) GetScheduleByUserID(ctx context.Context, userID uuid.UUID, dayOfWeek uint8) (*schedule.Schedule, error) {
	spr.log.Debug("build the sql select query by user's id and day_of_week",
		slog.String("user_id", userID.String()),
		slog.Uint64("day_of_week", uint64(dayOfWeek)),
	)
	query, args, err := spr.buildSelectQueryByUserID(userID, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to build select query by user's id and day_of_week: %w", err)
	}

	spr.log.Debug("exec the select query", slog.String("query", query))
	row := spr.db.QueryRowContext(ctx, query, args)

	spr.log.Debug("map the row of schedule",
		slog.String("user_id", userID.String()),
		slog.Uint64("day_of_week", uint64(dayOfWeek)),
	)

	record, err := spr.mapRow(row)
	if err != nil {
		if safeErr, ok := spr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get schedule by user's id and day of week: %w", err)
	}

	return recordToSchedule(record)
}

func (spr *schedulePersistenceRepository) buildSelectQueryByUserID(userID uuid.UUID, dayOfWeek uint8) (string, []any, error) {
	return spr.sqlBuilder.Select(
		"id",
		"user_id",
		"day_of_week",
		"start_time",
		"end_time",
		"breaks",
	).From("schedules").Where("user_id = ? AND day_of_week = ?", userID, dayOfWeek).ToSql()
}

func (spr *schedulePersistenceRepository) GetScheduleByID(ctx context.Context, id uuid.UUID) (*schedule.Schedule, error) {
	spr.log.Debug("build the select query by id", slog.String("schedule_id", id.String()))
	query, args, err := spr.buildSelectQueryByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	spr.log.Debug("exec the select query by schedule's id", slog.String("query", query))
	row := spr.db.QueryRowContext(ctx, query, args)

	spr.log.Debug("map the row of schedule", slog.String("schedule_id", id.String()))
	record, err := spr.mapRow(row)
	if err != nil {
		if safeErr, ok := spr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get schedule by id: %w", err)
	}

	return recordToSchedule(record)
}

func (spr *schedulePersistenceRepository) buildSelectQueryByID(id uuid.UUID) (string, []any, error) {
	return spr.sqlBuilder.Select(
		"id",
		"user_id",
		"day_of_week",
		"start_time",
		"end_time",
		"breaks",
	).From("schedules").Where("id = ?", id).ToSql()
}

func (spr *schedulePersistenceRepository) Update(ctx context.Context, updSchedule *schedule.Schedule) (*schedule.Schedule, error) {
	spr.log.Debug("convert upd schedule to record", slog.String("schedule_id", updSchedule.ID().String()))
	updRecord, err := scheduleToRecord(updSchedule)
	if err != nil {
		return nil, fmt.Errorf("failed to convert domain into record: %w", err)
	}

	spr.log.Debug("build the sql update query", slog.String("schedule_id", updRecord.ID.String()))
	query, args, err := spr.buildUpdateQuery(updRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to build update query: %w", err)
	}

	spr.log.Debug("exec the sql update query", slog.String("query", query))
	result, err := spr.db.ExecContext(ctx, query, args)
	if err != nil {
		if safeErr, ok := spr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to update schedule: %w", err)
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return nil, schedule.ErrScheduleNotFound
	}

	return recordToSchedule(updRecord)
}

func (spr *schedulePersistenceRepository) buildUpdateQuery(updRecord *scheduleRecord) (string, []any, error) {
	return spr.sqlBuilder.Update("schedules").SetMap(map[string]any{
		"start_time": updRecord.StartTime,
		"end_time":   updRecord.EndTime,
		"breaks":     updRecord.Breaks,
	}).Where("id = ?", updRecord.ID).ToSql()
}

func (spr *schedulePersistenceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	spr.log.Debug("build the sql delete query", slog.String("id", id.String()))
	query, args, err := spr.buildDeleteQuery(id)
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	spr.log.Debug("exec the delete query", slog.String("query", query))
	result, err := spr.db.ExecContext(ctx, query, args)
	if err != nil {
		if safeErr, ok := spr.handleError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to build delete query: %w", err)
	}

	if count, _ := result.RowsAffected(); count == 0 {
		return schedule.ErrScheduleNotFound
	}

	return nil
}

func (spr *schedulePersistenceRepository) buildDeleteQuery(id uuid.UUID) (string, []any, error) {
	return spr.sqlBuilder.Delete("schedules").Where("id = ?", id).ToSql()
}

func (spr *schedulePersistenceRepository) handleError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.Canceled),
		errors.Is(err, context.DeadlineExceeded):
		return err, true
	case errors.Is(err, sql.ErrNoRows):
		return schedule.ErrScheduleNotFound, true
	}

	var pqError *pq.Error
	if errors.As(err, &pqError) {
		switch pqError.Code {
		case "23505":
			return schedule.ErrScheduleAlreadyExists, true
		case "23503":
			return schedule.ErrUserNotFound, true
		}
	}

	return err, false
}

func (spr *schedulePersistenceRepository) mapRows(rows *sql.Rows) ([]scheduleRecord, error) {
	var records []scheduleRecord
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

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (spr *schedulePersistenceRepository) mapRow(row *sql.Row) (*scheduleRecord, error) {
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

	return &record, nil
}
