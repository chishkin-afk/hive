package userpg

import (
	"testing"
	"time"

	"github.com/devathh/hive/backend/internal/modules/auth/domain/user"
	"github.com/google/uuid"
)

func TestToRecord(t *testing.T) {
	t.Run("full domain to record conversion", func(t *testing.T) {
		id := uuid.New()
		now := time.Now().UTC()
		paidUntil := now.AddDate(1, 0, 0)

		domainUser := user.From(
			id,
			"John",
			"Doe",
			user.Email("john@example.com"),
			user.PhoneNumber("779998887766"),
			user.PasswordHash("$2a$10$abc123"),
			paidUntil,
			true,
			123456789,
			user.NewNotification(true),
			now,
			now.Add(time.Hour),
		)

		record, err := toRecord(domainUser)
		if err != nil {
			t.Fatalf("toRecord() error = %v", err)
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
		if record.PasswordHash != "$2a$10$abc123" {
			t.Errorf("PasswordHash = %v, want %v", record.PasswordHash, "$2a$10$abc123")
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
		if string(record.Notification) != `{"telegram_enable":true}` {
			t.Errorf("Notification = %v, want %v", string(record.Notification), `{"telegram_enable":true}`)
		}
		if !record.CreatedAt.Equal(now) {
			t.Errorf("CreatedAt = %v, want %v", record.CreatedAt, now)
		}
		if !record.UpdatedAt.Equal(now.Add(time.Hour)) {
			t.Errorf("UpdatedAt = %v, want %v", record.UpdatedAt, now.Add(time.Hour))
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
			user.PhoneNumber("79998887766"),
			user.PasswordHash("$2a$10$xyz789"),
			now,
			false,
			0,
			user.NewNotification(false),
			now,
			now,
		)

		record, err := toRecord(domainUser)
		if err != nil {
			t.Fatalf("toRecord() error = %v", err)
		}

		if record.Lastname != "" {
			t.Errorf("Lastname = %v, want empty", record.Lastname)
		}
		if string(record.Notification) != `{"telegram_enable":false}` {
			t.Errorf("Notification = %v, want %v", string(record.Notification), `{"telegram_enable":false}`)
		}
		if record.TgChatID != 0 {
			t.Errorf("TgChatID = %v, want 0", record.TgChatID)
		}
	})
}

func TestToDomain(t *testing.T) {
	t.Run("full record to domain conversion", func(t *testing.T) {
		id := uuid.New()
		now := time.Now().UTC()
		paidUntil := now.AddDate(1, 0, 0)

		record := &userRecord{
			ID:             id,
			Firstname:      "Alice",
			Lastname:       "Smith",
			ConfirmedEmail: "alice@example.com",
			PasswordHash:   "$2a$10$def456",
			PaidUntil:      paidUntil,
			OnVacation:     false,
			TgChatID:       987654321,
			Notification:   []byte(`{"telegram_enable":true}`),
			CreatedAt:      now,
			UpdatedAt:      now.Add(2 * time.Hour),
		}

		domainUser, err := toDomain(record)
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
		if domainUser.PasswordHash().String() != "$2a$10$def456" {
			t.Errorf("PasswordHash() = %v, want %v", domainUser.PasswordHash(), "$2a$10$def456")
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

		record := &userRecord{
			ID:             id,
			Firstname:      "Bob",
			Lastname:       "",
			ConfirmedEmail: "bob@example.com",
			PasswordHash:   "$2a$10$ghi012",
			PaidUntil:      now,
			OnVacation:     true,
			TgChatID:       0,
			Notification:   []byte(`{"telegram_enable":false}`),
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		domainUser, err := toDomain(record)
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
	})
}

func TestRoundTripConversion(t *testing.T) {
	t.Run("domain - record - domain preserves all fields", func(t *testing.T) {
		originalID := uuid.New()
		now := time.Now().UTC().Truncate(time.Microsecond)
		paidUntil := now.AddDate(0, 6, 0)

		originalUser := user.From(
			originalID,
			"Test",
			"User",
			user.Email("test.user@example.com"),
			user.PhoneNumber("79998887766"),
			user.PasswordHash("$2a$10$roundtrip"),
			paidUntil,
			true,
			555123456,
			user.NewNotification(true),
			now,
			now.Add(30*time.Minute),
		)

		record, err := toRecord(originalUser)
		if err != nil {
			t.Fatalf("toRecord() error = %v", err)
		}

		restoredUser, err := toDomain(record)
		if err != nil {
			t.Fatalf("toDomain() error = %v", err)
		}

		if restoredUser.ID() != originalUser.ID() {
			t.Errorf("ID mismatch: got %v, want %v", restoredUser.ID(), originalUser.ID())
		}
		if restoredUser.Firstname() != originalUser.Firstname() {
			t.Errorf("Firstname mismatch: got %v, want %v", restoredUser.Firstname(), originalUser.Firstname())
		}
		if restoredUser.Lastname() != originalUser.Lastname() {
			t.Errorf("Lastname mismatch: got %v, want %v", restoredUser.Lastname(), originalUser.Lastname())
		}
		if restoredUser.ConfirmedEmail() != originalUser.ConfirmedEmail() {
			t.Errorf("Email mismatch: got %v, want %v", restoredUser.ConfirmedEmail(), originalUser.ConfirmedEmail())
		}
		if restoredUser.PasswordHash() != originalUser.PasswordHash() {
			t.Errorf("PasswordHash mismatch: got %v, want %v", restoredUser.PasswordHash(), originalUser.PasswordHash())
		}
		if !restoredUser.PaidUntil().Equal(originalUser.PaidUntil()) {
			t.Errorf("PaidUntil mismatch: got %v, want %v", restoredUser.PaidUntil(), originalUser.PaidUntil())
		}
		if restoredUser.OnVacation() != originalUser.OnVacation() {
			t.Errorf("OnVacation mismatch: got %v, want %v", restoredUser.OnVacation(), originalUser.OnVacation())
		}
		if restoredUser.TgChatID() != originalUser.TgChatID() {
			t.Errorf("TgChatID mismatch: got %v, want %v", restoredUser.TgChatID(), originalUser.TgChatID())
		}
		if restoredUser.Notification().IsTelegramEnable() != originalUser.Notification().IsTelegramEnable() {
			t.Errorf("Notification mismatch: got %v, want %v", restoredUser.Notification().IsTelegramEnable(), originalUser.Notification().IsTelegramEnable())
		}
		if !restoredUser.CreatedAt().Equal(originalUser.CreatedAt()) {
			t.Errorf("CreatedAt mismatch: got %v, want %v", restoredUser.CreatedAt(), originalUser.CreatedAt())
		}
		if !restoredUser.UpdatedAt().Equal(originalUser.UpdatedAt()) {
			t.Errorf("UpdatedAt mismatch: got %v, want %v", restoredUser.UpdatedAt(), originalUser.UpdatedAt())
		}
	})
}

func TestToDomain_InvalidJSON(t *testing.T) {
	t.Run("malformed notification JSON returns error", func(t *testing.T) {
		record := &userRecord{
			ID:             uuid.New(),
			Firstname:      "Bad",
			Lastname:       "JSON",
			ConfirmedEmail: "bad@example.com",
			PasswordHash:   "$2a$10$bad",
			PaidUntil:      time.Now(),
			OnVacation:     false,
			TgChatID:       0,
			Notification:   []byte(`{invalid json}`),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		_, err := toDomain(record)
		if err == nil {
			t.Error("toDomain() should return error for invalid JSON")
		}
	})

	t.Run("empty notification bytes should work with defaults", func(t *testing.T) {
		record := &userRecord{
			ID:             uuid.New(),
			Firstname:      "Empty",
			Lastname:       "Notif",
			ConfirmedEmail: "empty@example.com",
			PasswordHash:   "$2a$10$empty",
			PaidUntil:      time.Now(),
			OnVacation:     false,
			TgChatID:       0,
			Notification:   []byte(`{}`),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		domainUser, err := toDomain(record)
		if err != nil {
			t.Fatalf("toDomain() error = %v", err)
		}
		if domainUser.Notification().IsTelegramEnable() != false {
			t.Error("Expected telegram_enable=false for empty JSON object")
		}
	})
}
