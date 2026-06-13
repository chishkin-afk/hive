package schedulepg

import (
	"encoding/json"

	"github.com/devathh/hive/backend/internal/modules/schedules/domain/schedule"
)

// TODO: write the From methods to VOs

func scheduleToRecord(schedule *schedule.Schedule) (*scheduleRecord, error) {
	breakRecords := make([]breakRecord, len(schedule.Breaks()))
	for i, br := range schedule.Breaks() {
		breakRecords[i] = breakRecord{
			StartTime: dayTimeToRecord(br.StartTime()),
			EndTime:   dayTimeToRecord(br.EndTime()),
		}
	}

	breakBytes, err := json.Marshal(breakRecords)
	if err != nil {
		return nil, err
	}

	startTimeBytes, err := json.Marshal(dayTimeToRecord(schedule.StartTime()))
	if err != nil {
		return nil, err
	}

	endTimeBytes, err := json.Marshal(dayTimeToRecord(schedule.EndTime()))
	if err != nil {
		return nil, err
	}

	return &scheduleRecord{
		ID:        schedule.ID(),
		UserID:    schedule.UserID(),
		DayOfWeek: schedule.DayOfWeek(),
		StartTime: startTimeBytes,
		EndTime:   endTimeBytes,
		Breaks:    breakBytes,
	}, nil
}

func recordToSchedule(record *scheduleRecord) (*schedule.Schedule, error) {
	var breakRecords []breakRecord
	if err := json.Unmarshal(record.Breaks, &breakRecords); err != nil {
		return nil, err
	}

	var startTimeRecord dayTimeRecord
	if err := json.Unmarshal(record.StartTime, &startTimeRecord); err != nil {
		return nil, err
	}
	startTime, err := schedule.NewDayTime(startTimeRecord.Hour, startTimeRecord.Minute)
	if err != nil {
		return nil, err
	}

	var endTimeRecord dayTimeRecord
	if err := json.Unmarshal(record.EndTime, &endTimeRecord); err != nil {
		return nil, err
	}
	endTime, err := schedule.NewDayTime(endTimeRecord.Hour, endTimeRecord.Minute)
	if err != nil {
		return nil, err
	}

	breaks, err := convertBreaksToDomain(breakRecords)
	if err != nil {
		return nil, err
	}

	return schedule.From(
		record.ID,
		record.UserID,
		record.DayOfWeek,
		startTime,
		endTime,
		breaks,
	), nil
}

func convertBreaksToDomain(breakRecords []breakRecord) ([]schedule.Break, error) {
	breaks := make([]schedule.Break, len(breakRecords))
	for i, brRecord := range breakRecords {
		startTime, err := schedule.NewDayTime(brRecord.StartTime.Hour, brRecord.StartTime.Minute)
		if err != nil {
			return nil, err
		}

		endTime, err := schedule.NewDayTime(brRecord.EndTime.Hour, brRecord.EndTime.Minute)
		if err != nil {
			return nil, err
		}

		br, err := schedule.NewBreak(startTime, endTime)
		if err != nil {
			return nil, err
		}

		breaks[i] = br
	}

	return breaks, nil
}

func dayTimeToRecord(dayTime schedule.DayTime) dayTimeRecord {
	return dayTimeRecord{
		Hour:   dayTime.Hour(),
		Minute: dayTime.Minute(),
	}
}
