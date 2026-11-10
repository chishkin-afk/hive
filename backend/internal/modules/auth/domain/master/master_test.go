package master

import (
	"errors"
	"testing"

	"github.com/chishkin-afk/hive/backend/internal/modules/auth/domain/client"
	"github.com/chishkin-afk/hive/backend/internal/modules/auth/domain/user"
)

func TestNew(t *testing.T) {
	m, err := New(" IVAN@Example.COM ", " Иван ", " Петров ", " Иванович ")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if m == nil || m.User() == nil {
		t.Fatal("New() must initialize the user")
	}
	u := m.User()
	if u.Email() != "ivan@example.com" || u.Firstname() != "Иван" || u.Lastname() != "Петров" || u.Surname() != "Иванович" {
		t.Error("New() did not normalize user data")
	}
	if err := u.ChangeFirstname(" Пётр "); err != nil {
		t.Fatal(err)
	}
	if m.User().Firstname() != "Пётр" {
		t.Error("user changes were not retained")
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
			m, err := New(tt.email, tt.firstname, tt.lastname, tt.surname)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
			}
			if m != nil {
				t.Fatal("New() must return nil on validation failure")
			}
		})
	}
}

func TestNewIndependentUsers(t *testing.T) {
	m, err := New("ivan@example.com", "Иван", "Петров", "")
	if err != nil {
		t.Fatal(err)
	}
	c, err := client.New("ivan@example.com", "Иван", "Петров", "")
	if err != nil {
		t.Fatal(err)
	}
	if m.User().ID() == c.User().ID() {
		t.Error("separately created master and client must have different IDs")
	}
	if err := m.User().ChangeFirstname("Пётр"); err != nil {
		t.Fatal(err)
	}
	if err := m.User().Notification().EnableEmail(); err != nil {
		t.Fatal(err)
	}
	if c.User().Firstname() != "Иван" || c.User().Notification().EmailEnabled() {
		t.Error("master changes modified client data")
	}
}
