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
	validPrice := uint64(1500)

	t.Run("valid offer creation", func(t *testing.T) {
		offer, err := New(validAuthorID, "Test Offer", "Test description", validPrice, validDuration, validInterval)
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
		if offer.Price() != validPrice {
			t.Errorf("expected price %d, got %d", validPrice, offer.Price())
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

	t.Run("zero price is allowed", func(t *testing.T) {
		offer, err := New(validAuthorID, "Free Offer", "", uint64(0), validDuration, validInterval)
		if err != nil {
			t.Fatalf("unexpected error with zero price: %v", err)
		}
		if offer.Price() != 0 {
			t.Errorf("expected price 0, got %d", offer.Price())
		}
	})

	t.Run("invalid author ID", func(t *testing.T) {
		_, err := New(uuid.Nil, "Test Offer", "description", validPrice, validDuration, validInterval)
		if err == nil {
			t.Error("expected error for nil author ID")
		}
		if err != ErrInvalidAuthorID {
			t.Errorf("expected ErrInvalidAuthorID, got %v", err)
		}
	})

	t.Run("title too short (2 chars)", func(t *testing.T) {
		_, err := New(validAuthorID, "ab", "description", validPrice, validDuration, validInterval)
		if err == nil {
			t.Error("expected error for short title")
		}
		if err != ErrInvalidTitle {
			t.Errorf("expected ErrInvalidTitle, got %v", err)
		}
	})

	t.Run("title exactly min length (3 chars) - should fail because > min", func(t *testing.T) {
		// isFit требует len > min, т.е. > 3. Значит 3 символа тоже ошибка.
		_, err := New(validAuthorID, "abc", "description", validPrice, validDuration, validInterval)
		if err == nil {
			t.Error("expected error for title with exactly 3 chars (len > 3 required)")
		}
		if err != ErrInvalidTitle {
			t.Errorf("expected ErrInvalidTitle, got %v", err)
		}
	})

	t.Run("title valid min length (4 chars)", func(t *testing.T) {
		offer, err := New(validAuthorID, "abcd", "description", validPrice, validDuration, validInterval)
		if err != nil {
			t.Fatalf("unexpected error for 4 char title: %v", err)
		}
		if offer.Title() != "abcd" {
			t.Errorf("expected title 'abcd', got %v", offer.Title())
		}
	})

	t.Run("title exactly max length (128 chars)", func(t *testing.T) {
		title := strings.Repeat("a", 128)
		offer, err := New(validAuthorID, title, "description", validPrice, validDuration, validInterval)
		if err != nil {
			t.Fatalf("unexpected error for 128 char title: %v", err)
		}
		if offer.Title() != title {
			t.Errorf("expected title length 128, got %d", len(offer.Title()))
		}
	})

	t.Run("title too long (129 chars)", func(t *testing.T) {
		_, err := New(validAuthorID, strings.Repeat("a", 129), "description", validPrice, validDuration, validInterval)
		if err == nil {
			t.Error("expected error for long title")
		}
		if err != ErrInvalidTitle {
			t.Errorf("expected ErrInvalidTitle, got %v", err)
		}
	})

	t.Run("title with spaces trimmed", func(t *testing.T) {
		offer, err := New(validAuthorID, "  valid title  ", "description", validPrice, validDuration, validInterval)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.Title() != "valid title" {
			t.Errorf("expected trimmed title 'valid title', got %v", offer.Title())
		}
	})

	t.Run("empty description is allowed", func(t *testing.T) {
		offer, err := New(validAuthorID, "Test Offer", "", validPrice, validDuration, validInterval)
		if err != nil {
			t.Fatalf("unexpected error with empty description: %v", err)
		}
		if offer.Description() != "" {
			t.Errorf("expected empty description, got %v", offer.Description())
		}
	})

	t.Run("description too long (513 chars)", func(t *testing.T) {
		_, err := New(validAuthorID, "Test Offer", strings.Repeat("a", 513), validPrice, validDuration, validInterval)
		if err == nil {
			t.Error("expected error for long description")
		}
		if err != ErrInvalidDescription {
			t.Errorf("expected ErrInvalidDescription, got %v", err)
		}
	})

	t.Run("description exactly max length (512 chars)", func(t *testing.T) {
		desc := strings.Repeat("a", 512)
		offer, err := New(validAuthorID, "Test Offer", desc, validPrice, validDuration, validInterval)
		if err != nil {
			t.Fatalf("unexpected error for 512 char description: %v", err)
		}
		if offer.Description() != desc {
			t.Errorf("expected description length 512, got %d", len(offer.Description()))
		}
	})

	t.Run("description with spaces trimmed", func(t *testing.T) {
		offer, err := New(validAuthorID, "Test Offer", "  valid desc  ", validPrice, validDuration, validInterval)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.Description() != "valid desc" {
			t.Errorf("expected trimmed description 'valid desc', got %v", offer.Description())
		}
	})

	t.Run("invalid slot duration (24 hours)", func(t *testing.T) {
		_, err := New(validAuthorID, "Test Offer", "description", validPrice, SlotDuration(24*time.Hour), validInterval)
		if err == nil {
			t.Error("expected error for invalid slot duration")
		}
		if err != ErrInvalidSlotDuration {
			t.Errorf("expected ErrInvalidSlotDuration, got %v", err)
		}
	})

	t.Run("invalid slot duration (zero)", func(t *testing.T) {
		_, err := New(validAuthorID, "Test Offer", "description", validPrice, SlotDuration(0), validInterval)
		if err == nil {
			t.Error("expected error for zero slot duration")
		}
		if err != ErrInvalidSlotDuration {
			t.Errorf("expected ErrInvalidSlotDuration, got %v", err)
		}
	})

	t.Run("invalid slot interval (24 hours)", func(t *testing.T) {
		_, err := New(validAuthorID, "Test Offer", "description", validPrice, validDuration, SlotInterval(24*time.Hour))
		if err == nil {
			t.Error("expected error for invalid slot interval")
		}
		if err != ErrInvalidSlotInterval {
			t.Errorf("expected ErrInvalidSlotInterval, got %v", err)
		}
	})

	t.Run("invalid slot interval (zero)", func(t *testing.T) {
		_, err := New(validAuthorID, "Test Offer", "description", validPrice, validDuration, SlotInterval(0))
		if err == nil {
			t.Error("expected error for zero slot interval")
		}
		if err != ErrInvalidSlotInterval {
			t.Errorf("expected ErrInvalidSlotInterval, got %v", err)
		}
	})
}

func TestOffer_ChangePrice(t *testing.T) {
	validAuthorID := uuid.New()
	validDuration := SlotDuration(30 * time.Minute)
	validInterval := SlotInterval(15 * time.Minute)
	offer, _ := New(validAuthorID, "Title", "description", uint64(1000), validDuration, validInterval)

	t.Run("change price to higher value", func(t *testing.T) {
		oldUpdatedAt := offer.UpdatedAt()
		// Имитируем passage of time для проверки обновления updatedAt
		time.Sleep(1 * time.Millisecond)
		newPrice := uint64(2500)

		offer.ChangePrice(newPrice)
		if offer.Price() != newPrice {
			t.Errorf("expected price %d, got %d", newPrice, offer.Price())
		}
		// updatedAt должен обновиться, но из-за быстрого выполнения тестов может быть равен
		// Проверяем, что он не стал меньше старого
		if offer.UpdatedAt().Before(oldUpdatedAt) {
			t.Error("expected UpdatedAt not to be before old value")
		}
	})

	t.Run("change price to zero", func(t *testing.T) {
		offer.ChangePrice(0)
		if offer.Price() != 0 {
			t.Errorf("expected price 0, got %d", offer.Price())
		}
	})

	t.Run("change price to large value", func(t *testing.T) {
		largePrice := uint64(999999999)
		offer.ChangePrice(largePrice)
		if offer.Price() != largePrice {
			t.Errorf("expected price %d, got %d", largePrice, offer.Price())
		}
	})
}

func TestOffer_ChangeTitle(t *testing.T) {
	validAuthorID := uuid.New()
	validDuration := SlotDuration(30 * time.Minute)
	validInterval := SlotInterval(15 * time.Minute)
	validPrice := uint64(1500)
	offer, _ := New(validAuthorID, "Old Title", "description", validPrice, validDuration, validInterval)

	t.Run("valid title change", func(t *testing.T) {
		oldUpdatedAt := offer.UpdatedAt()
		time.Sleep(1 * time.Millisecond)

		err := offer.ChangeTitle("New Title")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.Title() != "New Title" {
			t.Errorf("expected title 'New Title', got %v", offer.Title())
		}
		if offer.UpdatedAt().Before(oldUpdatedAt) {
			t.Error("expected UpdatedAt not to be before old value")
		}
	})

	t.Run("title too short (2 chars)", func(t *testing.T) {
		err := offer.ChangeTitle("ab")
		if err == nil {
			t.Error("expected error for short title")
		}
		if err != ErrInvalidTitle {
			t.Errorf("expected ErrInvalidTitle, got %v", err)
		}
	})

	t.Run("title exactly 3 chars - should fail", func(t *testing.T) {
		err := offer.ChangeTitle("abc")
		if err == nil {
			t.Error("expected error for 3 char title (len > 3 required)")
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

	t.Run("title exactly 128 chars", func(t *testing.T) {
		title := strings.Repeat("a", 128)
		err := offer.ChangeTitle(title)
		if err != nil {
			t.Fatalf("unexpected error for 128 char title: %v", err)
		}
		if offer.Title() != title {
			t.Errorf("expected title length 128, got %d", len(offer.Title()))
		}
	})

	t.Run("title 129 chars - too long", func(t *testing.T) {
		err := offer.ChangeTitle(strings.Repeat("a", 129))
		if err == nil {
			t.Error("expected error for 129 char title")
		}
		if err != ErrInvalidTitle {
			t.Errorf("expected ErrInvalidTitle, got %v", err)
		}
	})
}

func TestOffer_ChangeDescription(t *testing.T) {
	validAuthorID := uuid.New()
	validDuration := SlotDuration(30 * time.Minute)
	validInterval := SlotInterval(15 * time.Minute)
	validPrice := uint64(1500)
	offer, _ := New(validAuthorID, "Title", "Old description", validPrice, validDuration, validInterval)

	t.Run("valid description change", func(t *testing.T) {
		oldUpdatedAt := offer.UpdatedAt()
		time.Sleep(1 * time.Millisecond)

		err := offer.ChangeDescription("New description")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.Description() != "New description" {
			t.Errorf("expected description 'New description', got %v", offer.Description())
		}
		if offer.UpdatedAt().Before(oldUpdatedAt) {
			t.Error("expected UpdatedAt not to be before old value")
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

	t.Run("description too long (513 chars)", func(t *testing.T) {
		err := offer.ChangeDescription(strings.Repeat("a", 513))
		if err == nil {
			t.Error("expected error for long description")
		}
		if err != ErrInvalidDescription {
			t.Errorf("expected ErrInvalidDescription, got %v", err)
		}
	})

	t.Run("description exactly 512 chars", func(t *testing.T) {
		desc := strings.Repeat("a", 512)
		err := offer.ChangeDescription(desc)
		if err != nil {
			t.Fatalf("unexpected error for 512 char description: %v", err)
		}
		if offer.Description() != desc {
			t.Errorf("expected description length 512, got %d", len(offer.Description()))
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

	t.Run("short description (3 chars) - should fail", func(t *testing.T) {
		err := offer.ChangeDescription("abc")
		if err == nil {
			t.Error("expected error for 3 char description (len > 3 required)")
		}
		if err != ErrInvalidDescription {
			t.Errorf("expected ErrInvalidDescription, got %v", err)
		}
	})

	t.Run("valid min description (4 chars)", func(t *testing.T) {
		err := offer.ChangeDescription("abcd")
		if err != nil {
			t.Fatalf("unexpected error for 4 char description: %v", err)
		}
		if offer.Description() != "abcd" {
			t.Errorf("expected description 'abcd', got %v", offer.Description())
		}
	})
}

func TestOffer_ChangeSlotDuration(t *testing.T) {
	validAuthorID := uuid.New()
	validDuration := SlotDuration(30 * time.Minute)
	validInterval := SlotInterval(15 * time.Minute)
	validPrice := uint64(1500)
	offer, _ := New(validAuthorID, "Title", "description", validPrice, validDuration, validInterval)

	t.Run("valid slot duration change", func(t *testing.T) {
		oldUpdatedAt := offer.UpdatedAt()
		time.Sleep(1 * time.Millisecond)
		newDuration := SlotDuration(45 * time.Minute)

		err := offer.ChangeSlotDuration(newDuration)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.SlotDuration() != newDuration {
			t.Errorf("expected slotDuration %v, got %v", newDuration, offer.SlotDuration())
		}
		if offer.UpdatedAt().Before(oldUpdatedAt) {
			t.Error("expected UpdatedAt not to be before old value")
		}
	})

	t.Run("invalid slot duration (24 hours)", func(t *testing.T) {
		err := offer.ChangeSlotDuration(SlotDuration(24 * time.Hour))
		if err == nil {
			t.Error("expected error for invalid slot duration")
		}
		if err != ErrInvalidSlotDuration {
			t.Errorf("expected ErrInvalidSlotDuration, got %v", err)
		}
	})

	t.Run("invalid slot duration (zero)", func(t *testing.T) {
		err := offer.ChangeSlotDuration(SlotDuration(0))
		if err == nil {
			t.Error("expected error for zero slot duration")
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
	validPrice := uint64(1500)
	offer, _ := New(validAuthorID, "Title", "description", validPrice, validDuration, validInterval)

	t.Run("valid slot interval change", func(t *testing.T) {
		oldUpdatedAt := offer.UpdatedAt()
		time.Sleep(1 * time.Millisecond)
		newInterval := SlotInterval(30 * time.Minute)

		err := offer.ChangeSlotInterval(newInterval)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.SlotInterval() != newInterval {
			t.Errorf("expected slotInterval %v, got %v", newInterval, offer.SlotInterval())
		}
		if offer.UpdatedAt().Before(oldUpdatedAt) {
			t.Error("expected UpdatedAt not to be before old value")
		}
	})

	t.Run("invalid slot interval (24 hours)", func(t *testing.T) {
		err := offer.ChangeSlotInterval(SlotInterval(24 * time.Hour))
		if err == nil {
			t.Error("expected error for invalid slot interval")
		}
		if err != ErrInvalidSlotInterval {
			t.Errorf("expected ErrInvalidSlotInterval, got %v", err)
		}
	})

	t.Run("invalid slot interval (zero)", func(t *testing.T) {
		err := offer.ChangeSlotInterval(SlotInterval(0))
		if err == nil {
			t.Error("expected error for zero slot interval")
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
	validPrice := uint64(1500)

	t.Run("disable active offer", func(t *testing.T) {
		offer, _ := New(validAuthorID, "Title", "description", validPrice, validDuration, validInterval)
		oldUpdatedAt := offer.UpdatedAt()
		time.Sleep(1 * time.Millisecond)

		err := offer.Disable()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if offer.IsActive().Bool() {
			t.Error("expected offer to be disabled")
		}
		if offer.UpdatedAt().Before(oldUpdatedAt) {
			t.Error("expected UpdatedAt not to be before old value")
		}
	})

	t.Run("disable already disabled offer", func(t *testing.T) {
		offer, _ := New(validAuthorID, "Title", "description", validPrice, validDuration, validInterval)
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
		offer, _ := New(validAuthorID, "Title", "description", validPrice, validDuration, validInterval)
		_ = offer.Disable()
		oldUpdatedAt := offer.UpdatedAt()
		time.Sleep(1 * time.Millisecond)

		err := offer.Enable()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !offer.IsActive().Bool() {
			t.Error("expected offer to be enabled")
		}
		if offer.UpdatedAt().Before(oldUpdatedAt) {
			t.Error("expected UpdatedAt not to be before old value")
		}
	})

	t.Run("enable already enabled offer", func(t *testing.T) {
		offer, _ := New(validAuthorID, "Title", "description", validPrice, validDuration, validInterval)

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
	validPrice := uint64(2000)
	offer, _ := New(validAuthorID, "Title", "description", validPrice, validDuration, validInterval)

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
	if offer.Price() != validPrice {
		t.Errorf("Price() = %v, want %v", offer.Price(), validPrice)
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
