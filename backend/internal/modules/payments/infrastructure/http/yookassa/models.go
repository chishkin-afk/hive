package yookassa

import "github.com/google/uuid"

type CreatePaymentRequest struct {
	Amount       Amount       `json:"amount"`
	Capture      bool         `json:"capture"`
	Confirmation Confirmation `json:"confirmation"`
	Metadata     Metadata     `json:"metadata"`
	Description  string       `json:"description"`
}

type Metadata struct {
	UserID string `json:"user_id"`
	RateID string `json:"rate_id"`
}

type Confirmation struct {
	Type      string `json:"type"`
	ReturnURL string `json:"return_url"`
}

type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type CreatePaymentResponse struct {
	ID           uuid.UUID `json:"id"`
	Confirmation struct {
		Type            string `json:"type"`
		ConfirmationURL string `json:"confirmation_url"`
	}
}

type GetPaymentResponse struct {
	Status string `json:"status"`
}
