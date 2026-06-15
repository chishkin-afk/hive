package schedulepg

import (
	"reflect"
	"testing"

	"github.com/devathh/hive/backend/internal/modules/schedules/domain/schedule"
	"github.com/google/uuid"
)

func TestScheduleToRecord(t *testing.T) {
	userID := uuid.New()
	id := uuid.New()
	startTime := schedule.FromDayTime(9, 0)
	endTime := schedule.FromDayTime(18, 0)
	break1 := schedule.FromBreak(schedule.FromDayTime(13, 0), schedule.FromDayTime(13, 30))
	breaks := []schedule.Break{break1}

	sch := schedule.From(id, userID, 1, startTime, endTime, breaks)

	record, err := scheduleToRecord(sch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if record.ID != id {
		t.Errorf("expected ID %v, got %v", id, record.ID)
	}

	if record.UserID != userID {
		t.Errorf("expected UserID %v, got %v", userID, record.UserID)
	}

	if record.DayOfWeek != 1 {
		t.Errorf("expected DayOfWeek 1, got %d", record.DayOfWeek)
	}
}

func TestRecordToSchedule(t *testing.T) {
	userID := uuid.New()
	id := uuid.New()

	startTimeBytes := []byte(`{"hours":9,"minutes":0}`)
	endTimeBytes := []byte(`{"hours":18,"minutes":0}`)
	breaksBytes := []byte(`[{"start_time":{"hours":13,"minutes":0},"end_time":{"hours":13,"minutes":30}}]`)

	record := &scheduleRecord{
		ID:        id,
		UserID:    userID,
		DayOfWeek: 1,
		StartTime: startTimeBytes,
		EndTime:   endTimeBytes,
		Breaks:    breaksBytes,
	}

	sch, err := recordToSchedule(record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sch.ID() != id {
		t.Errorf("expected ID %v, got %v", id, sch.ID())
	}

	if sch.UserID() != userID {
		t.Errorf("expected UserID %v, got %v", userID, sch.UserID())
	}

	if sch.DayOfWeek() != 1 {
		t.Errorf("expected DayOfWeek 1, got %d", sch.DayOfWeek())
	}

	if sch.StartTime().Hour() != 9 || sch.StartTime().Minute() != 0 {
		t.Errorf("expected StartTime 9:00, got %d:%d", sch.StartTime().Hour(), sch.StartTime().Minute())
	}

	if sch.EndTime().Hour() != 18 || sch.EndTime().Minute() != 0 {
		t.Errorf("expected EndTime 18:00, got %d:%d", sch.EndTime().Hour(), sch.EndTime().Minute())
	}

	if len(sch.Breaks()) != 1 {
		t.Fatalf("expected 1 break, got %d", len(sch.Breaks()))
	}

	br := sch.Breaks()[0]
	if br.StartTime().Hour() != 13 || br.StartTime().Minute() != 0 {
		t.Errorf("expected Break StartTime 13:00, got %d:%d", br.StartTime().Hour(), br.StartTime().Minute())
	}
	if br.EndTime().Hour() != 13 || br.EndTime().Minute() != 30 {
		t.Errorf("expected Break EndTime 13:30, got %d:%d", br.EndTime().Hour(), br.EndTime().Minute())
	}
}

func TestRoundTrip(t *testing.T) {
	userID := uuid.New()
	id := uuid.New()
	startTime := schedule.FromDayTime(10, 30)
	endTime := schedule.FromDayTime(19, 45)
	break1 := schedule.FromBreak(schedule.FromDayTime(14, 0), schedule.FromDayTime(14, 15))
	break2 := schedule.FromBreak(schedule.FromDayTime(16, 0), schedule.FromDayTime(16, 10))
	breaks := []schedule.Break{break1, break2}

	original := schedule.From(id, userID, 3, startTime, endTime, breaks)

	record, err := scheduleToRecord(original)
	if err != nil {
		t.Fatalf("scheduleToRecord error: %v", err)
	}

	restored, err := recordToSchedule(record)
	if err != nil {
		t.Fatalf("recordToSchedule error: %v", err)
	}

	if !reflect.DeepEqual(original.ID(), restored.ID()) {
		t.Errorf("ID mismatch")
	}
	if !reflect.DeepEqual(original.UserID(), restored.UserID()) {
		t.Errorf("UserID mismatch")
	}
	if original.DayOfWeek() != restored.DayOfWeek() {
		t.Errorf("DayOfWeek mismatch")
	}
	if original.StartTime().AllMinutes() != restored.StartTime().AllMinutes() {
		t.Errorf("StartTime mismatch")
	}
	if original.EndTime().AllMinutes() != restored.EndTime().AllMinutes() {
		t.Errorf("EndTime mismatch")
	}
	if len(original.Breaks()) != len(restored.Breaks()) {
		t.Fatalf("Breaks length mismatch")
	}
	for i := range original.Breaks() {
		if original.Breaks()[i].StartTime().AllMinutes() != restored.Breaks()[i].StartTime().AllMinutes() {
			t.Errorf("Break %d StartTime mismatch", i)
		}
		if original.Breaks()[i].EndTime().AllMinutes() != restored.Breaks()[i].EndTime().AllMinutes() {
			t.Errorf("Break %d EndTime mismatch", i)
		}
	}
}
