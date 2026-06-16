package responses

import "github.com/google/uuid"

type Schedule struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	DayOfWeek uint8     `json:"day_of_week"`
	StartTime DayTime   `json:"start_time"`
	EndTime   DayTime   `json:"end_time"`
	Breaks    []Break   `json:"breaks"`
}

type Schedules struct {
	List []Schedule `json:"list"`
}

type DayTime struct {
	Hour   uint8 `json:"hour"`
	Minute uint8 `json:"minute"`
}

type Break struct {
	StartTime DayTime `json:"start_time"`
	EndTime   DayTime `json:"end_time"`
}
