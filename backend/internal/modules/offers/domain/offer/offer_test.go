package offer

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewOffer(t *testing.T) {
	validAuthorID := uuid.New()
	validDuration := SlotDuration(30 * time.Minute)
	validInterval := SlotInterval(15 * time.Minute)

	t.Run("valid offer creation", func(t *testing.T) {
		offer, err := New(validAuthorID, "Test Offer", "Test description", validDuration, validInterval)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.ID() == uuid.Nil {
			t.Error("expected non-nil UUID")
		}
		if offer.AuthorID() != validAuthorID {
			t.Errorf("expected authorID %v, got %v", validAuthorID, offer.AuthorID())
		}
		if offer.Title() != "Test Offer" {
			t.Errorf("expected title 'Test Offer', got %v", offer.Title())
		}
		if offer.Description() != "Test description" {
			t.Errorf("expected description 'Test description', got %v", offer.Description())
		}
		if offer.SlotDuration() != validDuration {
			t.Errorf("expected slotDuration %v, got %v", validDuration, offer.SlotDuration())
		}
		if offer.SlotInterval() != validInterval {
			t.Errorf("expected slotInterval %v, got %v", validInterval, offer.SlotInterval())
		}
		if !offer.IsActive().Bool() {
			t.Error("expected offer to be active by default")
		}
		if offer.CreatedAt().IsZero() {
			t.Error("expected non-zero CreatedAt")
		}
		if offer.UpdatedAt().IsZero() {
			t.Error("expected non-zero UpdatedAt")
		}
	})

	t.Run("empty description is allowed", func(t *testing.T) {
		offer, err := New(validAuthorID, "Test Offer", "", validDuration, validInterval)
		if err != nil {
			t.Fatalf("unexpected error with empty description: %v", err)
		}
		if offer.Description() != "" {
			t.Errorf("expected empty description, got %v", offer.Description())
		}
	})

	t.Run("invalid author ID", func(t *testing.T) {
		_, err := New(uuid.Nil, "Test Offer", "description", validDuration, validInterval)
		if err == nil {
			t.Error("expected error for nil author ID")
		}
		if err != ErrInvalidAuthorID {
			t.Errorf("expected ErrInvalidAuthorID, got %v", err)
		}
	})

	t.Run("title too short", func(t *testing.T) {
		_, err := New(validAuthorID, "ab", "description", validDuration, validInterval)
		if err == nil {
			t.Error("expected error for short title")
		}
		if err != ErrInvalidTitle {
			t.Errorf("expected ErrInvalidTitle, got %v", err)
		}
	})

	t.Run("title too long", func(t *testing.T) {
		_, err := New(validAuthorID, strings.Repeat("a", 129), "description", validDuration, validInterval)
		if err == nil {
			t.Error("expected error for long title")
		}
		if err != ErrInvalidTitle {
			t.Errorf("expected ErrInvalidTitle, got %v", err)
		}
	})

	t.Run("title with spaces trimmed", func(t *testing.T) {
		offer, err := New(validAuthorID, "  valid title  ", "description", validDuration, validInterval)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.Title() != "valid title" {
			t.Errorf("expected trimmed title 'valid title', got %v", offer.Title())
		}
	})

	t.Run("description too long", func(t *testing.T) {
		_, err := New(validAuthorID, "Test Offer", strings.Repeat("a", 513), validDuration, validInterval)
		if err == nil {
			t.Error("expected error for long description")
		}
		if err != ErrInvalidDescription {
			t.Errorf("expected ErrInvalidDescription, got %v", err)
		}
	})

	t.Run("description with spaces trimmed", func(t *testing.T) {
		offer, err := New(validAuthorID, "Test Offer", "  valid desc  ", validDuration, validInterval)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.Description() != "valid desc" {
			t.Errorf("expected trimmed description 'valid desc', got %v", offer.Description())
		}
	})

	t.Run("invalid slot duration", func(t *testing.T) {
		_, err := New(validAuthorID, "Test Offer", "description", SlotDuration(24*time.Hour), validInterval)
		if err == nil {
			t.Error("expected error for invalid slot duration")
		}
		if err != ErrInvalidSlotDuration {
			t.Errorf("expected ErrInvalidSlotDuration, got %v", err)
		}
	})

	t.Run("invalid slot interval", func(t *testing.T) {
		_, err := New(validAuthorID, "Test Offer", "description", validDuration, SlotInterval(24*time.Hour))
		if err == nil {
			t.Error("expected error for invalid slot interval")
		}
		if err != ErrInvalidSlotInterval {
			t.Errorf("expected ErrInvalidSlotInterval, got %v", err)
		}
	})
}

