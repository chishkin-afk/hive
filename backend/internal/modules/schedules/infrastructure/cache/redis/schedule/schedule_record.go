package scheduleredis

import "github.com/google/uuid"

type scheduleRecord struct {
	ID        uuid.UUID     `json:"id"`
	UserID    uuid.UUID     `json:"user_id"`
	DayOfWeek uint8         `json:"day_of_week"`
	StartTime dayTimeRecord `json:"start_time"`
	EndTime   dayTimeRecord `json:"end_time"`
	Breaks    []breakRecord `json:"breaks"`
}

type dayTimeRecord struct {
	Hour   uint8 `json:"hour"`
	Minute uint8 `json:"minute"`
}

type breakRecord struct {
	StartTime dayTimeRecord `json:"start_time"`
	EndTime   dayTimeRecord `json:"end_time"`
}
