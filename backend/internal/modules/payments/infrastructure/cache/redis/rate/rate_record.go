package rateredis

import "github.com/google/uuid"

type rateRecord struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	AmountPenny uint64    `json:"amount_penny"`
	PaidDays    uint32    `json:"paid_days"`
}
