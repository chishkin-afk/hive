package responses

import "github.com/google/uuid"

type Client struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	OfferID   uuid.UUID `json:"offer_id"`
	Visits    uint64    `json:"visits"`
	LastVisit int64     `json:"last_visit"`
	CreatedAt int64     `json:"created_at"`
}

type Clients struct {
	List        []*Client `json:"list"`
	TotalPages  uint32    `json:"total_pages"`
	HasNextPage bool      `json:"has_next_page"`
	HasPrevPage bool      `json:"has_prev_page"`
}
