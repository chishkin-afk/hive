package scheduleservices

import (
	"context"
	"errors"
	"log/slog"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/devathh/hive/backend/internal/common/auth"
	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/modules/schedules/domain/schedule"
	"github.com/devathh/hive/backend/pkg/errs"
	"github.com/google/uuid"
)

type ScheduleService struct {
	cfg                 *config.Config
	log                 *slog.Logger
	schedulePersistence schedule.SchedulePersistenceRepository
	scheduleCache       schedule.ScheduleCacheRepository
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	schedulePersistence schedule.SchedulePersistenceRepository,
	scheduleCache schedule.ScheduleCacheRepository,
) *ScheduleService {
	return &ScheduleService{
		cfg:                 cfg,
		log:                 log,
		schedulePersistence: schedulePersistence,
		scheduleCache:       scheduleCache,
	}
}

// Create is a service's method that
// validate the new schedule & saves its into db
func (s *ScheduleService) Create(ctx context.Context, req *requests.CreateSchedule) (*responses.Schedule, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
	}

	schedule, err := s.createSchedule(req, userID)
	if err != nil {
		return nil, errs.NewErrorWithKind(err, errs.KindRequest)
	}

	savedSchedule, err := s.schedulePersistence.Save(ctx, schedule)
	if err != nil {
		safeErr, ok := s.handleRepoError(err)
		if !ok {
			s.log.Error("failed to save schedule into db",
				slog.String("error", err.Error()),
				slog.Uint64("day_of_week", uint64(schedule.DayOfWeek())),
				slog.String("user_id", userID.String()),
			)
		}

		return nil, safeErr
	}

	return s.scheduleToDTO(savedSchedule), nil
}

// GetSchedulesByUserID is a service's method that
// at first getting all the schedules from cache,
// if they doesn't exist - take from db
func (s *ScheduleService) GetSchedulesByUserID(ctx context.Context, userID uuid.UUID) (*responses.Schedules, error) {
	if list, err := s.scheduleCache.GetSchedules(ctx, userID); err == nil {
		s.log.Debug("schedules has been taken from cache",
			slog.String("user_id", userID.String()),
		)

		return s.schedulesToDTO(list), nil
	} else if !errors.Is(err, schedule.ErrScheduleNotFound) {
		s.log.Warn("failed to get schedules by user's id from cache",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
		)
	}

	list, err := s.schedulePersistence.GetSchedulesByUserID(ctx, userID)
	if err != nil {
		safeErr, ok := s.handleRepoError(err)
		if !ok {
			s.log.Error("failed to get schedules by user's id",
				slog.String("error", err.Error()),
				slog.String("user_id", userID.String()),
			)
		}

		return nil, safeErr
	}

	if err := s.scheduleCache.SaveSchedules(ctx, userID, list); err != nil {
		s.log.Warn("failed to save schedules into cache",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
		)
	}

	return s.schedulesToDTO(list), nil
}

// GetScheduleByUserID is a service's method that
// returns the schedule from db by user's id & day of week
func (s *ScheduleService) GetScheduleByUserID(ctx context.Context, userID uuid.UUID, dayOfWeek uint8) (*responses.Schedule, error) {
	schedule, err := s.schedulePersistence.GetScheduleByUserID(ctx, userID, dayOfWeek)
	if err != nil {
		safeErr, ok := s.handleRepoError(err)
		if !ok {
			s.log.Error("failed to get schedule of the day by user id",
				slog.String("user_id", userID.String()),
				slog.String("error", err.Error()),
			)
		}

		return nil, safeErr
	}

	return s.scheduleToDTO(schedule), nil
}

// Update is a service's method that
// updates the fields of the schedule
func (s *ScheduleService) Update(ctx context.Context, req *requests.UpdateSchedule) (*responses.Schedule, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
	}

	updSchedule, err := s.schedulePersistence.GetScheduleByID(ctx, req.ScheduleID)
	if err != nil {
		safeErr, ok := s.handleRepoError(err)
		if !ok {
			s.log.Error("failed to get schedule by id",
				slog.String("error", err.Error()),
				slog.String("schedule_id", req.ScheduleID.String()),
			)
		}

		return nil, safeErr
	}

	if updSchedule.UserID() != userID {
		return nil, errs.NewErrorWithKind(errs.ErrNotEnoughRights, errs.KindPermissionDenied)
	}

	if err := s.applyChanges(updSchedule, req); err != nil {
		return nil, errs.NewErrorWithKind(err, errs.KindRequest)
	}

	updatedSchedule, err := s.schedulePersistence.Update(ctx, updSchedule)
	if err != nil {
		safeErr, ok := s.handleRepoError(err)
		if !ok {
			s.log.Error("failed to update schedule in db",
				slog.String("error", err.Error()),
				slog.String("schedule_id", updSchedule.ID().String()),
			)
		}

		return nil, safeErr
	}

	if err := s.scheduleCache.DeleteSchedules(ctx, userID); err != nil {
		s.log.Warn("failed to delete schedules from cache",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
		)
	}

	return s.scheduleToDTO(updatedSchedule), nil
}

