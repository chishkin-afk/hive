package user

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewUser(t *testing.T) {
	email := Email("test@example.com")

	t.Run("valid user creation", func(t *testing.T) {
		user, err := New("John", "Doe", email, "securePass123")
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		if user.firstname != "John" {
			t.Errorf("Firstname = %v, want %v", user.firstname, "John")
		}
		if user.lastname != "Doe" {
			t.Errorf("Lastname = %v, want %v", user.lastname, "Doe")
		}
		if user.confirmedEmail != email {
			t.Errorf("Email = %v, want %v", user.confirmedEmail, email)
		}
		if user.passwordHash == "" {
			t.Error("PasswordHash should not be empty")
		}
		if user.id == uuid.Nil {
			t.Error("ID should not be nil")
		}
		if user.createdAt.IsZero() {
			t.Error("CreatedAt should be set")
		}
		if user.updatedAt.IsZero() {
			t.Error("UpdatedAt should be set")
		}
	})

	t.Run("invalid firstname too short", func(t *testing.T) {
		_, err := New("Jo", "Doe", email, "securePass123")
		if err == nil {
			t.Error("Expected error for short firstname")
		}
		if _, ok := err.(*InvalidFirstnameError); !ok {
			t.Errorf("Expected InvalidFirstnameError, got %T", err)
		}
	})

	t.Run("invalid firstname too long", func(t *testing.T) {
		_, err := New("ThisNameIsWayTooLongAndExceedsLimit", "Doe", email, "securePass123")
		if err == nil {
			t.Error("Expected error for long firstname")
		}
	})

	t.Run("invalid lastname", func(t *testing.T) {
		_, err := New("John", "A", email, "securePass123")
		if err == nil {
			t.Error("Expected error for short lastname")
		}
		if _, ok := err.(*InvalidLastnameError); !ok {
			t.Errorf("Expected InvalidLastnameError, got %T", err)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		_, err := New("John", "Doe", Email("invalid"), "securePass123")
		if err == nil {
			t.Error("Expected error for invalid email")
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		_, err := New("John", "Doe", email, "12345")
		if err == nil {
			t.Error("Expected error for short password")
		}
	})

	t.Run("trim names", func(t *testing.T) {
		user, err := New("  John  ", "  Doe  ", email, "securePass123")
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		if user.firstname != "John" {
			t.Errorf("Firstname = %v, want %v", user.firstname, "John")
		}
		if user.lastname != "Doe" {
			t.Errorf("Lastname = %v, want %v", user.lastname, "Doe")
		}
	})
}

func TestUser_ChangeFirstname(t *testing.T) {
	email := Email("test@example.com")
	user, _ := New("John", "Doe", email, "securePass123")
	oldUpdated := user.updatedAt

	t.Run("valid change", func(t *testing.T) {
		err := user.ChangeFirstname("Jane")
		if err != nil {
			t.Errorf("ChangeFirstname() error = %v", err)
		}
		if user.firstname != "Jane" {
			t.Errorf("Firstname = %v, want %v", user.firstname, "Jane")
		}
		if !user.updatedAt.After(oldUpdated) {
			t.Error("UpdatedAt should be updated")
		}
	})

	t.Run("invalid name too short", func(t *testing.T) {
		err := user.ChangeFirstname("J")
		if err == nil {
			t.Error("Expected error for short name")
		}
	})

	t.Run("trim input", func(t *testing.T) {
		err := user.ChangeFirstname("  Alex  ")
		if err != nil {
			t.Errorf("ChangeFirstname() error = %v", err)
		}
		if user.firstname != "Alex" {
			t.Errorf("Firstname = %v, want %v", user.firstname, "Alex")
		}
	})
}

func TestUser_ChangeLastname(t *testing.T) {
	email := Email("test@example.com")
	user, _ := New("John", "Doe", email, "securePass123")

	t.Run("valid change", func(t *testing.T) {
		err := user.ChangeLastname("Smith")
		if err != nil {
			t.Errorf("ChangeLastname() error = %v", err)
		}
		if user.lastname != "Smith" {
			t.Errorf("Lastname = %v, want %v", user.lastname, "Smith")
		}
	})

	t.Run("invalid name", func(t *testing.T) {
		err := user.ChangeLastname("A")
		if err == nil {
			t.Error("Expected error for short lastname")
		}
	})
}

func TestUser_ChangePassword(t *testing.T) {
	email := Email("test@example.com")
	user, _ := New("John", "Doe", email, "securePass123")
	oldHash := user.passwordHash

	t.Run("valid password change", func(t *testing.T) {
		err := user.ChangePassword("newSecurePass456")
		if err != nil {
			t.Errorf("ChangePassword() error = %v", err)
		}
		if user.passwordHash == oldHash {
			t.Error("PasswordHash should be updated")
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		err := user.ChangePassword("123")
		if err == nil {
			t.Error("Expected error for short password")
		}
	})
}

func TestUser_ChangeTgChatID(t *testing.T) {
	email := Email("test@example.com")
	user, _ := New("John", "Doe", email, "securePass123")

	t.Run("valid ID", func(t *testing.T) {
		err := user.ChangeTgChatID(123456789)
		if err != nil {
			t.Errorf("ChangeTgChatID() error = %v", err)
		}
		if user.tgChatID != 123456789 {
			t.Errorf("TgChatID = %v, want %v", user.tgChatID, 123456789)
		}
	})

	t.Run("negative ID", func(t *testing.T) {
		err := user.ChangeTgChatID(-1)
		if err == nil {
			t.Error("Expected error for negative ID")
		}
	})

	t.Run("zero ID allowed", func(t *testing.T) {
		err := user.ChangeTgChatID(0)
		if err != nil {
			t.Errorf("ChangeTgChatID() error = %v", err)
		}
	})
}

func TestUser_ChangeNotification(t *testing.T) {
	email := Email("test@example.com")
	user, _ := New("John", "Doe", email, "securePass123")

	t.Run("disable telegram notification", func(t *testing.T) {
		notif := NewNotification(false)
		err := user.ChangeNotification(notif)
		if err != nil {
			t.Errorf("ChangeNotification() error = %v", err)
		}
	})

	t.Run("enable telegram without chat ID", func(t *testing.T) {
		notif := NewNotification(true)
		err := user.ChangeNotification(notif)
		if err == nil {
			t.Error("Expected error when enabling telegram without chat ID")
		}
		if err != ErrNotificationTgDisabled {
			t.Errorf("Expected ErrNotificationTgDisabled, got %v", err)
		}
	})

	t.Run("enable telegram with chat ID", func(t *testing.T) {
		user.ChangeTgChatID(999)
		notif := NewNotification(true)
		err := user.ChangeNotification(notif)
		if err != nil {
			t.Errorf("ChangeNotification() error = %v", err)
		}
	})
}

func TestUser_Getters(t *testing.T) {
	email := Email("test@example.com")
	user, _ := New("John", "Doe", email, "securePass123")

	if user.ID() == uuid.Nil {
		t.Error("ID() should return valid UUID")
	}
	if user.Firstname() != "John" {
		t.Errorf("Firstname() = %v, want %v", user.Firstname(), "John")
	}
	if user.Lastname() != "Doe" {
		t.Errorf("Lastname() = %v, want %v", user.Lastname(), "Doe")
	}
	if user.ConfirmedEmail() != email {
		t.Errorf("ConfirmedEmail() = %v, want %v", user.ConfirmedEmail(), email)
	}
	if user.PasswordHash() == "" {
		t.Error("PasswordHash() should not be empty")
	}
	if user.PaidUntil().IsZero() {
		t.Error("PaidUntil() should return valid time")
	}
	if user.OnVacation() != false {
		t.Error("OnVacation() should return false by default")
	}
	if user.TgChatID() != 0 {
		t.Errorf("TgChatID() = %v, want %v", user.TgChatID(), 0)
	}
	if user.Notification().IsTelegramEnable() != false {
		t.Error("Notification().IsTelegramEnable() should return false by default")
	}
	if user.CreatedAt().IsZero() {
		t.Error("CreatedAt() should return valid time")
	}
	if user.UpdatedAt().IsZero() {
		t.Error("UpdatedAt() should return valid time")
	}
}

func TestIsValidName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"too short", "A", false},
		{"min valid", "ABC", true},
		{"max valid", "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456", true},
		{"too long", "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567", false},
		{"unicode valid", "Аня", true},
		{"unicode too long", "АняАняАняАняАняАняАняАняАняАняАня", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidName(tt.input); got != tt.expected {
				t.Errorf("isValidName(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