func TestOffer_ChangeTitle(t *testing.T) {
	validAuthorID := uuid.New()
	validDuration := SlotDuration(30 * time.Minute)
	validInterval := SlotInterval(15 * time.Minute)
	offer, _ := New(validAuthorID, "Old Title", "description", validDuration, validInterval)

	t.Run("valid title change", func(t *testing.T) {
		oldUpdatedAt := offer.UpdatedAt()
		time.Sleep(10 * time.Millisecond)

		err := offer.ChangeTitle("New Title")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.Title() != "New Title" {
			t.Errorf("expected title 'New Title', got %v", offer.Title())
		}
		if !offer.UpdatedAt().After(oldUpdatedAt) {
			t.Error("expected UpdatedAt to be updated")
		}
	})

	t.Run("title too short", func(t *testing.T) {
		err := offer.ChangeTitle("ab")
		if err == nil {
			t.Error("expected error for short title")
		}
		if err != ErrInvalidTitle {
			t.Errorf("expected ErrInvalidTitle, got %v", err)
		}
	})

	t.Run("title with spaces trimmed", func(t *testing.T) {
		err := offer.ChangeTitle("  trimmed  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.Title() != "trimmed" {
			t.Errorf("expected trimmed title 'trimmed', got %v", offer.Title())
		}
	})
}

func TestOffer_ChangeDescription(t *testing.T) {
	validAuthorID := uuid.New()
	validDuration := SlotDuration(30 * time.Minute)
	validInterval := SlotInterval(15 * time.Minute)
	offer, _ := New(validAuthorID, "Title", "Old description", validDuration, validInterval)

	t.Run("valid description change", func(t *testing.T) {
		oldUpdatedAt := offer.UpdatedAt()
		time.Sleep(10 * time.Millisecond)

		err := offer.ChangeDescription("New description")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.Description() != "New description" {
			t.Errorf("expected description 'New description', got %v", offer.Description())
		}
		if !offer.UpdatedAt().After(oldUpdatedAt) {
			t.Error("expected UpdatedAt to be updated")
		}
	})

	t.Run("empty description is allowed", func(t *testing.T) {
		err := offer.ChangeDescription("")
		if err != nil {
			t.Fatalf("unexpected error with empty description: %v", err)
		}
		if offer.Description() != "" {
			t.Errorf("expected empty description, got %v", offer.Description())
		}
	})

	t.Run("description too long", func(t *testing.T) {
		err := offer.ChangeDescription(strings.Repeat("a", 513))
		if err == nil {
			t.Error("expected error for long description")
		}
		if err != ErrInvalidDescription {
			t.Errorf("expected ErrInvalidDescription, got %v", err)
		}
	})

	t.Run("description with spaces trimmed", func(t *testing.T) {
		err := offer.ChangeDescription("  trimmed desc  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.Description() != "trimmed desc" {
			t.Errorf("expected trimmed description 'trimmed desc', got %v", offer.Description())
		}
	})
}

func TestOffer_ChangeSlotDuration(t *testing.T) {
	validAuthorID := uuid.New()
	validDuration := SlotDuration(30 * time.Minute)
	validInterval := SlotInterval(15 * time.Minute)
	offer, _ := New(validAuthorID, "Title", "description", validDuration, validInterval)

	t.Run("valid slot duration change", func(t *testing.T) {
		oldUpdatedAt := offer.UpdatedAt()
		time.Sleep(10 * time.Millisecond)
		newDuration := SlotDuration(45 * time.Minute)

		err := offer.ChangeSlotDuration(newDuration)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.SlotDuration() != newDuration {
			t.Errorf("expected slotDuration %v, got %v", newDuration, offer.SlotDuration())
		}
		if !offer.UpdatedAt().After(oldUpdatedAt) {
			t.Error("expected UpdatedAt to be updated")
		}
	})

	t.Run("invalid slot duration", func(t *testing.T) {
		err := offer.ChangeSlotDuration(SlotDuration(24 * time.Hour))
		if err == nil {
			t.Error("expected error for invalid slot duration")
		}
		if err != ErrInvalidSlotDuration {
			t.Errorf("expected ErrInvalidSlotDuration, got %v", err)
		}
	})
}

