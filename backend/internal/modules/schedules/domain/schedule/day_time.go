package schedule

import "errors"

var (
	ErrInvalidHour   = errors.New("invalid hour")
	ErrInvalidMinute = errors.New("invalid minute")
)

type DayTime struct {
	hour   uint8
	minute uint8
}

func NewDayTime(hour, minute uint8) (DayTime, error) {
	if hour > 23 {
		return DayTime{}, ErrInvalidHour
	}

	if minute > 59 {
		return DayTime{}, ErrInvalidMinute
	}

	return DayTime{
		hour:   hour,
		minute: minute,
	}, nil
}

func (dt *DayTime) Hour() uint8 {
	return dt.hour
}

func (dt *DayTime) Minute() uint8 {
	return dt.minute
}

func (dt *DayTime) AllMinutes() int {
	return int(dt.Minute()) + int(dt.hour)*60
}
