package user

import (
	"errors"
	"testing"
)

func TestEmailValidate(t *testing.T) {
	tests := []struct {
		name    string
		email   Email
		wantErr error
	}{
		{name: "valid", email: "ivan@example.com"},
		{name: "uppercase", email: "IVAN@EXAMPLE.COM"},
		{name: "alias and subdomain", email: "ivan.petrov+booking@mail.example.com"},
		{name: "empty", wantErr: ErrInvalidEmail},
		{name: "whitespace", email: " \t\n", wantErr: ErrInvalidEmail},
		{name: "missing at", email: "ivan.example.com", wantErr: ErrInvalidEmail},
		{name: "missing local part", email: "@example.com", wantErr: ErrInvalidEmail},
		{name: "missing domain", email: "ivan@", wantErr: ErrInvalidEmail},
		{name: "multiple at signs", email: "ivan@@example.com", wantErr: ErrInvalidEmail},
		{name: "space in local part", email: "ivan petrov@example.com", wantErr: ErrInvalidEmail},
		{name: "consecutive dots", email: "ivan..petrov@example.com", wantErr: ErrInvalidEmail},
		{name: "multiple addresses", email: "ivan@example.com, petr@example.com", wantErr: ErrInvalidEmail},
		{name: "newline in address", email: "ivan\n@example.com", wantErr: ErrInvalidEmail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.email.Validate(); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmailNorm(t *testing.T) {
	tests := []struct {
		name  string
		email Email
		want  Email
	}{
		{name: "already normalized", email: "ivan@example.com", want: "ivan@example.com"},
		{name: "uppercase and whitespace", email: " \tIVAN@Example.COM\r\n", want: "ivan@example.com"},
		{name: "unicode whitespace", email: "\u00a0Ivan@Example.com\u2003", want: "ivan@example.com"},
		{name: "alias preserved", email: "Ivan+Booking@Example.com", want: "ivan+booking@example.com"},
		{name: "internal space preserved", email: " Ivan Petrov@Example.com ", want: "ivan petrov@example.com"},
		{name: "empty"},
		{name: "whitespace only", email: " \t\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.email.Norm()
			if got != tt.want {
				t.Fatalf("Norm() = %q, want %q", got, tt.want)
			}
			if got.Norm() != got {
				t.Fatal("Norm() is not idempotent")
			}
			if tt.email.String() != string(tt.email) {
				t.Fatal("String() must preserve the original value")
			}
		})
	}
}
