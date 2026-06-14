package schedulepg

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/devathh/hive/backend/internal/modules/schedules/domain/schedule"
	"github.com/google/uuid"
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
	panic("not impl")
}

func (spr *schedulePersistenceRepository) GetSchedulesByUserID(ctx context.Context, userID uuid.UUID) ([]*schedule.Schedule, error) {
	panic("not impl")
}

func (spr *schedulePersistenceRepository) GetScheduleByUserID(ctx context.Context, userID uuid.UUID, dayOfWeek uint8) (*schedule.Schedule, error) {
	panic("not impl")
}

func (spr *schedulePersistenceRepository) Update(ctx context.Context, updSchedule *schedule.Schedule) (*schedule.Schedule, error) {
	panic("not impl")
}

func (spr *schedulePersistenceRepository) Delete(ctx context.Context, userID uuid.UUID, dayOfWeek uint8) error {
	panic("not impl")
}
