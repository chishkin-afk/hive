package client

import (
	"errors"
	"testing"
	"time"

	"github.com/chishkin-afk/hive/backend/internal/modules/auth/domain/user"
)

func TestNew(t *testing.T) {
	c, err := New(" IVAN@Example.COM ", " Иван ", " Петров ", " Иванович ")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if c == nil || c.User() == nil {
		t.Fatal("New() must initialize the user")
	}
	u := c.User()
	if u.Email() != "ivan@example.com" || u.Firstname() != "Иван" || u.Lastname() != "Петров" || u.Surname() != "Иванович" {
		t.Error("New() did not normalize user data")
	}
	if err := u.ChangeFirstname(" Пётр "); err != nil {
		t.Fatal(err)
	}
	if c.User().Firstname() != "Пётр" {
		t.Error("user changes were not retained")
	}
	if c.LastVisitAt() != nil {
		t.Error("LastVisitAt() must be nil for a new client")
	}
}

func TestNewInvalidUser(t *testing.T) {
	tests := []struct {
		name      string
		email     user.Email
		firstname string
		lastname  string
		surname   string
		wantErr   error
	}{
		{name: "email", email: "invalid", firstname: "Иван", lastname: "Петров", wantErr: user.ErrInvalidEmail},
		{name: "firstname", email: "ivan@example.com", firstname: "Я", lastname: "Петров", wantErr: user.ErrInvalidFirstname},
		{name: "lastname", email: "ivan@example.com", firstname: "Иван", lastname: "Я", wantErr: user.ErrInvalidLastname},
		{name: "surname", email: "ivan@example.com", firstname: "Иван", lastname: "Петров", surname: "Я", wantErr: user.ErrInvalidSurname},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.email, tt.firstname, tt.lastname, tt.surname)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
			}
			if c != nil {
				t.Fatal("New() must return nil on validation failure")
			}
		})
	}
}

func TestClientAddVisit(t *testing.T) {
	for _, visited := range []bool{false, true} {
		name := "first visit"
		if visited {
			name = "repeat visit"
		}
		t.Run(name, func(t *testing.T) {
			c := newTestClient(t)
			previousVisit := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
			if visited {
				c.lastVisitAt = &previousVisit
			}
			expected := *c
			expectedUser := *c.User()
			before := time.Now().UTC()
			c.AddVisit()
			after := time.Now().UTC()
			if c.LastVisitAt() == nil {
				t.Fatal("LastVisitAt() is nil after AddVisit()")
			}
			assertTimeBetween(t, *c.LastVisitAt(), before, after)
			expected.lastVisitAt = c.LastVisitAt()
			if *c != expected || *c.User() != expectedUser {
				t.Error("AddVisit() modified unrelated client state")
			}
		})
	}
}

func newTestClient(t *testing.T) *Client {
	t.Helper()
	c, err := New("ivan@example.com", "Иван", "Петров", "Иванович")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if c == nil {
		t.Fatal("New() returned nil")
	}
	return c
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
