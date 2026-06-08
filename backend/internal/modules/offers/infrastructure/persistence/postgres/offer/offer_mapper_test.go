package offerpg

import (
	"testing"
	"time"

	"github.com/devathh/hive/backend/internal/modules/offers/domain/offer"
	"github.com/google/uuid"
)

func TestDomainToRecord(t *testing.T) {
	authorID := uuid.New()

	tests := []struct {
		name           string
		domainFactory  func() *offer.Offer
		expectedRecord *offerRecord
	}{
		{
			name: "full offer mapping",
			domainFactory: func() *offer.Offer {
				o, _ := offer.New(
					authorID,
					"Test Offer",
					"Test description",
					uint64(1500),
					offer.SlotDuration(45*time.Minute),
					offer.SlotInterval(15*time.Minute),
				)
				return o
			},
			expectedRecord: &offerRecord{
				AuthorID:            authorID,
				Title:               "Test Offer",
				Description:         "Test description",
				Price:               1500,
				SlotDurationMinutes: 45,
				SlotIntervalMinutes: 15,
				IsActive:            true,
			},
		},
		{
			name: "offer with zero price",
			domainFactory: func() *offer.Offer {
				o, _ := offer.New(
					authorID,
					"Free Offer",
					"",
					uint64(0),
					offer.SlotDuration(10*time.Minute),
					offer.SlotInterval(5*time.Minute),
				)
				return o
			},
			expectedRecord: &offerRecord{
				AuthorID:            authorID,
				Title:               "Free Offer",
				Description:         "",
				Price:               0,
				SlotDurationMinutes: 10,
				SlotIntervalMinutes: 5,
				IsActive:            true,
			},
		},
		{
			name: "disabled offer with high price",
			domainFactory: func() *offer.Offer {
				o, _ := offer.New(
					authorID,
					"Premium Offer",
					"description",
					uint64(999999),
					offer.SlotDuration(60*time.Minute),
					offer.SlotInterval(30*time.Minute),
				)
				_ = o.Disable()
				return o
			},
			expectedRecord: &offerRecord{
				AuthorID:            authorID,
				Title:               "Premium Offer",
				Description:         "description",
				Price:               999999,
				SlotDurationMinutes: 60,
				SlotIntervalMinutes: 30,
				IsActive:            false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain := tt.domainFactory()
			record := domainToRecord(domain)

			if record.AuthorID != tt.expectedRecord.AuthorID {
				t.Errorf("AuthorID = %v, want %v", record.AuthorID, tt.expectedRecord.AuthorID)
			}
			if record.Title != tt.expectedRecord.Title {
				t.Errorf("Title = %v, want %v", record.Title, tt.expectedRecord.Title)
			}
			if record.Description != tt.expectedRecord.Description {
				t.Errorf("Description = %v, want %v", record.Description, tt.expectedRecord.Description)
			}
			if record.Price != tt.expectedRecord.Price {
				t.Errorf("Price = %v, want %v", record.Price, tt.expectedRecord.Price)
			}
			if record.SlotDurationMinutes != tt.expectedRecord.SlotDurationMinutes {
				t.Errorf("SlotDurationMinutes = %v, want %v", record.SlotDurationMinutes, tt.expectedRecord.SlotDurationMinutes)
			}
			if record.SlotIntervalMinutes != tt.expectedRecord.SlotIntervalMinutes {
				t.Errorf("SlotIntervalMinutes = %v, want %v", record.SlotIntervalMinutes, tt.expectedRecord.SlotIntervalMinutes)
			}
			if record.IsActive != tt.expectedRecord.IsActive {
				t.Errorf("IsActive = %v, want %v", record.IsActive, tt.expectedRecord.IsActive)
			}
		})
	}
}

