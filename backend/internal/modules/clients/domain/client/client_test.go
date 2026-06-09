package client

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewClient(t *testing.T) {
	validUserID := uuid.New()
	validOfferID := uuid.New()

	t.Run("valid client creation", func(t *testing.T) {
		client, err := New(validUserID, validOfferID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client.ID() == uuid.Nil {
			t.Error("expected non-nil UUID")
		}
		if client.UserID() != validUserID {
			t.Errorf("expected userID %v, got %v", validUserID, client.UserID())
		}
		if client.OfferID() != validOfferID {
			t.Errorf("expected offerID %v, got %v", validOfferID, client.OfferID())
		}
		if client.Visits() != 1 {
			t.Errorf("expected visits 1, got %d", client.Visits())
		}
		if client.CreatedAt().IsZero() {
			t.Error("expected non-zero CreatedAt")
		}
	})

	t.Run("invalid user ID", func(t *testing.T) {
		_, err := New(uuid.Nil, validOfferID)
		if err == nil {
			t.Error("expected error for nil user ID")
		}
		if err != ErrInvalidUserID {
			t.Errorf("expected ErrInvalidUserID, got %v", err)
		}
	})

	t.Run("invalid offer ID", func(t *testing.T) {
		_, err := New(validUserID, uuid.Nil)
		if err == nil {
			t.Error("expected error for nil offer ID")
		}
		if err != ErrInvalidOfferID {
			t.Errorf("expected ErrInvalidOfferID, got %v", err)
		}
	})

	t.Run("both IDs invalid", func(t *testing.T) {
		_, err := New(uuid.Nil, uuid.Nil)
		if err == nil {
			t.Error("expected error for both nil IDs")
		}
		if err != ErrInvalidUserID {
			t.Errorf("expected ErrInvalidUserID, got %v", err)
		}
	})
}

func TestClient_AddVisit(t *testing.T) {
	validUserID := uuid.New()
	validOfferID := uuid.New()
	client, _ := New(validUserID, validOfferID)

	t.Run("first visit increments counter", func(t *testing.T) {
		initialVisits := client.Visits()
		time.Sleep(1 * time.Millisecond)

		client.AddVisit()

		if client.Visits() != initialVisits+1 {
			t.Errorf("expected visits %d, got %d", initialVisits+1, client.Visits())
		}
	})

	t.Run("multiple visits increment correctly", func(t *testing.T) {
		initialVisits := client.Visits()

		for i := 0; i < 5; i++ {
			client.AddVisit()
		}

		expectedVisits := initialVisits + 5
		if client.Visits() != expectedVisits {
			t.Errorf("expected visits %d, got %d", expectedVisits, client.Visits())
		}
	})

	t.Run("lastVisit is updated", func(t *testing.T) {
		oldLastVisit := client.LastVisit()
		time.Sleep(1 * time.Millisecond)

		client.AddVisit()

		if !client.LastVisit().After(oldLastVisit) {
			t.Error("expected LastVisit to be updated")
		}
	})
}

func TestClient_Getters(t *testing.T) {
	validUserID := uuid.New()
	validOfferID := uuid.New()
	client, _ := New(validUserID, validOfferID)

	t.Run("ID returns non-nil UUID", func(t *testing.T) {
		if client.ID() == uuid.Nil {
			t.Error("ID() returned nil UUID")
		}
	})

	t.Run("UserID returns correct value", func(t *testing.T) {
		if client.UserID() != validUserID {
			t.Errorf("UserID() = %v, want %v", client.UserID(), validUserID)
		}
	})

	t.Run("OfferID returns correct value", func(t *testing.T) {
		if client.OfferID() != validOfferID {
			t.Errorf("OfferID() = %v, want %v", client.OfferID(), validOfferID)
		}
	})

	t.Run("Visits returns initial value 1", func(t *testing.T) {
		if client.Visits() != 1 {
			t.Errorf("Visits() = %d, want 1", client.Visits())
		}
	})

	t.Run("CreatedAt returns non-zero time", func(t *testing.T) {
		if client.CreatedAt().IsZero() {
			t.Error("CreatedAt() returned zero time")
		}
	})

	t.Run("LastVisit returns non-zero time", func(t *testing.T) {
		if client.LastVisit().IsZero() {
			t.Error("LastVisit() returned zero time")
		}
	})

	t.Run("CreatedAt and LastVisit are equal initially", func(t *testing.T) {
		if !client.CreatedAt().Equal(client.LastVisit()) {
			t.Errorf("CreatedAt() = %v, LastVisit() = %v, expected equal", client.CreatedAt(), client.LastVisit())
		}
	})
}

func TestClient_Immutability(t *testing.T) {
	validUserID := uuid.New()
	validOfferID := uuid.New()
	client, _ := New(validUserID, validOfferID)

	originalID := client.ID()
	originalUserID := client.UserID()
	originalOfferID := client.OfferID()
	originalCreatedAt := client.CreatedAt()

	t.Run("userID cannot be changed", func(t *testing.T) {
		client.AddVisit()
		if client.UserID() != originalUserID {
			t.Errorf("userID changed from %v to %v", originalUserID, client.UserID())
		}
	})

	t.Run("offerID cannot be changed", func(t *testing.T) {
		client.AddVisit()
		if client.OfferID() != originalOfferID {
			t.Errorf("offerID changed from %v to %v", originalOfferID, client.OfferID())
		}
	})

	t.Run("id cannot be changed", func(t *testing.T) {
		client.AddVisit()
		if client.ID() != originalID {
			t.Errorf("id changed from %v to %v", originalID, client.ID())
		}
	})

	t.Run("createdAt cannot be changed", func(t *testing.T) {
		client.AddVisit()
		if !client.CreatedAt().Equal(originalCreatedAt) {
			t.Errorf("createdAt changed from %v to %v", originalCreatedAt, client.CreatedAt())
		}
	})
}
