package requests

import "github.com/google/uuid"

type UpsertClient struct {
	OfferID uuid.UUID `json:"offer_id"`
	Visit   Time      `json:"visit"`
}
