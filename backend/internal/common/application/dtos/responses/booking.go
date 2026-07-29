package responses

import "github.com/google/uuid"

type Booking struct {
	ID          uuid.UUID `json:"id"`
	OfferID     uuid.UUID `json:"offer_id"`
	UserID      uuid.UUID `json:"user_id"`
	Comment     string    `json:"comment"`
	StartTime   int64     `json:"start_time"`
	IsCancelled bool      `json:"is_cancelled"`
	CreatedAt   int64     `json:"created_at"`
	UpdatedAt   int64     `json:"updated_at"`
}

type Bookings struct {
	List        []ListElement `json:"list"`
	TotalPages  uint32        `json:"total_pages"`
	HasNextPage bool          `json:"has_next_page"`
	HasPrevPage bool          `json:"has_prev_page"`
}

type ListElement struct {
	Booking *Booking     `json:"booking"`
	User    *BookingUser `json:"user"`
}

type BookingUser struct {
	ID             uuid.UUID `json:"id"`
	Firstname      string    `json:"firstname"`
	Lastname       string    `json:"lastname"`
	ConfirmedEmail string    `json:"confirmed_email"`
}

type Slots struct {
	Date  string    `json:"date"`
	Times []DayTime `json:"times"`
}

type Days struct {
	List []Day `json:"list"`
}

type Day struct {
	Time     int64 `json:"time"`
	IsActive bool  `json:"is_active"` // true if it has free slots & < paid until
}
