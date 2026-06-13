package schedule

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidSchedule  = errors.New("times of schedule must be more than 5")
	ErrInvalidDayOfWeek = errors.New("invalid day of week")
	ErrInvalidUserID    = errors.New("invalid user's id")
)

type Schedule struct {
	id        uuid.UUID
	userID    uuid.UUID
	dayOfWeek uint8 // 0 - sunday
	startTime DayTime
	endTime   DayTime
	breaks    []Break
}

func New(
	userID uuid.UUID,
	dayOfWeek uint8,
	startTime DayTime,
	endTime DayTime,
	breaks []Break,
) (*Schedule, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	if endTime.AllMinutes()-startTime.AllMinutes() < 5 {
		return nil, ErrInvalidSchedule
	}

	for _, b := range breaks {
		if !b.IsValid(startTime, endTime) {
			return nil, ErrInvalidBreak
		}
	}

	if dayOfWeek > 6 {
		return nil, ErrInvalidDayOfWeek
	}

	return &Schedule{
		id:        uuid.New(),
		userID:    userID,
		dayOfWeek: dayOfWeek,
		startTime: startTime,
		endTime:   endTime,
		breaks:    breaks,
	}, nil
}

func (s *Schedule) ChangeSchedule(startTime, endTime DayTime) error {
	if endTime.AllMinutes()-startTime.AllMinutes() < 5 {
		return ErrInvalidSchedule
	}

	s.startTime = startTime
	s.endTime = endTime

	return nil
}

func (s *Schedule) ChangeBreaks(breaks []Break) error {
	for _, b := range breaks {
		if !b.IsValid(s.startTime, s.endTime) {
			return ErrInvalidBreak
		}
	}

	s.breaks = breaks

	return nil
}

func (s *Schedule) ID() uuid.UUID {
	return s.id
}

func (s *Schedule) UserID() uuid.UUID {
	return s.userID
}

func (s *Schedule) DayOfWeek() uint8 {
	return s.dayOfWeek
}

func (s *Schedule) StartTime() DayTime {
	return s.startTime
}

func (s *Schedule) EndTime() DayTime {
	return s.endTime
}

func (s *Schedule) Breaks() []Break {
	return s.breaks
}
