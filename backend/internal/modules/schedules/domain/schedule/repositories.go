package schedule

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type SchedulePersistenceRepository interface {
	Save(ctx context.Context, schedule *Schedule) (*Schedule, error)
	GetSchedulesByUserID(ctx context.Context, userID uuid.UUID) ([]*Schedule, error)
	GetScheduleByUserID(ctx context.Context, userID uuid.UUID, dayOfWeek uint8) (*Schedule, error)
	GetScheduleByID(ctx context.Context, id uuid.UUID) (*Schedule, error)
	Update(ctx context.Context, updSchedule *Schedule) (*Schedule, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ScheduleCacheRepository interface {
	SaveSchedules(ctx context.Context, userID uuid.UUID, schedules []*Schedule) error
	GetSchedules(ctx context.Context, userID uuid.UUID) ([]*Schedule, error)
	DeleteSchedules(ctx context.Context, userID uuid.UUID) error
}

var (
	ErrScheduleNotFound      = errors.New("schedule not found")
	ErrScheduleAlreadyExists = errors.New("schedule already exists")
)
