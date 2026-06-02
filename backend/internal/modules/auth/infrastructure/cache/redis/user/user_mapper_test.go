package userredis

import (
	"testing"
	"time"

	"github.com/devathh/hive/backend/internal/modules/auth/domain/user"
	"github.com/google/uuid"
)

func TestToBytes(t *testing.T) {
	t.Run("full domain to json bytes conversion", func(t *testing.T) {
		id := uuid.New()
		now := time.Now().UTC()
		paidUntil := now.AddDate(1, 0, 0)

		domainUser := user.From(
			id,
			"John",
			"Doe",
			user.Email("john@example.com"),
			user.PasswordHash("$2a$10$should_not_be_cached"),
			paidUntil,
			true,
			123456789,
			user.NewNotification(true),
			now,
			now.Add(time.Hour),
		)

		bytes, err := toBytes(domainUser)
		if err != nil {
			t.Fatalf("toBytes() error = %v", err)
		}

		expectedJSON := `{"id":"` + id.String() + `","firstname":"John","lastname":"Doe","confirmed_email":"john@example.com","paid_until":"` + paidUntil.Format(time.RFC3339Nano) + `","on_vacation":true,"tg_chat_id":123456789,"notification":{"telegram_enable":true},"created_at":"` + now.Format(time.RFC3339Nano) + `","updated_at":"` + now.Add(time.Hour).Format(time.RFC3339Nano) + `"}`

		if string(bytes) != expectedJSON {
			t.Errorf("toBytes() JSON mismatch\ngot:  %s\nwant: %s", string(bytes), expectedJSON)
		}

		// Critical: password hash MUST NOT be in cached bytes
		if bytesContain(bytes, "password") || bytesContain(bytes, "hash") {
			t.Error("toBytes() should NOT include password_hash in cached data")
		}
	})

	t.Run("empty lastname and disabled notification", func(t *testing.T) {
		id := uuid.New()
		now := time.Now().UTC()

		domainUser := user.From(
			id,
			"Jane",
			"",
			user.Email("jane@example.com"),
			user.PasswordHash("$2a$10$secret"),
			now,
			false,
			0,
			user.NewNotification(false),
			now,
			now,
		)

		bytes, err := toBytes(domainUser)
		if err != nil {
			t.Fatalf("toBytes() error = %v", err)
		}

		if !bytesContain(bytes, `"lastname":""`) {
			t.Error("toBytes() should include empty lastname")
		}
		if !bytesContain(bytes, `"telegram_enable":false`) {
			t.Error("toBytes() should include telegram_enable:false")
		}
		if bytesContain(bytes, "password") {
			t.Error("toBytes() should NOT cache password hash")
		}
	})
}

func TestToDomain(t *testing.T) {
	t.Run("full json bytes to domain conversion", func(t *testing.T) {
		id := uuid.New()
		now := time.Now().UTC()
		paidUntil := now.AddDate(1, 0, 0)

		jsonBytes := []byte(`{
			"id":"` + id.String() + `",
			"firstname":"Alice",
			"lastname":"Smith",
			"confirmed_email":"alice@example.com",
			"paid_until":"` + paidUntil.Format(time.RFC3339Nano) + `",
			"on_vacation":false,
			"tg_chat_id":987654321,
			"notification":{"telegram_enable":true},
			"created_at":"` + now.Format(time.RFC3339Nano) + `",
			"updated_at":"` + now.Add(2*time.Hour).Format(time.RFC3339Nano) + `"
		}`)

		domainUser, err := toDomain(jsonBytes)
		if err != nil {
			t.Fatalf("toDomain() error = %v", err)
		}

		if domainUser.ID() != id {
			t.Errorf("ID() = %v, want %v", domainUser.ID(), id)
		}
		if domainUser.Firstname() != "Alice" {
			t.Errorf("Firstname() = %v, want %v", domainUser.Firstname(), "Alice")
		}
		if domainUser.Lastname() != "Smith" {
			t.Errorf("Lastname() = %v, want %v", domainUser.Lastname(), "Smith")
		}
		if domainUser.ConfirmedEmail().String() != "alice@example.com" {
			t.Errorf("ConfirmedEmail() = %v, want %v", domainUser.ConfirmedEmail(), "alice@example.com")
		}
		// Critical: password hash from cache should be empty (not stored)
		if domainUser.PasswordHash().String() != "" {
			t.Errorf("PasswordHash() = %v, want empty (not cached)", domainUser.PasswordHash())
		}
		if !domainUser.PaidUntil().Equal(paidUntil) {
			t.Errorf("PaidUntil() = %v, want %v", domainUser.PaidUntil(), paidUntil)
		}
		if domainUser.OnVacation() != false {
			t.Errorf("OnVacation() = %v, want %v", domainUser.OnVacation(), false)
		}
		if domainUser.TgChatID() != 987654321 {
			t.Errorf("TgChatID() = %v, want %v", domainUser.TgChatID(), 987654321)
		}
		if !domainUser.Notification().IsTelegramEnable() {
			t.Error("Notification().IsTelegramEnable() = false, want true")
		}
		if !domainUser.CreatedAt().Equal(now) {
			t.Errorf("CreatedAt() = %v, want %v", domainUser.CreatedAt(), now)
		}
		if !domainUser.UpdatedAt().Equal(now.Add(2 * time.Hour)) {
			t.Errorf("UpdatedAt() = %v, want %v", domainUser.UpdatedAt(), now.Add(2*time.Hour))
		}
	})

	t.Run("empty fields and disabled notification", func(t *testing.T) {
		id := uuid.New()
		now := time.Now().UTC()

		jsonBytes := []byte(`{
			"id":"` + id.String() + `",
			"firstname":"Bob",
			"lastname":"",
			"confirmed_email":"bob@example.com",
			"paid_until":"` + now.Format(time.RFC3339Nano) + `",
			"on_vacation":true,
			"tg_chat_id":0,
			"notification":{"telegram_enable":false},
			"created_at":"` + now.Format(time.RFC3339Nano) + `",
			"updated_at":"` + now.Format(time.RFC3339Nano) + `"
		}`)

		domainUser, err := toDomain(jsonBytes)
		if err != nil {
			t.Fatalf("toDomain() error = %v", err)
		}

		if domainUser.Lastname() != "" {
			t.Errorf("Lastname() = %v, want empty", domainUser.Lastname())
		}
		if domainUser.TgChatID() != 0 {
			t.Errorf("TgChatID() = %v, want 0", domainUser.TgChatID())
		}
		if domainUser.Notification().IsTelegramEnable() {
			t.Error("Notification().IsTelegramEnable() = true, want false")
		}
		if domainUser.PasswordHash().String() != "" {
			t.Error("PasswordHash should be empty when loading from cache")
		}
	})
}

