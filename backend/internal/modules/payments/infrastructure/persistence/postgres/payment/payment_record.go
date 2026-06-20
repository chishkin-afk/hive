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
	IsConfirmed bool
	CreatedAt   time.Time
	ConfirmedAt *time.Time
}
