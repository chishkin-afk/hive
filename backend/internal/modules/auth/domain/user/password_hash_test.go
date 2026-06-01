package user

import (
	"encoding/hex"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestNewPasswordHash(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid password",
			password: "secure123",
			wantErr:  false,
		},
		{
			name:        "password too short",
			password:    "12345",
			wantErr:     true,
			errContains: "len of password must be more than 6",
		},
		{
			name:        "password too long",
			password:    "this_password_is_way_too_long_and_exceeds_32_chars",
			wantErr:     true,
			errContains: "len of password must be more than 6",
		},
		{
			name:     "password with spaces trimmed",
			password: "  valid123  ",
			wantErr:  false,
		},
		{
			name:     "minimum valid length",
			password: "123456",
			wantErr:  false,
		},
		{
			name:     "maximum valid length",
			password: "12345678901234567890123456789012",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := NewPasswordHash(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" && err != nil {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("NewPasswordHash() error = %v, want to contain %v", err.Error(), tt.errContains)
				}
			}
			if !tt.wantErr && hash == "" {
				t.Error("NewPasswordHash() returned empty hash for valid password")
			}
		})
	}
}

func TestPasswordHash_IsEqual(t *testing.T) {
	password := "testPassword123"
	hash, err := NewPasswordHash(password)
	if err != nil {
		t.Fatalf("Failed to create hash: %v", err)
	}

	decodedHash, err := hex.DecodeString(string(hash))
	if err != nil {
		t.Fatalf("Failed to decode hash: %v", err)
	}

	if bcrypt.CompareHashAndPassword(decodedHash, []byte(password)) != nil {
		t.Error("IsEqual() returned false for correct password")
	}

	if bcrypt.CompareHashAndPassword(decodedHash, []byte("wrongPassword")) == nil {
		t.Error("IsEqual() returned true for incorrect password")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOf(s, substr) != -1)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