func TestRoundTripConversion_Cache(t *testing.T) {
	t.Run("domain → bytes → domain preserves non-sensitive fields", func(t *testing.T) {
		originalID := uuid.New()
		now := time.Now().UTC().Truncate(time.Microsecond)
		paidUntil := now.AddDate(0, 6, 0)

		originalUser := user.From(
			originalID,
			"CacheTest",
			"User",
			user.Email("cache.test@example.com"),
			user.PasswordHash("$2a$10$secret_password_hash"),
			paidUntil,
			true,
			555123456,
			user.NewNotification(true),
			now,
			now.Add(30*time.Minute),
		)

		bytes, err := toBytes(originalUser)
		if err != nil {
			t.Fatalf("toBytes() error = %v", err)
		}

		restoredUser, err := toDomain(bytes)
		if err != nil {
			t.Fatalf("toDomain() error = %v", err)
		}

		// All fields except password should match
		if restoredUser.ID() != originalUser.ID() {
			t.Errorf("ID mismatch")
		}
		if restoredUser.Firstname() != originalUser.Firstname() {
			t.Errorf("Firstname mismatch")
		}
		if restoredUser.Lastname() != originalUser.Lastname() {
			t.Errorf("Lastname mismatch")
		}
		if restoredUser.ConfirmedEmail() != originalUser.ConfirmedEmail() {
			t.Errorf("Email mismatch")
		}
		// EXPECTED: password hash is NOT preserved in cache round-trip
		if restoredUser.PasswordHash().String() != "" {
			t.Errorf("PasswordHash should be empty after cache round-trip, got %v", restoredUser.PasswordHash())
		}
		if !restoredUser.PaidUntil().Equal(originalUser.PaidUntil()) {
			t.Errorf("PaidUntil mismatch")
		}
		if restoredUser.OnVacation() != originalUser.OnVacation() {
			t.Errorf("OnVacation mismatch")
		}
		if restoredUser.TgChatID() != originalUser.TgChatID() {
			t.Errorf("TgChatID mismatch")
		}
		if restoredUser.Notification().IsTelegramEnable() != originalUser.Notification().IsTelegramEnable() {
			t.Errorf("Notification mismatch")
		}
		if !restoredUser.CreatedAt().Equal(originalUser.CreatedAt()) {
			t.Errorf("CreatedAt mismatch")
		}
		if !restoredUser.UpdatedAt().Equal(originalUser.UpdatedAt()) {
			t.Errorf("UpdatedAt mismatch")
		}
	})
}

func TestToDomain_InvalidJSON(t *testing.T) {
	t.Run("malformed json returns error", func(t *testing.T) {
		invalidJSON := []byte(`{invalid json}`)
		_, err := toDomain(invalidJSON)
		if err == nil {
			t.Error("toDomain() should return error for invalid JSON")
		}
	})

	t.Run("missing required fields handled by domain constructor", func(t *testing.T) {
		// Empty JSON object - domain.From should handle defaults
		emptyJSON := []byte(`{}`)
		domainUser, err := toDomain(emptyJSON)
		if err != nil {
			t.Fatalf("toDomain() error = %v", err)
		}
		// Should return a user with zero values, not panic
		if domainUser == nil {
			t.Error("toDomain() should return non-nil user for empty JSON")
		}
	})
}

// Helper: check if byte slice contains substring (for JSON field checks)
func bytesContain(b []byte, substr string) bool {
	return len(b) >= len(substr) && string(b) != "" && indexOfString(string(b), substr) != -1
}

func indexOfString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}