func TestOffer_ChangeSlotInterval(t *testing.T) {
	validAuthorID := uuid.New()
	validDuration := SlotDuration(30 * time.Minute)
	validInterval := SlotInterval(15 * time.Minute)
	offer, _ := New(validAuthorID, "Title", "description", validDuration, validInterval)

	t.Run("valid slot interval change", func(t *testing.T) {
		oldUpdatedAt := offer.UpdatedAt()
		time.Sleep(10 * time.Millisecond)
		newInterval := SlotInterval(30 * time.Minute)

		err := offer.ChangeSlotInterval(newInterval)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.SlotInterval() != newInterval {
			t.Errorf("expected slotInterval %v, got %v", newInterval, offer.SlotInterval())
		}
		if !offer.UpdatedAt().After(oldUpdatedAt) {
			t.Error("expected UpdatedAt to be updated")
		}
	})

	t.Run("invalid slot interval", func(t *testing.T) {
		err := offer.ChangeSlotInterval(SlotInterval(24 * time.Hour))
		if err == nil {
			t.Error("expected error for invalid slot interval")
		}
		if err != ErrInvalidSlotInterval {
			t.Errorf("expected ErrInvalidSlotInterval, got %v", err)
		}
	})
}

func TestOffer_EnableDisable(t *testing.T) {
	validAuthorID := uuid.New()
	validDuration := SlotDuration(30 * time.Minute)
	validInterval := SlotInterval(15 * time.Minute)

	t.Run("disable active offer", func(t *testing.T) {
		offer, _ := New(validAuthorID, "Title", "description", validDuration, validInterval)
		oldUpdatedAt := offer.UpdatedAt()
		time.Sleep(10 * time.Millisecond)

		err := offer.Disable()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.IsActive().Bool() {
			t.Error("expected offer to be disabled")
		}
		if !offer.UpdatedAt().After(oldUpdatedAt) {
			t.Error("expected UpdatedAt to be updated")
		}
	})

	t.Run("disable already disabled offer", func(t *testing.T) {
		offer, _ := New(validAuthorID, "Title", "description", validDuration, validInterval)
		_ = offer.Disable()

		err := offer.Disable()
		if err == nil {
			t.Error("expected error when disabling already disabled offer")
		}
		if err != ErrOfferAlreadyDisabled {
			t.Errorf("expected ErrOfferAlreadyDisabled, got %v", err)
		}
	})

	t.Run("enable disabled offer", func(t *testing.T) {
		offer, _ := New(validAuthorID, "Title", "description", validDuration, validInterval)
		_ = offer.Disable()
		oldUpdatedAt := offer.UpdatedAt()
		time.Sleep(10 * time.Millisecond)

		err := offer.Enable()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !offer.IsActive().Bool() {
			t.Error("expected offer to be enabled")
		}
		if !offer.UpdatedAt().After(oldUpdatedAt) {
			t.Error("expected UpdatedAt to be updated")
		}
	})

	t.Run("enable already enabled offer", func(t *testing.T) {
		offer, _ := New(validAuthorID, "Title", "description", validDuration, validInterval)

		err := offer.Enable()
		if err == nil {
			t.Error("expected error when enabling already enabled offer")
		}
		if err != ErrOfferAlreadyEnabled {
			t.Errorf("expected ErrOfferAlreadyEnabled, got %v", err)
		}
	})
}

func TestOffer_Getters(t *testing.T) {
	validAuthorID := uuid.New()
	validDuration := SlotDuration(30 * time.Minute)
	validInterval := SlotInterval(15 * time.Minute)
	offer, _ := New(validAuthorID, "Title", "description", validDuration, validInterval)

	if offer.ID() == uuid.Nil {
		t.Error("ID() returned nil UUID")
	}
	if offer.AuthorID() != validAuthorID {
		t.Errorf("AuthorID() = %v, want %v", offer.AuthorID(), validAuthorID)
	}
	if offer.Title() != "Title" {
		t.Errorf("Title() = %v, want 'Title'", offer.Title())
	}
	if offer.Description() != "description" {
		t.Errorf("Description() = %v, want 'description'", offer.Description())
	}
	if offer.SlotDuration() != validDuration {
		t.Errorf("SlotDuration() = %v, want %v", offer.SlotDuration(), validDuration)
	}
	if offer.SlotInterval() != validInterval {
		t.Errorf("SlotInterval() = %v, want %v", offer.SlotInterval(), validInterval)
	}
	if !offer.IsActive().Bool() {
		t.Error("IsActive() returned false, expected true")
	}
	if offer.CreatedAt().IsZero() {
		t.Error("CreatedAt() returned zero time")
	}
	if offer.UpdatedAt().IsZero() {
		t.Error("UpdatedAt() returned zero time")
	}
}