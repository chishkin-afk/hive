package ratepg

import "github.com/google/uuid"

type rateRecord struct {
	ID          uuid.UUID
	Name        string
	AmountPenny uint64
	PaidDays    uint32
}
