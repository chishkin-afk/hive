package paymentpg

import "github.com/devathh/hive/backend/internal/modules/payments/domain/payment"

func paymentToRecord(payment *payment.Payment) *paymentRecord {
	return &paymentRecord{
		ID:          payment.ID(),
		PaymentID:   payment.PaymentID(),
		UserID:      payment.UserID(),
		AmountPenny: payment.AmountPenny(),
		IsConfirmed: payment.IsConfirmed(),
		CreatedAt:   payment.CreatedAt(),
		ConfirmedAt: payment.ConfirmedAt(),
	}
}

func recordToPayment(record *paymentRecord) *payment.Payment {
	return payment.From(
		record.ID,
		record.PaymentID,
		record.UserID,
		record.AmountPenny,
		record.IsConfirmed,
		record.CreatedAt,
		record.ConfirmedAt,
	)
}
