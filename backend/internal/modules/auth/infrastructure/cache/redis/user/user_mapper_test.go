package userredis

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/devathh/hive/backend/internal/modules/auth/domain/user"
	"github.com/google/uuid"
)

func TestDomainToRecord(t *testing.T) {
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

		bytes, err := domainToRecord(domainUser)
		if err != nil {
			t.Fatalf("domainToRecord() error = %v", err)
		}

		// Parse JSON to verify structure
		var record userRecord
		if err := json.Unmarshal(bytes, &record); err != nil {
			t.Fatalf("failed to unmarshal result: %v", err)
		}

		if record.ID != id {
			t.Errorf("ID = %v, want %v", record.ID, id)
		}
		if record.Firstname != "John" {
			t.Errorf("Firstname = %v, want %v", record.Firstname, "John")
		}
		if record.Lastname != "Doe" {
			t.Errorf("Lastname = %v, want %v", record.Lastname, "Doe")
		}
		if record.ConfirmedEmail != "john@example.com" {
			t.Errorf("ConfirmedEmail = %v, want %v", record.ConfirmedEmail, "john@example.com")
		}
		// PasswordHash should be omitted when empty due to omitempty tag
		if record.PasswordHash != "" {
			t.Errorf("PasswordHash = %v, want empty (omitted from cache)", record.PasswordHash)
		}
		if !record.PaidUntil.Equal(paidUntil) {
			t.Errorf("PaidUntil = %v, want %v", record.PaidUntil, paidUntil)
		}
		if record.OnVacation != true {
			t.Errorf("OnVacation = %v, want %v", record.OnVacation, true)
		}
		if record.TgChatID != 123456789 {
			t.Errorf("TgChatID = %v, want %v", record.TgChatID, 123456789)
		}
		if !record.Notification.TelegramEnable {
			t.Error("Notification.TelegramEnable = false, want true")
		}
		if !record.CreatedAt.Equal(now) {
			t.Errorf("CreatedAt = %v, want %v", record.CreatedAt, now)
		}
		if !record.UpdatedAt.Equal(now.Add(time.Hour)) {
			t.Errorf("UpdatedAt = %v, want %v", record.UpdatedAt, now.Add(time.Hour))
		}
	})

	t.Run("password_hash omitted due to omitempty", func(t *testing.T) {
		id := uuid.New()
		now := time.Now().UTC()

		domainUser := user.From(
			id,
			"Jane",
			"",
			user.Email("jane@example.com"),
			user.PasswordHash(""),
			now,
			false,
			0,
			user.NewNotification(false),
			now,
			now,
		)

		bytes, err := domainToRecord(domainUser)
		if err != nil {
			t.Fatalf("domainToRecord() error = %v", err)
		}

		// Verify password_hash field is NOT present in JSON when empty
		if bytesContain(bytes, `"password_hash"`) {
			t.Error("domainToRecord() should omit password_hash field when empty (omitempty)")
		}
	})
}

