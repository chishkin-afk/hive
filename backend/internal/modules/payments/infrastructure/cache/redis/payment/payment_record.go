package paymentredis

import (
	"time"

	"github.com/google/uuid"
)

type paymentRecord struct {
	ID          uuid.UUID `json:"id"`
	PaymentID   uuid.UUID `json:"payment_id"`
	UserID      uuid.UUID `json:"user_id"`
	AmountPenny uint64    `json:"amount_penny"`
	StatusInt   int       `json:"status_int"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
