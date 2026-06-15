package schedule

import "errors"

var (
	ErrInvalidBreak = errors.New("interval beetween break's must be more than 5 minutes")
	ErrBreakOverlap = errors.New("some breaks is overlap")
)

type Break struct {
	startTime DayTime
	endTime   DayTime
}

func NewBreak(startTime, endTime DayTime) (Break, error) {
	duration := endTime.AllMinutes() - startTime.AllMinutes()
	if duration < 5 || duration >= 24*60 {
		return Break{}, ErrInvalidBreak
	}

	return Break{
		startTime: startTime,
		endTime:   endTime,
	}, nil
}

func FromBreak(startTime, endTime DayTime) Break {
	return Break{
		startTime: startTime,
		endTime:   endTime,
	}
}

func (b Break) StartTime() DayTime {
	return b.startTime
}

func (b Break) EndTime() DayTime {
	return b.endTime
}

func (b Break) IsValid(startTime, endTime DayTime) bool {
	if b.startTime.AllMinutes() < startTime.AllMinutes() ||
		b.startTime.AllMinutes() > endTime.AllMinutes() {
		return false
	}

	if b.endTime.AllMinutes() > endTime.AllMinutes() ||
		b.endTime.AllMinutes() < startTime.AllMinutes() {
		return false
	}

	return b.endTime.AllMinutes()-b.startTime.AllMinutes() >= 5
}

func (b Break) IsOverlap(another Break) bool {
	return b.EndTime().AllMinutes() > another.StartTime().AllMinutes() &&
		b.StartTime().AllMinutes() < another.EndTime().AllMinutes()
}
