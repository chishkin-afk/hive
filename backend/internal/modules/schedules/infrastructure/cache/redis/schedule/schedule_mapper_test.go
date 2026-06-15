package scheduleredis

import (
	"reflect"
	"testing"

	"github.com/devathh/hive/backend/internal/modules/schedules/domain/schedule"
	"github.com/google/uuid"
)

func TestSchedulesToBytes(t *testing.T) {
	userID := uuid.New()
	id1 := uuid.New()
	id2 := uuid.New()

	sch1 := schedule.From(id1, userID, 1, schedule.FromDayTime(9, 0), schedule.FromDayTime(18, 0), []schedule.Break{})
	sch2 := schedule.From(id2, userID, 2, schedule.FromDayTime(10, 0), schedule.FromDayTime(19, 0), []schedule.Break{})

	schedules := []*schedule.Schedule{sch1, sch2}

	bytes, err := schedulesToBytes(schedules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(bytes) == 0 {
		t.Fatal("expected non-empty bytes")
	}
}

func TestBytesToSchedules(t *testing.T) {
	userID := uuid.New()
	id1 := uuid.New()

	sch1 := schedule.From(id1, userID, 1, schedule.FromDayTime(9, 0), schedule.FromDayTime(18, 0), []schedule.Break{})
	schedules := []*schedule.Schedule{sch1}

	bytes, err := schedulesToBytes(schedules)
	if err != nil {
		t.Fatalf("schedulesToBytes error: %v", err)
	}

	restored, err := bytesToSchedules(bytes)
	if err != nil {
		t.Fatalf("bytesToSchedules error: %v", err)
	}

	if len(restored) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(restored))
	}

	if restored[0].ID() != id1 {
		t.Errorf("expected ID %v, got %v", id1, restored[0].ID())
	}

	if restored[0].UserID() != userID {
		t.Errorf("expected UserID %v, got %v", userID, restored[0].UserID())
	}

	if restored[0].DayOfWeek() != 1 {
		t.Errorf("expected DayOfWeek 1, got %d", restored[0].DayOfWeek())
	}
}

func TestRoundTripRedis(t *testing.T) {
	userID := uuid.New()
	id1 := uuid.New()
	id2 := uuid.New()

	break1 := schedule.FromBreak(schedule.FromDayTime(13, 0), schedule.FromDayTime(13, 30))
	sch1 := schedule.From(id1, userID, 1, schedule.FromDayTime(9, 0), schedule.FromDayTime(18, 0), []schedule.Break{break1})

	break2 := schedule.FromBreak(schedule.FromDayTime(14, 0), schedule.FromDayTime(14, 15))
	sch2 := schedule.From(id2, userID, 3, schedule.FromDayTime(10, 30), schedule.FromDayTime(20, 0), []schedule.Break{break2})

	original := []*schedule.Schedule{sch1, sch2}

	bytes, err := schedulesToBytes(original)
	if err != nil {
		t.Fatalf("schedulesToBytes error: %v", err)
	}

	restored, err := bytesToSchedules(bytes)
	if err != nil {
		t.Fatalf("bytesToSchedules error: %v", err)
	}

	if len(restored) != len(original) {
		t.Fatalf("length mismatch")
	}

	for i := range original {
		if !reflect.DeepEqual(original[i].ID(), restored[i].ID()) {
			t.Errorf("schedule %d ID mismatch", i)
		}
		if !reflect.DeepEqual(original[i].UserID(), restored[i].UserID()) {
			t.Errorf("schedule %d UserID mismatch", i)
		}
		if original[i].DayOfWeek() != restored[i].DayOfWeek() {
			t.Errorf("schedule %d DayOfWeek mismatch", i)
		}
		if original[i].StartTime().AllMinutes() != restored[i].StartTime().AllMinutes() {
			t.Errorf("schedule %d StartTime mismatch", i)
		}
		if original[i].EndTime().AllMinutes() != restored[i].EndTime().AllMinutes() {
			t.Errorf("schedule %d EndTime mismatch", i)
		}
		if len(original[i].Breaks()) != len(restored[i].Breaks()) {
			t.Fatalf("schedule %d Breaks length mismatch", i)
		}
		for j := range original[i].Breaks() {
			if original[i].Breaks()[j].StartTime().AllMinutes() != restored[i].Breaks()[j].StartTime().AllMinutes() {
				t.Errorf("schedule %d break %d StartTime mismatch", i, j)
			}
			if original[i].Breaks()[j].EndTime().AllMinutes() != restored[i].Breaks()[j].EndTime().AllMinutes() {
				t.Errorf("schedule %d break %d EndTime mismatch", i, j)
			}
		}
	}
}

func TestEmptySchedules(t *testing.T) {
	schedules := []*schedule.Schedule{}

	bytes, err := schedulesToBytes(schedules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	restored, err := bytesToSchedules(bytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(restored) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(restored))
	}
}
