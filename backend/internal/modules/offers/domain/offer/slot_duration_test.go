package offer

import (
	"testing"
	"time"
)

func TestSlotDuration_Validate(t *testing.T) {
	tests := []struct {
		name     string
		duration SlotDuration
		wantErr  bool
	}{
		{"zero duration", 0, true},
		{"positive duration", SlotDuration(30 * time.Minute), false},
		{"max valid duration", SlotDuration(24*time.Hour - time.Second), false},
		{"negative duration", SlotDuration(-time.Minute), true},
		{"exactly 24 hours", SlotDuration(24 * time.Hour), true},
		{"more than 24 hours", SlotDuration(25 * time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.duration.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && err != ErrInvalidSlotDuration {
				t.Errorf("expected ErrInvalidSlotDuration, got %v", err)
			}
		})
	}
}

func TestSlotDuration_Minutes(t *testing.T) {
	tests := []struct {
		name     string
		duration SlotDuration
		expected int64
	}{
		{"zero", 0, 0},
		{"30 minutes", SlotDuration(30 * time.Minute), 30},
		{"1.5 hours", SlotDuration(90 * time.Minute), 90},
		{"23 hours 59 minutes", SlotDuration(23*time.Hour + 59*time.Minute), 1439},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.duration.Minutes(); got != tt.expected {
				t.Errorf("Minutes() = %v, want %v", got, tt.expected)
			}
		})
	}
}
