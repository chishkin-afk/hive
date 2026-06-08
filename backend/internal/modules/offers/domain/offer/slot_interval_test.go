package offer

import (
	"testing"
	"time"
)

func TestSlotInterval_Validate(t *testing.T) {
	tests := []struct {
		name     string
		interval SlotInterval
		wantErr  bool
	}{
		{"zero interval", 0, true},
		{"positive interval", SlotInterval(15 * time.Minute), false},
		{"max valid interval", SlotInterval(24*time.Hour - time.Second), false},
		{"negative interval", SlotInterval(-time.Minute), true},
		{"exactly 24 hours", SlotInterval(24 * time.Hour), true},
		{"more than 24 hours", SlotInterval(25 * time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.interval.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && err != ErrInvalidSlotInterval {
				t.Errorf("expected ErrInvalidSlotInterval, got %v", err)
			}
		})
	}
}

func TestSlotInterval_Minutes(t *testing.T) {
	tests := []struct {
		name     string
		interval SlotInterval
		expected int64
	}{
		{"zero", 0, 0},
		{"15 minutes", SlotInterval(15 * time.Minute), 15},
		{"1 hour", SlotInterval(time.Hour), 60},
		{"23 hours 59 minutes", SlotInterval(23*time.Hour + 59*time.Minute), 1439},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.interval.Minutes(); got != tt.expected {
				t.Errorf("Minutes() = %v, want %v", got, tt.expected)
			}
		})
	}
}
