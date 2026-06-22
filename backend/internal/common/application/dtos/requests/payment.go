package requests

import "github.com/google/uuid"

type CreatePayment struct {
	RateID uuid.UUID `json:"rate_id"`
}

type ConfirmPayment struct {
	UserID    uuid.UUID `json:"user_id"`
	PaymentID uuid.UUID `json:"payment_id"`
	RateID    uuid.UUID `json:"rate_id"`
}

type CancelPayment struct {
	UserID    uuid.UUID `json:"user_id"`
	PaymentID uuid.UUID `json:"payment_id"`
}
