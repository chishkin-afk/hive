package schedulepg

import "github.com/google/uuid"

type scheduleRecord struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	DayOfWeek uint8
	StartTime []byte
	EndTime   []byte
	Breaks    []byte
}

type dayTimeRecord struct {
	Hour   uint8 `json:"hours"`
	Minute uint8 `json:"minutes"`
}

type breakRecord struct {
	StartTime dayTimeRecord `json:"start_time"`
	EndTime   dayTimeRecord `json:"end_time"`
}
