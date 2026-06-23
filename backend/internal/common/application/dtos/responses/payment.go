package responses

import "github.com/google/uuid"

type CreatedPayment struct {
	Payment *Payment `json:"payment"`
	URL     string   `json:"url"`
}

type Payment struct {
	ID           uuid.UUID `json:"id"`
	PaymentID    uuid.UUID `json:"payment_id"`
	UserID       uuid.UUID `json:"user_id"`
	AmountPenny  uint64    `json:"amount_penny"`
	StatusString string    `json:"status_string"`
	StatusInt    int       `json:"status_int"`
	CreatedAt    int64     `json:"created_at"`
	UpdatedAt    int64     `json:"updated_at"`
}

type Payments struct {
	List        []*Payment `json:"list"`
	HasNextPage bool       `json:"has_next_page"`
	HasPrevPage bool       `json:"has_prev_page"`
	TotalPages  uint32     `json:"total_pages"`
}

type Rate struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	AmountPenny uint64    `json:"amount_penny"`
	PaidDays    uint32    `json:"paid_days"`
}

type Rates struct {
	List []*Rate `json:"rate"`
}
