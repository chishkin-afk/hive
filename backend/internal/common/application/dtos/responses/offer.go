package responses

type Offer struct {
	ID                  string `json:"id"`
	AuthorID            string `json:"author_id"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	Price               uint64 `json:"price"`
	SlotDurationMinutes uint32 `json:"slot_duration_minutes"`
	SlotIntervalMinutes uint32 `json:"slot_interval_minutes"`
	IsActive            bool   `json:"is_active"`
	CreatedAt           int64  `json:"created_at"`
	UpdatedAt           int64  `json:"updated_at"`
}

type Offers struct {
	List        []*Offer `json:"list"`
	TotalPages  uint32   `json:"total_pages"`
	HasNextPage bool     `json:"has_next_page"`
	HasPrevPage bool     `json:"has_prev_page"`
}