func TestRecordToDomain(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	authorID := uuid.New()
	offerID := uuid.New()

	tests := []struct {
		name         string
		record       *offerRecord
		validateFunc func(*testing.T, *offer.Offer)
	}{
		{
			name: "active offer with price",
			record: &offerRecord{
				ID:                  offerID,
				AuthorID:            authorID,
				Title:               "Record Title",
				Description:         "Record description",
				Price:               2500,
				SlotDurationMinutes: 30,
				SlotIntervalMinutes: 10,
				IsActive:            true,
				CreatedAt:           now,
				UpdatedAt:           now,
			},
			validateFunc: func(t *testing.T, o *offer.Offer) {
				if o.ID() != offerID {
					t.Errorf("ID() = %v, want %v", o.ID(), offerID)
				}
				if o.AuthorID() != authorID {
					t.Errorf("AuthorID() = %v, want %v", o.AuthorID(), authorID)
				}
				if o.Title() != "Record Title" {
					t.Errorf("Title() = %v, want 'Record Title'", o.Title())
				}
				if o.Description() != "Record description" {
					t.Errorf("Description() = %v, want 'Record description'", o.Description())
				}
				if o.Price() != 2500 {
					t.Errorf("Price() = %v, want 2500", o.Price())
				}
				if o.SlotDuration().Minutes() != 30 {
					t.Errorf("SlotDuration().Minutes() = %v, want 30", o.SlotDuration().Minutes())
				}
				if o.SlotInterval().Minutes() != 10 {
					t.Errorf("SlotInterval().Minutes() = %v, want 10", o.SlotInterval().Minutes())
				}
				if !o.IsActive().Bool() {
					t.Error("IsActive().Bool() = false, want true")
				}
				if !o.CreatedAt().Equal(now) {
					t.Errorf("CreatedAt() = %v, want %v", o.CreatedAt(), now)
				}
				if !o.UpdatedAt().Equal(now) {
					t.Errorf("UpdatedAt() = %v, want %v", o.UpdatedAt(), now)
				}
			},
		},
		{
			name: "disabled offer with zero price",
			record: &offerRecord{
				ID:                  uuid.New(),
				AuthorID:            uuid.New(),
				Title:               "Minimal",
				Description:         "",
				Price:               0,
				SlotDurationMinutes: 5,
				SlotIntervalMinutes: 1,
				IsActive:            false,
				CreatedAt:           now.Add(-time.Hour),
				UpdatedAt:           now,
			},
			validateFunc: func(t *testing.T, o *offer.Offer) {
				if o.Description() != "" {
					t.Errorf("Description() = %v, want empty", o.Description())
				}
				if o.Price() != 0 {
					t.Errorf("Price() = %v, want 0", o.Price())
				}
				if o.IsActive().Bool() {
					t.Error("IsActive().Bool() = true, want false")
				}
				if o.SlotDuration().Minutes() != 5 {
					t.Errorf("SlotDuration().Minutes() = %v, want 5", o.SlotDuration().Minutes())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain := recordToDomain(tt.record)
			tt.validateFunc(t, domain)
		})
	}
}

func TestMapperRoundTrip(t *testing.T) {
	authorID := uuid.New()

	t.Run("domain → record → domain preserves all values", func(t *testing.T) {
		originalDomain, err := offer.New(
			authorID,
			"RoundTrip Offer",
			"Testing mapper consistency",
			uint64(3500),
			offer.SlotDuration(25*time.Minute),
			offer.SlotInterval(5*time.Minute),
		)
		if err != nil {
			t.Fatalf("failed to create original domain: %v", err)
		}

		record := domainToRecord(originalDomain)
		restoredDomain := recordToDomain(record)

		if restoredDomain.AuthorID() != originalDomain.AuthorID() {
			t.Errorf("AuthorID mismatch: got %v, want %v", restoredDomain.AuthorID(), originalDomain.AuthorID())
		}
		if restoredDomain.Title() != originalDomain.Title() {
			t.Errorf("Title mismatch: got %v, want %v", restoredDomain.Title(), originalDomain.Title())
		}
		if restoredDomain.Description() != originalDomain.Description() {
			t.Errorf("Description mismatch: got %v, want %v", restoredDomain.Description(), originalDomain.Description())
		}
		if restoredDomain.Price() != originalDomain.Price() {
			t.Errorf("Price mismatch: got %v, want %v", restoredDomain.Price(), originalDomain.Price())
		}
		if restoredDomain.SlotDuration().Minutes() != originalDomain.SlotDuration().Minutes() {
			t.Errorf("SlotDuration mismatch: got %v, want %v", restoredDomain.SlotDuration().Minutes(), originalDomain.SlotDuration().Minutes())
		}
		if restoredDomain.SlotInterval().Minutes() != originalDomain.SlotInterval().Minutes() {
			t.Errorf("SlotInterval mismatch: got %v, want %v", restoredDomain.SlotInterval().Minutes(), originalDomain.SlotInterval().Minutes())
		}
		if restoredDomain.IsActive().Bool() != originalDomain.IsActive().Bool() {
			t.Errorf("IsActive mismatch: got %v, want %v", restoredDomain.IsActive().Bool(), originalDomain.IsActive().Bool())
		}
		if restoredDomain.ID() != originalDomain.ID() {
			t.Errorf("ID mismatch after round-trip: got %v, want %v", restoredDomain.ID(), originalDomain.ID())
		}
		if !restoredDomain.CreatedAt().Equal(originalDomain.CreatedAt()) {
			t.Errorf("CreatedAt mismatch: got %v, want %v", restoredDomain.CreatedAt(), originalDomain.CreatedAt())
		}
		if !restoredDomain.UpdatedAt().Equal(originalDomain.UpdatedAt()) {
			t.Errorf("UpdatedAt mismatch: got %v, want %v", restoredDomain.UpdatedAt(), originalDomain.UpdatedAt())
		}
	})

	t.Run("disabled state and price preserved in round-trip", func(t *testing.T) {
		domain, _ := offer.New(
			authorID,
			"Disabled Test",
			"desc",
			uint64(7500),
			offer.SlotDuration(15*time.Minute),
			offer.SlotInterval(15*time.Minute),
		)
		_ = domain.Disable()

		record := domainToRecord(domain)
		restored := recordToDomain(record)

		if restored.IsActive().Bool() {
			t.Error("IsActive state not preserved: expected disabled, got enabled")
		}
		if restored.Price() != 7500 {
			t.Errorf("Price not preserved: got %v, want 7500", restored.Price())
		}
	})
}

func TestSlotConversionAccuracy(t *testing.T) {
	tests := []struct {
		name              string
		domainDuration    offer.SlotDuration
		domainInterval    offer.SlotInterval
		price             uint64
		expectedDuration  int64
		expectedInterval  int64
	}{
		{"exact minutes", offer.SlotDuration(30 * time.Minute), offer.SlotInterval(10 * time.Minute), 1500, 30, 10},
		{"hours to minutes", offer.SlotDuration(2 * time.Hour), offer.SlotInterval(30 * time.Minute), 9999, 120, 30},
		{"mixed units", offer.SlotDuration(1*time.Hour + 30*time.Minute), offer.SlotInterval(45 * time.Minute), 500, 90, 45},
		{"seconds truncated", offer.SlotDuration(30*time.Minute + 30*time.Second), offer.SlotInterval(10*time.Minute + 59*time.Second), 2500, 30, 10},
		{"minimal valid values", offer.SlotDuration(1 * time.Minute), offer.SlotInterval(1 * time.Minute), 100, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authorID := uuid.New()
			domain, err := offer.New(
				authorID,
				"Conversion Test",
				"",
				tt.price,
				tt.domainDuration,
				tt.domainInterval,
			)
			if err != nil {
				t.Fatalf("failed to create domain: %v", err)
			}

			record := domainToRecord(domain)

			if record.Price != tt.price {
				t.Errorf("Price = %v, want %v", record.Price, tt.price)
			}
			if record.SlotDurationMinutes != tt.expectedDuration {
				t.Errorf("SlotDurationMinutes = %v, want %v", record.SlotDurationMinutes, tt.expectedDuration)
			}
			if record.SlotIntervalMinutes != tt.expectedInterval {
				t.Errorf("SlotIntervalMinutes = %v, want %v", record.SlotIntervalMinutes, tt.expectedInterval)
			}

			restored := recordToDomain(record)
			if restored.Price() != tt.price {
				t.Errorf("Restored Price() = %v, want %v", restored.Price(), tt.price)
			}
			if restored.SlotDuration().Minutes() != tt.expectedDuration {
				t.Errorf("Restored SlotDuration().Minutes() = %v, want %v", restored.SlotDuration().Minutes(), tt.expectedDuration)
			}
			if restored.SlotInterval().Minutes() != tt.expectedInterval {
				t.Errorf("Restored SlotInterval().Minutes() = %v, want %v", restored.SlotInterval().Minutes(), tt.expectedInterval)
			}
		})
	}
}
