package schedule

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type SchedulePersistenceRepository interface {
	Save(ctx context.Context, schedule *Schedule) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*Schedule, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, dayOfWeek uint8) (*Schedule, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Schedule, error)
	Update(ctx context.Context, updSchedule *Schedule) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ScheduleCacheRepository interface {
	Save(ctx context.Context, userID uuid.UUID, schedules []*Schedule) error
	Get(ctx context.Context, userID uuid.UUID) ([]*Schedule, error)
	Delete(ctx context.Context, userID uuid.UUID) error
}

var (
	ErrScheduleNotFound      = errors.New("schedule not found")
	ErrScheduleAlreadyExists = errors.New("schedule already exists")
	ErrUserNotFound          = errors.New("user not found")
)
