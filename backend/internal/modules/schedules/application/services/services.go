package scheduleservices

import (
	"context"
	"errors"
	"log/slog"
	"slices"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	contextvalue "github.com/devathh/hive/backend/internal/common/context"
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
		return nil, err
	}

	schedule, err := s.createSchedule(req, userID)
	if err != nil {
		return nil, err
	}

	savedSchedule, err := s.storeOfferPersistence(ctx, schedule)
	if err != nil {
		return nil, err
	}

	if err := s.clearSchedulesCache(ctx, userID); err != nil {
		s.log.Warn("failed to clear schedules cache", slog.String("error", err.Error()))
	}

	return s.scheduleToDTO(savedSchedule), nil
}

// GetSchedulesByUserID is a service's method that
// at first getting all the schedules from cache,
// if they doesn't exist - take from db
func (s *ScheduleService) GetSchedulesByUserID(ctx context.Context, userID uuid.UUID) (*responses.Schedules, error) {
	if list, ok := s.listSchedulesFromCache(ctx, userID); ok {
		return s.schedulesToDTO(list), nil
	}

	list, err := s.listSchedulesFromPersistence(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.storeScheduleCache(ctx, list, userID); err != nil {
		s.log.Warn("failed to store schedule into cache", slog.String("error", err.Error()))
	}

	return s.schedulesToDTO(list), nil
}

// GetScheduleByUserID is a service's method that
// returns the schedule from db by user's id & day of week
func (s *ScheduleService) GetScheduleByUserID(ctx context.Context, userID uuid.UUID, dayOfWeek uint8) (*responses.Schedule, error) {
	schedule, err := s.getScheduleFromPersistence(ctx, userID, dayOfWeek)
	if err != nil {
		return nil, err
	}

	return s.scheduleToDTO(schedule), nil
}

// Update is a service's method that
// updates the fields of the schedule
func (s *ScheduleService) Update(ctx context.Context, req *requests.UpdateSchedule) (*responses.Schedule, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	updSchedule, err := s.getSchedulePersistenceByID(ctx, req.ScheduleID)
	if err != nil {
		return nil, err
	}

	if err := s.checkRights(updSchedule, userID); err != nil {
		return nil, err
	}

	if err := s.applyChanges(updSchedule, req); err != nil {
		return nil, err
	}

	updatedSchedule, err := s.updateSchedulePersistence(ctx, updSchedule)
	if err != nil {
		return nil, err
	}

	if err := s.clearSchedulesCache(ctx, userID); err != nil {
		s.log.Warn("failed to clear schedules cache", slog.String("error", err.Error()))
	}

	return s.scheduleToDTO(updatedSchedule), nil
}

// Delete is a service's method that
// removes schedule from db by id
func (s *ScheduleService) Delete(ctx context.Context, id uuid.UUID) error {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return err
	}

	schedule, err := s.getSchedulePersistenceByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.checkRights(schedule, userID); err != nil {
		return err
	}

	if err := s.clearSchedulesCache(ctx, userID); err != nil {
		return err
	}

	if err := s.deleteSchedulePersistence(ctx, id); err != nil {
		return err
	}

	return nil
}

func (s *ScheduleService) getUserID(ctx context.Context) (uuid.UUID, error) {
	raw := ctx.Value(contextvalue.KeyUserID)
	if raw == nil {
		return uuid.Nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindRequest)
	}

	if id, ok := raw.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindRequest)
}

func (s *ScheduleService) storeOfferPersistence(ctx context.Context, schedule *schedule.Schedule) (*schedule.Schedule, error) {
	if err := s.schedulePersistence.Save(ctx, schedule); err != nil {
		s.log.Error("failed to save schedule into db",
			slog.String("error", err.Error()),
			slog.String("schedule_id", schedule.ID().String()),
		)

		return nil, s.handleRepoError(err)
	}

	return schedule, nil
}

