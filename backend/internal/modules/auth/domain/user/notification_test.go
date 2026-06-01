package user

import (
	"testing"
)

func TestNewNotification(t *testing.T) {
	tests := []struct {
		name           string
		telegramEnable bool
		wantEnable     bool
	}{
		{
			name:           "telegram enabled",
			telegramEnable: true,
			wantEnable:     true,
		},
		{
			name:           "telegram disabled",
			telegramEnable: false,
			wantEnable:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewNotification(tt.telegramEnable)
			if got := n.IsTelegramEnable(); got != tt.wantEnable {
				t.Errorf("Notification.IsTelegramEnable() = %v, want %v", got, tt.wantEnable)
			}
		})
	}
}

func TestNotification_IsTelegramEnable(t *testing.T) {
	n := NewNotification(true)
	if !n.IsTelegramEnable() {
		t.Error("IsTelegramEnable() returned false for enabled notification")
	}

	n2 := NewNotification(false)
	if n2.IsTelegramEnable() {
		t.Error("IsTelegramEnable() returned true for disabled notification")
	}
}