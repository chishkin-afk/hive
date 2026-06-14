package scheduleredis

import (
	"encoding/json"

	"github.com/devathh/hive/backend/internal/modules/schedules/domain/schedule"
)

func schedulesToBytes(schedules []*schedule.Schedule) ([]byte, error) {
	records := make([]scheduleRecord, len(schedules))
	for i, schedule := range schedules {
		record := scheduleRecord{
			ID:        schedule.ID(),
			UserID:    schedule.UserID(),
			DayOfWeek: schedule.DayOfWeek(),
			StartTime: convertDayTimeToRecord(schedule.StartTime()),
			EndTime:   convertDayTimeToRecord(schedule.EndTime()),
			Breaks:    make([]breakRecord, len(schedule.Breaks())),
		}

		for j, br := range schedule.Breaks() {
			record.Breaks[j] = breakRecord{
				StartTime: convertDayTimeToRecord(br.StartTime()),
				EndTime:   convertDayTimeToRecord(br.EndTime()),
			}
		}

		records[i] = record
	}

	return json.Marshal(records)
}

func bytesToSchedules(bytes []byte) ([]*schedule.Schedule, error) {
	var records []scheduleRecord
	if err := json.Unmarshal(bytes, &records); err != nil {
		return nil, err
	}

	schedules := make([]*schedule.Schedule, len(records))
	for i, record := range records {
		breaks := make([]schedule.Break, len(record.Breaks))
		for j, br := range record.Breaks {
			breaks[j] = schedule.FromBreak(
				schedule.FromDayTime(
					br.StartTime.Hour,
					br.StartTime.Minute,
				),
				schedule.FromDayTime(
					br.EndTime.Hour,
					br.EndTime.Minute,
				),
			)
		}

		schedules[i] = schedule.From(
			record.ID,
			record.UserID,
			record.DayOfWeek,
			schedule.FromDayTime(
				record.StartTime.Hour,
				record.StartTime.Minute,
			),
			schedule.FromDayTime(
				record.EndTime.Hour,
				record.EndTime.Minute,
			),
			breaks,
		)
	}

	return schedules, nil
}

func convertDayTimeToRecord(dayTime schedule.DayTime) dayTimeRecord {
	return dayTimeRecord{
		Hour:   dayTime.Hour(),
		Minute: dayTime.Minute(),
	}
}
