package offerredis

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/devathh/hive/backend/internal/modules/offers/domain/offer"
	"github.com/google/uuid"
)

func TestDomainToBytes(t *testing.T) {
	authorID := uuid.New()

	tests := []struct {
		name          string
		domainFactory func() *offer.Offer
		validateJSON  func(t *testing.T, data []byte)
	}{
		{
			name: "active offer with price",
			domainFactory: func() *offer.Offer {
				o, _ := offer.New(
					authorID,
					"Cache Offer",
					"Stored in Redis",
					uint64(1500),
					offer.SlotDuration(45*time.Minute),
					offer.SlotInterval(15*time.Minute),
				)
				return o
			},
			validateJSON: func(t *testing.T, data []byte) {
				var rec offerRecord
				if err := json.Unmarshal(data, &rec); err != nil {
					t.Fatalf("failed to unmarshal for validation: %v", err)
				}
				if rec.Title != "Cache Offer" {
					t.Errorf("title = %v, want 'Cache Offer'", rec.Title)
				}
				if rec.Description != "Stored in Redis" {
					t.Errorf("description = %v, want 'Stored in Redis'", rec.Description)
				}
				if rec.Price != 1500 {
					t.Errorf("price = %v, want 1500", rec.Price)
				}
				if rec.SlotDurationMinutes != 45 {
					t.Errorf("slot_duration_minutes = %v, want 45", rec.SlotDurationMinutes)
				}
				if rec.SlotIntervalMinutes != 15 {
					t.Errorf("slot_interval_minutes = %v, want 15", rec.SlotIntervalMinutes)
				}
				if !rec.IsActive {
					t.Error("is_active = false, want true")
				}
			},
		},
		{
			name: "disabled offer with zero price",
			domainFactory: func() *offer.Offer {
				o, _ := offer.New(
					authorID,
					"Free Offer",
					"",
					uint64(0),
					offer.SlotDuration(10*time.Minute),
					offer.SlotInterval(5*time.Minute),
				)
				_ = o.Disable()
				return o
			},
			validateJSON: func(t *testing.T, data []byte) {
				var rec offerRecord
				_ = json.Unmarshal(data, &rec)
				if rec.Description != "" {
					t.Errorf("description = %v, want empty", rec.Description)
				}
				if rec.Price != 0 {
					t.Errorf("price = %v, want 0", rec.Price)
				}
				if rec.IsActive {
					t.Error("is_active = true, want false")
				}
				if rec.SlotDurationMinutes != 10 {
					t.Errorf("slot_duration_minutes = %v, want 10", rec.SlotDurationMinutes)
				}
			},
		},
		{
			name: "offer with large price value",
			domainFactory: func() *offer.Offer {
				o, _ := offer.New(
					authorID,
					"Premium",
					"High value offer",
					uint64(999999999),
					offer.SlotDuration(60*time.Minute),
					offer.SlotInterval(30*time.Minute),
				)
				return o
			},
			validateJSON: func(t *testing.T, data []byte) {
				var rec offerRecord
				_ = json.Unmarshal(data, &rec)
				if rec.Price != 999999999 {
					t.Errorf("price = %v, want 999999999", rec.Price)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain := tt.domainFactory()
			data, err := domainToBytes(domain)
			if err != nil {
				t.Fatalf("domainToBytes() error = %v", err)
			}
			if data == nil {
				t.Fatal("domainToBytes() returned nil bytes")
			}
			tt.validateJSON(t, data)
		})
	}
}

func TestBytesToDomain(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	authorID := uuid.New()
	offerID := uuid.New()

	tests := []struct {
		name         string
		record       offerRecord
		validateFunc func(*testing.T, *offer.Offer)
		wantErr      bool
	}{
		{
			name: "valid active offer with price",
			record: offerRecord{
				ID:                  offerID,
				AuthorID:            authorID,
				Title:               "Redis Title",
				Description:         "Redis desc",
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
				if o.Title() != "Redis Title" {
					t.Errorf("Title() = %v, want 'Redis Title'", o.Title())
				}
				if o.Price() != 2500 {
					t.Errorf("Price() = %v, want 2500", o.Price())
				}
				if o.SlotDuration().Minutes() != 30 {
					t.Errorf("SlotDuration().Minutes() = %v, want 30", o.SlotDuration().Minutes())
				}
				if !o.IsActive().Bool() {
					t.Error("IsActive().Bool() = false, want true")
				}
			},
			wantErr: false,
		},
		{
			name: "valid disabled offer with zero price",
			record: offerRecord{
				ID:                  uuid.New(),
				AuthorID:            uuid.New(),
				Title:               "Disabled",
				Description:         "",
				Price:               0,
				SlotDurationMinutes: 60,
				SlotIntervalMinutes: 30,
				IsActive:            false,
				CreatedAt:           now.Add(-time.Hour),
				UpdatedAt:           now,
			},
			validateFunc: func(t *testing.T, o *offer.Offer) {
				if o.IsActive().Bool() {
					t.Error("IsActive().Bool() = true, want false")
				}
				if o.Price() != 0 {
					t.Errorf("Price() = %v, want 0", o.Price())
				}
				if o.Description() != "" {
					t.Errorf("Description() = %v, want empty", o.Description())
				}
			},
			wantErr: false,
		},
		{
			name: "invalid JSON bytes",
			record: offerRecord{
				Title: "test",
			},
			validateFunc: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			var err error
			if tt.wantErr {
				data = []byte("invalid json {{{")
			} else {
				data, err = json.Marshal(tt.record)
				if err != nil {
					t.Fatalf("failed to prepare test data: %v", err)
				}
			}

			domain, err := bytesToDomain(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("bytesToDomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validateFunc != nil {
				tt.validateFunc(t, domain)
			}
		})
	}
}

func TestMapperRoundTrip(t *testing.T) {
	authorID := uuid.New()

	t.Run("domain → bytes → domain preserves all values", func(t *testing.T) {
		originalDomain, err := offer.New(
			authorID,
			"RoundTrip Cache",
			"Testing JSON mapper",
			uint64(3500),
			offer.SlotDuration(25*time.Minute),
			offer.SlotInterval(5*time.Minute),
		)
		if err != nil {
			t.Fatalf("failed to create original domain: %v", err)
		}

		bytes, err := domainToBytes(originalDomain)
		if err != nil {
			t.Fatalf("domainToBytes() error = %v", err)
		}

		restoredDomain, err := bytesToDomain(bytes)
		if err != nil {
			t.Fatalf("bytesToDomain() error = %v", err)
		}

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

	t.Run("disabled state and price preserved in JSON round-trip", func(t *testing.T) {
		domain, _ := offer.New(
			authorID,
			"Disabled Cache Test",
			"desc",
			uint64(7500),
			offer.SlotDuration(15*time.Minute),
			offer.SlotInterval(15*time.Minute),
		)
		_ = domain.Disable()

		bytes, _ := domainToBytes(domain)
		restored, _ := bytesToDomain(bytes)

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
		name             string
		domainDuration   offer.SlotDuration
		domainInterval   offer.SlotInterval
		price            uint64
		expectedDuration int64
		expectedInterval int64
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

			bytes, err := domainToBytes(domain)
			if err != nil {
				t.Fatalf("domainToBytes() error = %v", err)
			}

			var rec offerRecord
			if err := json.Unmarshal(bytes, &rec); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if rec.Price != tt.price {
				t.Errorf("Price = %v, want %v", rec.Price, tt.price)
			}
			if rec.SlotDurationMinutes != tt.expectedDuration {
				t.Errorf("SlotDurationMinutes = %v, want %v", rec.SlotDurationMinutes, tt.expectedDuration)
			}
			if rec.SlotIntervalMinutes != tt.expectedInterval {
				t.Errorf("SlotIntervalMinutes = %v, want %v", rec.SlotIntervalMinutes, tt.expectedInterval)
			}

			restored, err := bytesToDomain(bytes)
			if err != nil {
				t.Fatalf("bytesToDomain() error = %v", err)
			}
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

func TestJSONFieldTags(t *testing.T) {
	authorID := uuid.New()
	domain, _ := offer.New(
		authorID,
		"Tag Test",
		"desc",
		uint64(2000),
		offer.SlotDuration(20*time.Minute),
		offer.SlotInterval(10*time.Minute),
	)

	bytes, err := domainToBytes(domain)
	if err != nil {
		t.Fatalf("domainToBytes() error = %v", err)
	}

	jsonStr := string(bytes)
	expectedFields := []string{
		`"id"`,
		`"author_id"`,
		`"title"`,
		`"description"`,
		`"price"`,
		`"slot_duration_minutes"`,
		`"slot_interval_minutes"`,
		`"is_active"`,
		`"created_at"`,
		`"updated_at"`,
	}

	for _, field := range expectedFields {
		if !contains(jsonStr, field) {
			t.Errorf("JSON missing expected field tag: %s", field)
		}
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