func TestRecordToDomain(t *testing.T) {
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

		domainUser, err := recordToDomain(jsonBytes)
		if err != nil {
			t.Fatalf("recordToDomain() error = %v", err)
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
		// PasswordHash should be empty when loading from standard cache
		if domainUser.PasswordHash().String() != "" {
			t.Errorf("PasswordHash() = %v, want empty (not in standard cache)", domainUser.PasswordHash())
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
}

func TestDomainToPrepared(t *testing.T) {
	t.Run("domain + generated password to prepared bytes", func(t *testing.T) {
		id := uuid.New()
		now := time.Now().UTC()
		genPassword := "TempPass123!"

		domainUser := user.From(
			id,
			"NewUser",
			"Test",
			user.Email("new@test.com"),
			user.PasswordHash("$2a$10$hash"),
			now,
			false,
			0,
			user.NewNotification(false),
			now,
			now,
		)

		bytes, err := domainToPrepared(domainUser, genPassword)
		if err != nil {
			t.Fatalf("domainToPrepared() error = %v", err)
		}

		// Parse to verify structure
		var prepared preparedUserRecord
		if err := json.Unmarshal(bytes, &prepared); err != nil {
			t.Fatalf("failed to unmarshal prepared: %v", err)
		}

		if prepared.User.ID != id {
			t.Errorf("User.ID = %v, want %v", prepared.User.ID, id)
		}
		if prepared.GeneratedPassword != genPassword {
			t.Errorf("GeneratedPassword = %v, want %v", prepared.GeneratedPassword, genPassword)
		}
	})

	t.Run("empty generated password is omitted", func(t *testing.T) {
		id := uuid.New()
		now := time.Now().UTC()

		domainUser := user.From(
			id,
			"User",
			"NoPass",
			user.Email("user@test.com"),
			user.PasswordHash("$2a$10$hash"),
			now,
			false,
			0,
			user.NewNotification(false),
			now,
			now,
		)

		bytes, err := domainToPrepared(domainUser, "")
		if err != nil {
			t.Fatalf("domainToPrepared() error = %v", err)
		}

		// Verify generated_password field is omitted when empty
		if bytesContain(bytes, `"generated_password"`) {
			t.Error("domainToPrepared() should omit generated_password when empty (omitempty)")
		}
	})
}

func TestPreparedToDomain(t *testing.T) {
	t.Run("prepared bytes to domain + generated password", func(t *testing.T) {
		id := uuid.New()
		now := time.Now().UTC()
		expectedGenPass := "GeneratedPass456!"

		preparedJSON := []byte(`{
			"user_record":{
				"id":"` + id.String() + `",
				"firstname":"Prepared",
				"lastname":"User",
				"confirmed_email":"prepared@test.com",
				"password_hash":"$2a$10$fromprepared",
				"paid_until":"` + now.Format(time.RFC3339Nano) + `",
				"on_vacation":true,
				"tg_chat_id":111222333,
				"notification":{"telegram_enable":true},
				"created_at":"` + now.Format(time.RFC3339Nano) + `",
				"updated_at":"` + now.Add(15*time.Minute).Format(time.RFC3339Nano) + `"
			},
			"generated_password":"` + expectedGenPass + `"
		}`)

		domainUser, genPass, err := preparedToDomain(preparedJSON)
		if err != nil {
			t.Fatalf("preparedToDomain() error = %v", err)
		}

		if domainUser.ID() != id {
			t.Errorf("ID() = %v, want %v", domainUser.ID(), id)
		}
		if domainUser.Firstname() != "Prepared" {
			t.Errorf("Firstname() = %v, want Prepared", domainUser.Firstname())
		}
		if domainUser.PasswordHash().String() != "$2a$10$fromprepared" {
			t.Errorf("PasswordHash() = %v, want $2a$10$fromprepared", domainUser.PasswordHash())
		}
		if domainUser.TgChatID() != 111222333 {
			t.Errorf("TgChatID() = %v, want 111222333", domainUser.TgChatID())
		}
		if !domainUser.Notification().IsTelegramEnable() {
			t.Error("Notification().IsTelegramEnable() = false, want true")
		}
		if genPass != expectedGenPass {
			t.Errorf("returned generated_password = %v, want %v", genPass, expectedGenPass)
		}
	})

	t.Run("without generated_password field", func(t *testing.T) {
		id := uuid.New()
		now := time.Now().UTC()

		preparedJSON := []byte(`{
			"user_record":{
				"id":"` + id.String() + `",
				"firstname":"NoGen",
				"lastname":"Pass",
				"confirmed_email":"nogen@test.com",
				"password_hash":"$2a$10$hash",
				"paid_until":"` + now.Format(time.RFC3339Nano) + `",
				"on_vacation":false,
				"tg_chat_id":0,
				"notification":{"telegram_enable":false},
				"created_at":"` + now.Format(time.RFC3339Nano) + `",
				"updated_at":"` + now.Format(time.RFC3339Nano) + `"
			}
		}`)

		domainUser, genPass, err := preparedToDomain(preparedJSON)
		if err != nil {
			t.Fatalf("preparedToDomain() error = %v", err)
		}

		if domainUser.Firstname() != "NoGen" {
			t.Errorf("Firstname() = %v, want NoGen", domainUser.Firstname())
		}
		if genPass != "" {
			t.Errorf("returned generated_password = %v, want empty", genPass)
		}
	})
}

func TestRoundTrip_StandardCache(t *testing.T) {
	t.Run("domain → record → domain preserves non-sensitive fields", func(t *testing.T) {
		originalID := uuid.New()
		now := time.Now().UTC().Truncate(time.Microsecond)
		paidUntil := now.AddDate(0, 6, 0)

		originalUser := user.From(
			originalID,
			"RoundTrip",
			"Test",
			user.Email("roundtrip@test.com"),
			user.PasswordHash("$2a$10$original_hash"),
			paidUntil,
			true,
			555123456,
			user.NewNotification(true),
			now,
			now.Add(30*time.Minute),
		)

		bytes, err := domainToRecord(originalUser)
		if err != nil {
			t.Fatalf("domainToRecord() error = %v", err)
		}

		restoredUser, err := recordToDomain(bytes)
		if err != nil {
			t.Fatalf("recordToDomain() error = %v", err)
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
		// Standard cache does NOT preserve password_hash
		if restoredUser.PasswordHash().String() != "" {
			t.Errorf("PasswordHash should be empty after standard cache round-trip")
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

func TestRoundTrip_PreparedCache(t *testing.T) {
	t.Run("domain → prepared → domain preserves ALL fields including password", func(t *testing.T) {
		originalID := uuid.New()
		now := time.Now().UTC().Truncate(time.Microsecond)
		originalHash := user.PasswordHash("$2a$10$prepared_hash")
		genPass := "TempGen123!"

		originalUser := user.From(
			originalID,
			"PreparedRound",
			"Trip",
			user.Email("prepared@test.com"),
			originalHash,
			now,
			false,
			999888777,
			user.NewNotification(false),
			now,
			now.Add(10*time.Minute),
		)

		bytes, err := domainToPrepared(originalUser, genPass)
		if err != nil {
			t.Fatalf("domainToPrepared() error = %v", err)
		}

		restoredUser, restoredGenPass, err := preparedToDomain(bytes)
		if err != nil {
			t.Fatalf("preparedToDomain() error = %v", err)
		}

		// ALL fields should match including password_hash
		if restoredUser.ID() != originalUser.ID() {
			t.Errorf("ID mismatch")
		}
		if restoredUser.PasswordHash() != originalHash {
			t.Errorf("PasswordHash mismatch: got %v, want %v", restoredUser.PasswordHash(), originalHash)
		}
		if restoredUser.TgChatID() != 999888777 {
			t.Errorf("TgChatID mismatch")
		}
		if restoredGenPass != genPass {
			t.Errorf("GeneratedPassword mismatch: got %v, want %v", restoredGenPass, genPass)
		}
	})
}

func TestInvalidJSON(t *testing.T) {
	t.Run("recordToDomain with invalid json", func(t *testing.T) {
		_, err := recordToDomain([]byte(`{invalid}`))
		if err == nil {
			t.Error("recordToDomain() should return error for invalid JSON")
		}
	})

	t.Run("preparedToDomain with invalid json", func(t *testing.T) {
		_, _, err := preparedToDomain([]byte(`{invalid}`))
		if err == nil {
			t.Error("preparedToDomain() should return error for invalid JSON")
		}
	})

	t.Run("preparedToDomain with malformed user_record", func(t *testing.T) {
		_, _, err := preparedToDomain([]byte(`{"user_record":{invalid}}`))
		if err == nil {
			t.Error("preparedToDomain() should return error for malformed user_record")
		}
	})
}

// Helper: check if byte slice contains substring
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
