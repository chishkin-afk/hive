package requests

import "github.com/google/uuid"

type CreateOffer struct {
	Title               string `json:"title"`
	Description         string `json:"description"`
	Price               uint64 `json:"price_penny"`
	SlotDurationMinutes uint32 `json:"slot_duration_minutes"`
	SlotIntervalMinutes uint32 `json:"slot_interval_minutes"`
}

type UpdateOffer struct {
	OfferID             uuid.UUID `json:"offer_id"`
	Title               *string   `json:"title"`
	Description         *string   `json:"description"`
	Price               *uint32   `json:"price"`
	SlotDurationMinutes *uint32   `json:"slot_duration_minutes"`
	SlotIntervalMinutes *uint32   `json:"slot_interval_minutes"`
	IsActive            *bool     `json:"is_active"`
}
