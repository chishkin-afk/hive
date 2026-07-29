package requests

import "github.com/google/uuid"

type CreateSchedule struct {
	DayOfWeek uint8   `json:"day_of_week"`
	StartTime DayTime `json:"start_time"`
	EndTime   DayTime `json:"end_time"`
	Breaks    []Break `json:"breaks"`
}

type UpdateSchedule struct {
	ScheduleID    uuid.UUID `json:"schedule_id"`
	StartTime     *DayTime  `json:"start_time"`
	EndTime       *DayTime  `json:"end_time"`
	DoUpdateBreak bool      `json:"do_update_break"`
	Breaks        []Break   `json:"breaks"`
}

type DayTime struct {
	Hour   uint8 `json:"hour"`
	Minute uint8 `json:"minute"`
}

type Break struct {
	StartTime DayTime `json:"start_time"`
	EndTime   DayTime `json:"end_time"`
}
