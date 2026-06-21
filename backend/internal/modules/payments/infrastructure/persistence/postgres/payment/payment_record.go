package paymentpg

import (
	"time"

	"github.com/google/uuid"
)

type paymentRecord struct {
	ID          uuid.UUID
	PaymentID   uuid.UUID
	UserID      uuid.UUID
	AmountPenny uint64
	StatusInt   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