func (s *ScheduleService) createSchedule(req *requests.CreateSchedule, userID uuid.UUID) (*schedule.Schedule, error) {
	breaks := make([]schedule.Break, len(req.Breaks))
	for i, brDTO := range req.Breaks {
		startTime, err := schedule.NewDayTime(brDTO.StartTime.Hour, brDTO.StartTime.Minute)
		if err != nil {
			return nil, errs.NewErrorWithKind(err, errs.KindRequest)
		}

		endTime, err := schedule.NewDayTime(brDTO.EndTime.Hour, brDTO.EndTime.Minute)
		if err != nil {
			return nil, errs.NewErrorWithKind(err, errs.KindRequest)
		}

		br, err := schedule.NewBreak(startTime, endTime)
		if err != nil {
			return nil, errs.NewErrorWithKind(err, errs.KindRequest)
		}

		breaks[i] = br
	}

	startTime, err := schedule.NewDayTime(req.StartTime.Hour, req.StartTime.Minute)
	if err != nil {
		return nil, errs.NewErrorWithKind(err, errs.KindRequest)
	}

	endTime, err := schedule.NewDayTime(req.EndTime.Hour, req.EndTime.Minute)
	if err != nil {
		return nil, errs.NewErrorWithKind(err, errs.KindRequest)
	}

	schedule, err := schedule.New(
		userID,
		req.DayOfWeek,
		startTime,
		endTime,
		breaks,
	)

	if err != nil {
		return nil, errs.NewErrorWithKind(err, errs.KindRequest)
	}

	return schedule, nil
}

func (s *ScheduleService) listSchedulesFromCache(ctx context.Context, userID uuid.UUID) ([]*schedule.Schedule, bool) {
	list, err := s.scheduleCache.Get(ctx, userID)
	if err != nil {
		if !errors.Is(err, schedule.ErrScheduleNotFound) {
			s.log.Warn("failed to get schedules from cache",
				slog.String("user_id", userID.String()),
				slog.String("error", err.Error()),
			)
		}

		return nil, false
	}

	s.log.Debug("schedules has been taken from cache", slog.String("user_id", userID.String()))
	return list, true
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

func (s *ScheduleService) listSchedulesFromPersistence(ctx context.Context, userID uuid.UUID) ([]*schedule.Schedule, error) {
	list, err := s.schedulePersistence.ListByUserID(ctx, userID)
	if err != nil {
		s.log.Error("failed to get schedules from db",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
		)

		return nil, s.handleRepoError(err)
	}

	return list, nil
}

func (s *ScheduleService) storeScheduleCache(ctx context.Context, list []*schedule.Schedule, userID uuid.UUID) error {
	return s.scheduleCache.Save(ctx, userID, list)
}

func (s *ScheduleService) getScheduleFromPersistence(ctx context.Context, userID uuid.UUID, dayOfWeek uint8) (*schedule.Schedule, error) {
	schedule, err := s.schedulePersistence.GetByUserID(ctx, userID, dayOfWeek)
	if err != nil {
		s.log.Error("failed to get schedule from db",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
			slog.Uint64("day_of_week", uint64(dayOfWeek)),
		)

		return nil, s.handleRepoError(err)
	}

	return schedule, nil
}

func (s *ScheduleService) getSchedulePersistenceByID(ctx context.Context, id uuid.UUID) (*schedule.Schedule, error) {
	schedule, err := s.schedulePersistence.GetByID(ctx, id)
	if err != nil {
		s.log.Error("failed to get schedule by id",
			slog.String("error", err.Error()),
			slog.String("id", id.String()),
		)

		return nil, s.handleRepoError(err)
	}

	return schedule, nil
}

func (s *ScheduleService) checkRights(schedule *schedule.Schedule, userID uuid.UUID) error {
	if schedule.UserID() != userID {
		return errs.NewErrorWithKind(errs.ErrNotEnoughRights, errs.KindPermissionDenied)
	}

	return nil
}

func (s *ScheduleService) applyChanges(updSchedule *schedule.Schedule, req *requests.UpdateSchedule) error {
	if req.EndTime != nil && req.StartTime != nil {
		startTime, err := schedule.NewDayTime(req.StartTime.Hour, req.StartTime.Minute)
		if err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}

		endTime, err := schedule.NewDayTime(req.EndTime.Hour, req.EndTime.Minute)
		if err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}

		if err := updSchedule.ChangeSchedule(startTime, endTime); err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}
	} else if req.EndTime != nil {
		endTime, err := schedule.NewDayTime(req.EndTime.Hour, req.EndTime.Minute)
		if err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}

		if err := updSchedule.ChangeEndTime(endTime); err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}
	} else if req.StartTime != nil {
		startTime, err := schedule.NewDayTime(req.StartTime.Hour, req.StartTime.Minute)
		if err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}

		if err := updSchedule.ChangeStartTime(startTime); err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}
	}

	if req.DoUpdateBreak {
		var breaks []schedule.Break
		for _, brDTO := range req.Breaks {
			startTime, endTime, err := s.getBreakTimes(brDTO)
			if err != nil {
				return errs.NewErrorWithKind(err, errs.KindRequest)
			}

			br, err := schedule.NewBreak(startTime, endTime)
			if err != nil {
				return errs.NewErrorWithKind(err, errs.KindRequest)
			}

			if slices.ContainsFunc(breaks, br.IsOverlap) {
				return errs.NewErrorWithKind(schedule.ErrBreakOverlap, errs.KindRequest)
			}

			breaks = append(breaks, br)
		}

		if err := updSchedule.ChangeBreaks(breaks); err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}
	}

	return nil
}