// Delete is a service's method that
// removes schedule from db by id
func (s *ScheduleService) Delete(ctx context.Context, id uuid.UUID) error {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
	}

	schedule, err := s.schedulePersistence.GetScheduleByID(ctx, id)
	if err != nil {
		safeErr, ok := s.handleRepoError(err)
		if !ok {
			s.log.Error("failed to get schedule to delete",
				slog.String("schedule_id", id.String()),
				slog.String("user_id", userID.String()),
			)
		}

		return safeErr
	}

	if schedule.UserID() != userID {
		return errs.NewErrorWithKind(errs.ErrNotEnoughRights, errs.KindPermissionDenied)
	}

	if err := s.scheduleCache.DeleteSchedules(ctx, userID); err != nil {
		s.log.Error("failed to delete schedules from cache",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
		)
		return errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal)
	}

	if err := s.schedulePersistence.Delete(ctx, id); err != nil {
		safeErr, ok := s.handleRepoError(err)
		if !ok {
			s.log.Error("failed to delete schedule",
				slog.String("schedule_id", id.String()),
				slog.String("user_id", userID.String()),
			)
		}

		return safeErr
	}

	return nil
}

func (s *ScheduleService) applyChanges(updSchedule *schedule.Schedule, req *requests.UpdateSchedule) error {
	if req.EndTime != nil && req.StartTime != nil {
		startTime, err := schedule.NewDayTime(req.StartTime.Hour, req.StartTime.Minute)
		if err != nil {
			return err
		}

		endTime, err := schedule.NewDayTime(req.EndTime.Hour, req.EndTime.Minute)
		if err != nil {
			return err
		}

		if err := updSchedule.ChangeSchedule(startTime, endTime); err != nil {
			return err
		}
	} else if req.EndTime != nil {
		endTime, err := schedule.NewDayTime(req.EndTime.Hour, req.EndTime.Minute)
		if err != nil {
			return err
		}

		if err := updSchedule.ChangeEndTime(endTime); err != nil {
			return err
		}
	} else if req.StartTime != nil {
		startTime, err := schedule.NewDayTime(req.StartTime.Hour, req.StartTime.Minute)
		if err != nil {
			return err
		}

		if err := updSchedule.ChangeStartTime(startTime); err != nil {
			return err
		}
	}

	if req.DoUpdateBreak {
		breaks := make([]schedule.Break, len(req.Breaks))
		for i, brDTO := range req.Breaks {
			startTime, err := schedule.NewDayTime(brDTO.StartTime.Hour, brDTO.StartTime.Minute)
			if err != nil {
				return err
			}

			endTime, err := schedule.NewDayTime(brDTO.EndTime.Hour, brDTO.EndTime.Minute)
			if err != nil {
				return err
			}

			br, err := schedule.NewBreak(startTime, endTime)
			if err != nil {
				return err
			}

			breaks[i] = br
		}

		if err := updSchedule.ChangeBreaks(breaks); err != nil {
			return err
		}
	}

	return nil
}

func (s *ScheduleService) handleRepoError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.Canceled),
		errors.Is(err, context.DeadlineExceeded):
		return errs.NewErrorWithKind(err, errs.KindTimeout), true
	case errors.Is(err, schedule.ErrScheduleAlreadyExists):
		return errs.NewErrorWithKind(err, errs.KindConflict), true
	case errors.Is(err, schedule.ErrScheduleNotFound):
		return errs.NewErrorWithKind(err, errs.KindNotFound), true
	}

	return errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal), false
}

func (s *ScheduleService) schedulesToDTO(schedules []*schedule.Schedule) *responses.Schedules {
	response := responses.Schedules{
		List: make([]responses.Schedule, len(schedules)),
	}
	for i, schedule := range schedules {
		response.List[i] = *s.scheduleToDTO(schedule)
	}

	return &response
}

func (s *ScheduleService) scheduleToDTO(schedule *schedule.Schedule) *responses.Schedule {
	response := &responses.Schedule{
		ID:        schedule.ID().String(),
		UserID:    schedule.UserID().String(),
		DayOfWeek: schedule.DayOfWeek(),
		StartTime: s.dayTimeToDTO(schedule.StartTime()),
		EndTime:   s.dayTimeToDTO(schedule.EndTime()),
		Breaks:    make([]responses.Break, len(schedule.Breaks())),
	}

	for i, br := range schedule.Breaks() {
		response.Breaks[i] = responses.Break{
			StartTime: s.dayTimeToDTO(br.StartTime()),
			EndTime:   s.dayTimeToDTO(br.EndTime()),
		}
	}

	return response
}

func (s *ScheduleService) dayTimeToDTO(dayTime schedule.DayTime) responses.DayTime {
	return responses.DayTime{
		Hour:   dayTime.Hour(),
		Minute: dayTime.Minute(),
	}
}

func (s *ScheduleService) createSchedule(req *requests.CreateSchedule, userID uuid.UUID) (*schedule.Schedule, error) {
	breaks := make([]schedule.Break, len(req.Breaks))
	for i, brDTO := range req.Breaks {
		startTime, err := schedule.NewDayTime(brDTO.StartTime.Hour, brDTO.StartTime.Minute)
		if err != nil {
			return nil, err
		}

		endTime, err := schedule.NewDayTime(brDTO.EndTime.Hour, brDTO.EndTime.Minute)
		if err != nil {
			return nil, err
		}

		br, err := schedule.NewBreak(startTime, endTime)
		if err != nil {
			return nil, err
		}

		breaks[i] = br
	}

	startTime, err := schedule.NewDayTime(req.StartTime.Hour, req.StartTime.Minute)
	if err != nil {
		return nil, err
	}

	endTime, err := schedule.NewDayTime(req.EndTime.Hour, req.EndTime.Minute)
	if err != nil {
		return nil, err
	}

	return schedule.New(
		userID,
		req.DayOfWeek,
		startTime,
		endTime,
		breaks,
	)
}

func (s *ScheduleService) getUserID(ctx context.Context) (uuid.UUID, error) {
	raw := ctx.Value(auth.KeyUserID)
	if raw == nil {
		return uuid.Nil, errs.ErrInvalidToken
	}

	if id, ok := raw.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, errs.ErrInvalidToken
}
