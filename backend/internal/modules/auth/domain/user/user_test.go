package user

import (
	"errors"
	"strings"
	"testing"
	"time"
	"uuid"
)

func TestNew(t *testing.T) {
	before := time.Now().UTC()
	u, err := New(" \tIVAN@Example.COM\n", " Иван ", "\tПетров\n", "\u00a0Иванович\u2003")
	after := time.Now().UTC()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if u == nil {
		t.Fatal("New() returned nil")
	}
	if u.ID() == (uuid.UUID{}) {
		t.Error("ID() is zero")
	}
	if u.Email() != "ivan@example.com" {
		t.Errorf("Email() = %q, want ivan@example.com", u.Email())
	}
	if u.Firstname() != "Иван" || u.Lastname() != "Петров" || u.Surname() != "Иванович" {
		t.Errorf("names were not trimmed: %q, %q, %q", u.Firstname(), u.Lastname(), u.Surname())
	}
	assertTimeBetween(t, u.CreatedAt(), before, after)
	if !u.UpdatedAt().Equal(u.CreatedAt()) {
		t.Error("UpdatedAt() must equal CreatedAt() for a new user")
	}
	if u.Notification() == nil {
		t.Fatal("Notification() is nil")
	}
	if u.Notification().TgEnabled() || u.Notification().EmailEnabled() || u.Notification().TgChatID() != 0 {
		t.Error("new user must have disabled notifications and no chat ID")
	}

	other := newTestUser(t)
	if u.ID() == other.ID() {
		t.Error("different users have the same ID")
	}
	if err := u.Notification().EnableEmail(); err != nil {
		t.Fatal(err)
	}
	if other.Notification().EmailEnabled() {
		t.Error("users share notification preferences")
	}
}

func TestNewInvalidEmail(t *testing.T) {
	for _, email := range []Email{"", " \t\n", "invalid", "ivan@", "@example.com"} {
		t.Run(string(email), func(t *testing.T) {
			u, err := New(email, "Иван", "Петров", "")
			if !errors.Is(err, ErrInvalidEmail) {
				t.Fatalf("New() error = %v, want %v", err, ErrInvalidEmail)
			}
			if u != nil {
				t.Fatal("New() must return nil on validation failure")
			}
		})
	}
}

func TestUserNames(t *testing.T) {
	fields := []struct {
		name     string
		optional bool
		wantErr  error
		new      func(string) (*User, error)
		change   func(*User, string) error
		get      func(*User) string
	}{
		{
			name: "firstname", wantErr: ErrInvalidFirstname,
			new: func(value string) (*User, error) {
				return New("ivan@example.com", value, "Петров", "Иванович")
			},
			change: (*User).ChangeFirstname, get: (*User).Firstname,
		},
		{
			name: "lastname", wantErr: ErrInvalidLastname,
			new: func(value string) (*User, error) {
				return New("ivan@example.com", "Иван", value, "Иванович")
			},
			change: (*User).ChangeLastname, get: (*User).Lastname,
		},
		{
			name: "surname", optional: true, wantErr: ErrInvalidSurname,
			new: func(value string) (*User, error) {
				return New("ivan@example.com", "Иван", "Петров", value)
			},
			change: (*User).ChangeSurname, get: (*User).Surname,
		},
	}
	tests := []struct {
		name  string
		value string
		want  string
		valid bool
	}{
		{name: "empty"},
		{name: "whitespace", value: " \t\n\u00a0\u2003"},
		{name: "one character", value: "A", want: "A"},
		{name: "one unicode character", value: "Я", want: "Я"},
		{name: "minimum", value: "Al", want: "Al", valid: true},
		{name: "unicode minimum", value: "Ли", want: "Ли", valid: true},
		{name: "maximum", value: strings.Repeat("A", 63), want: strings.Repeat("A", 63), valid: true},
		{name: "unicode maximum", value: strings.Repeat("Я", 63), want: strings.Repeat("Я", 63), valid: true},
		{name: "upper boundary", value: strings.Repeat("A", 64), want: strings.Repeat("A", 64)},
		{name: "unicode upper boundary", value: strings.Repeat("Я", 64), want: strings.Repeat("Я", 64)},
		{name: "too long", value: strings.Repeat("A", 65), want: strings.Repeat("A", 65)},
		{name: "trimmed", value: " \t\u00a0Анна-Мария\u2003\n", want: "Анна-Мария", valid: true},
		{name: "trim before validation", value: " Я ", want: "Я"},
		{name: "trimmed maximum", value: " " + strings.Repeat("Я", 63) + " ", want: strings.Repeat("Я", 63), valid: true},
	}

	for _, field := range fields {
		t.Run(field.name, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					var wantErr error
					if !tt.valid && !(field.optional && tt.want == "") {
						wantErr = field.wantErr
					}

					t.Run("new", func(t *testing.T) {
						u, err := field.new(tt.value)
						if !errors.Is(err, wantErr) {
							t.Fatalf("New() error = %v, want %v", err, wantErr)
						}
						if wantErr != nil {
							if u != nil {
								t.Fatal("New() must return nil on validation failure")
							}
							return
						}
						if u == nil {
							t.Fatal("New() returned nil")
						}
						if got := field.get(u); got != tt.want {
							t.Errorf("name = %q, want %q", got, tt.want)
						}
					})

					t.Run("change", func(t *testing.T) {
						u := newTestUser(t)
						u.updatedAt = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
						original := *u
						before := time.Now().UTC()
						err := field.change(u, tt.value)
						after := time.Now().UTC()
						if !errors.Is(err, wantErr) {
							t.Fatalf("change error = %v, want %v", err, wantErr)
						}
						if wantErr != nil {
							if *u != original {
								t.Error("failed change modified user state")
							}
							return
						}
						if got := field.get(u); got != tt.want {
							t.Errorf("name = %q, want %q", got, tt.want)
						}
						assertTimeBetween(t, u.UpdatedAt(), before, after)
						expected := original
						expected.updatedAt = u.UpdatedAt()
						switch field.name {
						case "firstname":
							expected.firstname = tt.want
						case "lastname":
							expected.lastname = tt.want
						case "surname":
							expected.surname = tt.want
						}
						if *u != expected {
							t.Error("change modified unrelated user state")
						}
					})
				})
			}
		})
	}
}

func newTestUser(t *testing.T) *User {
	t.Helper()
	u, err := New("ivan@example.com", "Иван", "Петров", "Иванович")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if u == nil {
		t.Fatal("New() returned nil")
	}
	return u
}

func assertTimeBetween(t *testing.T, got, before, after time.Time) {
	t.Helper()
	if got.Before(before) || got.After(after) {
		t.Errorf("time = %v, want between %v and %v", got, before, after)
	}
	if got.Location() != time.UTC {
		t.Errorf("time location = %v, want UTC", got.Location())
	}
}
