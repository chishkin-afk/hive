package user

import (
	"testing"
)

func TestEmail_Validate(t *testing.T) {
	tests := []struct {
		name    string
		email   Email
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   Email("user@example.com"),
			wantErr: false,
		},
		{
			name:    "valid email with plus",
			email:   Email("user+tag@example.co.uk"),
			wantErr: false,
		},
		{
			name:    "invalid email no @",
			email:   Email("userexample.com"),
			wantErr: true,
		},
		{
			name:    "invalid email empty",
			email:   Email(""),
			wantErr: true,
		},
		{
			name:    "invalid email no domain",
			email:   Email("user@"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.email.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Email.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmail_Norm(t *testing.T) {
	tests := []struct {
		name     string
		email    Email
		expected Email
	}{
		{
			name:     "uppercase to lowercase",
			email:    Email("USER@Example.COM"),
			expected: Email("user@example.com"),
		},
		{
			name:     "trim spaces",
			email:    Email("  user@example.com  "),
			expected: Email("user@example.com"),
		},
		{
			name:     "already normalized",
			email:    Email("user@example.com"),
			expected: Email("user@example.com"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.email.Norm(); got != tt.expected {
				t.Errorf("Email.Norm() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEmail_String(t *testing.T) {
	email := Email("test@example.com")
	if got := email.String(); got != "test@example.com" {
		t.Errorf("Email.String() = %v, want %v", got, "test@example.com")
	}
}