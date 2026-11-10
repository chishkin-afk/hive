package notification

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	n := New()
	if n == nil {
		t.Fatal("New() returned nil")
	}
	assertPreferences(t, n, false, false, 0)
}

func TestRestore(t *testing.T) {
	tests := []struct {
		name         string
		tgEnabled    bool
		emailEnabled bool
		chatID       int64
	}{
		{name: "disabled"},
		{name: "email only", emailEnabled: true},
		{name: "telegram only", tgEnabled: true, chatID: 123},
		{name: "both enabled", tgEnabled: true, emailEnabled: true, chatID: 123},
		{name: "saved chat", chatID: 123},
		{name: "group chat", tgEnabled: true, chatID: -100123},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := Restore(tt.tgEnabled, tt.emailEnabled, tt.chatID)
			if n == nil {
				t.Fatal("Restore() returned nil")
			}
			assertPreferences(t, n, tt.tgEnabled, tt.emailEnabled, tt.chatID)
		})
	}
}

func TestNotificationSetChatID(t *testing.T) {
	tests := []struct {
		name      string
		enabled   bool
		oldChatID int64
		chatID    int64
		wantTg    bool
	}{
		{name: "set chat", chatID: 123},
		{name: "set group chat", chatID: -100123},
		{name: "replace disabled chat", oldChatID: 123, chatID: 456},
		{name: "replace enabled chat", enabled: true, oldChatID: 123, chatID: 456, wantTg: true},
		{name: "same chat", enabled: true, oldChatID: 123, chatID: 123, wantTg: true},
		{name: "clear enabled chat", enabled: true, oldChatID: 123},
		{name: "clear disabled chat", oldChatID: 123},
		{name: "clear empty chat"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := Restore(tt.enabled, true, tt.oldChatID)
			n.SetChatID(tt.chatID)
			assertPreferences(t, n, tt.wantTg, true, tt.chatID)
		})
	}
}

func TestNotificationToggle(t *testing.T) {
	tests := []struct {
		name      string
		tg        bool
		email     bool
		chatID    int64
		action    func(*Notification) error
		wantTg    bool
		wantEmail bool
		wantErr   error
	}{
		{name: "enable telegram without chat", email: true, action: (*Notification).EnableTg, wantEmail: true, wantErr: ErrNoChatID},
		{name: "enable telegram", email: true, chatID: 123, action: (*Notification).EnableTg, wantTg: true, wantEmail: true},
		{name: "enable telegram for group", chatID: -100123, action: (*Notification).EnableTg, wantTg: true},
		{name: "telegram already enabled", tg: true, email: true, chatID: 123, action: (*Notification).EnableTg, wantTg: true, wantEmail: true, wantErr: ErrTgAlreadyEnabled},
		{name: "disable telegram", tg: true, email: true, chatID: 123, action: (*Notification).DisableTg, wantEmail: true},
		{name: "telegram already disabled", email: true, chatID: 123, action: (*Notification).DisableTg, wantEmail: true, wantErr: ErrTgAlreadyDisabled},
		{name: "disable telegram without chat", action: (*Notification).DisableTg, wantErr: ErrTgAlreadyDisabled},
		{name: "enable email", tg: true, chatID: 123, action: (*Notification).EnableEmail, wantTg: true, wantEmail: true},
		{name: "enable email without chat", action: (*Notification).EnableEmail, wantEmail: true},
		{name: "email already enabled", tg: true, email: true, chatID: 123, action: (*Notification).EnableEmail, wantTg: true, wantEmail: true, wantErr: ErrEmailAlreadyEnabled},
		{name: "disable email", tg: true, email: true, chatID: 123, action: (*Notification).DisableEmail, wantTg: true},
		{name: "email already disabled", tg: true, chatID: 123, action: (*Notification).DisableEmail, wantTg: true, wantErr: ErrEmailAlreadyDisabled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := Restore(tt.tg, tt.email, tt.chatID)
			if err := tt.action(n); !errors.Is(err, tt.wantErr) {
				t.Fatalf("toggle error = %v, want %v", err, tt.wantErr)
			}
			assertPreferences(t, n, tt.wantTg, tt.wantEmail, tt.chatID)
		})
	}
}

func TestNotificationReconnectTelegram(t *testing.T) {
	n := New()
	n.SetChatID(123)
	if err := n.EnableTg(); err != nil {
		t.Fatal(err)
	}
	if err := n.DisableTg(); err != nil {
		t.Fatal(err)
	}
	assertPreferences(t, n, false, false, 123)
	if err := n.EnableTg(); err != nil {
		t.Fatalf("re-enable with saved chat: %v", err)
	}
	n.SetChatID(0)
	if err := n.EnableTg(); !errors.Is(err, ErrNoChatID) {
		t.Fatalf("enable after removing chat: error = %v, want %v", err, ErrNoChatID)
	}
	n.SetChatID(456)
	assertPreferences(t, n, false, false, 456)
	if err := n.EnableTg(); err != nil {
		t.Fatalf("enable with new chat: %v", err)
	}
	assertPreferences(t, n, true, false, 456)
}

func assertPreferences(t *testing.T, n *Notification, tg, email bool, chatID int64) {
	t.Helper()
	if n.TgEnabled() != tg {
		t.Errorf("TgEnabled() = %v, want %v", n.TgEnabled(), tg)
	}
	if n.EmailEnabled() != email {
		t.Errorf("EmailEnabled() = %v, want %v", n.EmailEnabled(), email)
	}
	if n.TgChatID() != chatID {
		t.Errorf("TgChatID() = %d, want %d", n.TgChatID(), chatID)
	}
}
