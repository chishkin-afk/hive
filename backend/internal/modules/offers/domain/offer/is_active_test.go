package offer

import (
	"testing"
)

func TestIsActive_Enable(t *testing.T) {
	t.Run("enable already enabled offer", func(t *testing.T) {
		active := IsActive(true)
		result, err := active.Enable()
		if err == nil {
			t.Error("expected error when enabling already enabled offer")
		}
		if err != ErrOfferAlreadyEnabled {
			t.Errorf("expected ErrOfferAlreadyEnabled, got %v", err)
		}
		if !result {
			t.Error("result should still be true")
		}
	})

	t.Run("enable disabled offer", func(t *testing.T) {
		active := IsActive(false)
		result, err := active.Enable()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result {
			t.Error("expected result to be true")
		}
	})
}

func TestIsActive_Disable(t *testing.T) {
	t.Run("disable already disabled offer", func(t *testing.T) {
		active := IsActive(false)
		result, err := active.Disable()
		if err == nil {
			t.Error("expected error when disabling already disabled offer")
		}
		if err != ErrOfferAlreadyDisabled {
			t.Errorf("expected ErrOfferAlreadyDisabled, got %v", err)
		}
		if result {
			t.Error("result should still be false")
		}
	})

	t.Run("disable enabled offer", func(t *testing.T) {
		active := IsActive(true)
		result, err := active.Disable()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result {
			t.Error("expected result to be false")
		}
	})
}

func TestIsActive_Bool(t *testing.T) {
	tests := []struct {
		name     string
		active   IsActive
		expected bool
	}{
		{"enabled", IsActive(true), true},
		{"disabled", IsActive(false), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.active.Bool(); got != tt.expected {
				t.Errorf("Bool() = %v, want %v", got, tt.expected)
			}
		})
	}
}
