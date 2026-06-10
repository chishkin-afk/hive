package responses

type Client struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	OfferID   string `json:"offer_id"`
	Visits    uint64 `json:"visits"`
	LastVisit int64  `json:"last_visit"`
	CreatedAt int64  `json:"created_at"`
}

type Clients struct {
	List        []*Client `json:"list"`
	TotalPages  uint32    `json:"total_pages"`
	HasNextPage bool      `json:"has_next_page"`
	HasPrevPage bool      `json:"has_prev_page"`
}
