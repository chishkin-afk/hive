package booking

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewBooking_Success(t *testing.T) {
	offerID := uuid.New()
	userID := uuid.New()
	comment := "Test comment"
	startTime := time.Now().UTC().Add(24 * time.Hour)

	booking, err := New(offerID, userID, comment, startTime)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if booking == nil {
		t.Fatal("expected booking to be not nil")
	}
	if booking.OfferID() != offerID {
		t.Errorf("expected offerID %v, got %v", offerID, booking.OfferID())
	}
	if booking.UserID() != userID {
		t.Errorf("expected userID %v, got %v", userID, booking.UserID())
	}
	if booking.Comment() != comment {
		t.Errorf("expected comment %q, got %q", comment, booking.Comment())
	}
	if !booking.StartTime().Equal(startTime) {
		t.Errorf("expected startTime %v, got %v", startTime, booking.StartTime())
	}
	if booking.IsCancelled() {
		t.Error("expected new booking to be not cancelled")
	}
	if booking.ID() == uuid.Nil {
		t.Error("expected new booking to have a non-nil ID")
	}
	if booking.CreatedAt().IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if booking.UpdatedAt().IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestNewBooking_EmptyOfferID(t *testing.T) {
	userID := uuid.New()
	startTime := time.Now().UTC()

	_, err := New(uuid.Nil, userID, "", startTime)

	if err != ErrEmptyOfferID {
		t.Errorf("expected ErrEmptyOfferID, got %v", err)
	}
}

func TestNewBooking_EmptyUserID(t *testing.T) {
	offerID := uuid.New()
	startTime := time.Now().UTC()

	_, err := New(offerID, uuid.Nil, "", startTime)

	if err != ErrEmptyUserID {
		t.Errorf("expected ErrEmptyUserID, got %v", err)
	}
}

func TestCancel_Success(t *testing.T) {
	booking := createValidBooking(t)

	err := booking.Cancel()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !booking.IsCancelled() {
		t.Error("expected booking to be cancelled")
	}
}

func TestCancel_AlreadyCancelled(t *testing.T) {
	booking := createValidBooking(t)
	_ = booking.Cancel()

	err := booking.Cancel()

	if err != ErrCancelled {
		t.Errorf("expected ErrCancelled, got %v", err)
	}
}

func TestChangeComment_Success(t *testing.T) {
	booking := createValidBooking(t)
	newComment := "Updated comment"

	err := booking.ChangeComment(newComment)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if booking.Comment() != newComment {
		t.Errorf("expected comment %q, got %q", newComment, booking.Comment())
	}
}

func TestChangeComment_TrimSpace(t *testing.T) {
	booking := createValidBooking(t)
	newComment := "  spaced comment  "

	err := booking.ChangeComment(newComment)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if booking.Comment() != "spaced comment" {
		t.Errorf("expected trimmed comment, got %q", booking.Comment())
	}
}

func TestChangeComment_CancelledBooking(t *testing.T) {
	booking := createValidBooking(t)
	_ = booking.Cancel()

	err := booking.ChangeComment("new comment")

	if err != ErrCancelled {
		t.Errorf("expected ErrCancelled, got %v", err)
	}
}

func TestChangeStartTime_Success(t *testing.T) {
	booking := createValidBooking(t)
	newStartTime := time.Now().UTC().Add(48 * time.Hour)

	err := booking.ChangeStartTime(newStartTime)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !booking.StartTime().Equal(newStartTime) {
		t.Errorf("expected startTime %v, got %v", newStartTime, booking.StartTime())
	}
}

func TestChangeStartTime_CancelledBooking(t *testing.T) {
	booking := createValidBooking(t)
	_ = booking.Cancel()
	newStartTime := time.Now().UTC().Add(48 * time.Hour)

	err := booking.ChangeStartTime(newStartTime)

	if err != ErrCancelled {
		t.Errorf("expected ErrCancelled, got %v", err)
	}
}

func createValidBooking(t *testing.T) *Booking {
	offerID := uuid.New()
	userID := uuid.New()
	startTime := time.Now().UTC().Add(24 * time.Hour)

	booking, err := New(offerID, userID, "test", startTime)
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}
	return booking
}
