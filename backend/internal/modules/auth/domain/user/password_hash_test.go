package user

import (
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

	// Тестируем метод IsEqual напрямую — он должен корректно сравнивать пароль с хешем
	if !hash.IsEqual(password) {
		t.Error("IsEqual() returned false for correct password")
	}

	if hash.IsEqual("wrongPassword") {
		t.Error("IsEqual() returned true for incorrect password")
	}
}

func TestPasswordHash_RawFormat(t *testing.T) {
	password := "rawFormatTest456"
	hash, err := NewPasswordHash(password)
	if err != nil {
		t.Fatalf("Failed to create hash: %v", err)
	}

	// Проверяем, что хеш хранится в оригинальном bcrypt-формате (начинается с $2a$)
	hashStr := string(hash)
	if len(hashStr) < 60 || hashStr[:4] != "$2a$" {
		t.Errorf("PasswordHash should store raw bcrypt format, got: %s", hashStr)
	}

	// Проверяем совместимость с bcrypt.CompareHashAndPassword напрямую
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		t.Errorf("bcrypt.CompareHashAndPassword failed with raw hash: %v", err)
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