func (s *ScheduleService) updateSchedulePersistence(ctx context.Context, updSchedule *schedule.Schedule) (*schedule.Schedule, error) {
	if err := s.schedulePersistence.Update(ctx, updSchedule); err != nil {
		s.log.Error("failed to update schedule in db",
			slog.String("error", err.Error()),
			slog.String("schedule_id", updSchedule.ID().String()),
		)

		return nil, s.handleRepoError(err)
	}

	return updSchedule, nil
}

func (s *ScheduleService) clearSchedulesCache(ctx context.Context, userID uuid.UUID) error {
	return s.scheduleCache.Delete(ctx, userID)
}

func (s *ScheduleService) deleteSchedulePersistence(ctx context.Context, id uuid.UUID) error {
	if err := s.schedulePersistence.Delete(ctx, id); err != nil {
		s.log.Error("failed to delete schedule from db",
			slog.String("error", err.Error()),
			slog.String("schedule_id", id.String()),
		)

		return s.handleRepoError(err)
	}

	return nil
}

func (s *ScheduleService) getBreakTimes(brDTO requests.Break) (schedule.DayTime, schedule.DayTime, error) {
	startTime, err := schedule.NewDayTime(brDTO.StartTime.Hour, brDTO.StartTime.Minute)
	if err != nil {
		return schedule.DayTime{}, schedule.DayTime{}, err
	}

	endTime, err := schedule.NewDayTime(brDTO.EndTime.Hour, brDTO.EndTime.Minute)
	if err != nil {
		return schedule.DayTime{}, schedule.DayTime{}, err
	}

	return startTime, endTime, nil
}

func (s *ScheduleService) handleRepoError(err error) error {
	switch {
	case errors.Is(err, context.Canceled),
		errors.Is(err, context.DeadlineExceeded):
		return errs.NewErrorWithKind(err, errs.KindTimeout)
	case errors.Is(err, schedule.ErrScheduleAlreadyExists):
		return errs.NewErrorWithKind(err, errs.KindConflict)
	case errors.Is(err, schedule.ErrScheduleNotFound),
		errors.Is(err, schedule.ErrUserNotFound):
		return errs.NewErrorWithKind(err, errs.KindNotFound)
	}

	return errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal)
}

func (s *ScheduleService) scheduleToDTO(schedule *schedule.Schedule) *responses.Schedule {
	response := &responses.Schedule{
		ID:        schedule.ID(),
		UserID:    schedule.UserID(),
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
