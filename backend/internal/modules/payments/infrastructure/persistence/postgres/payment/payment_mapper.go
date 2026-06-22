package paymentpg

import "github.com/devathh/hive/backend/internal/modules/payments/domain/payment"

func paymentToRecord(payment *payment.Payment) *paymentRecord {
	return &paymentRecord{
		ID:          payment.ID(),
		PaymentID:   payment.PaymentID(),
		UserID:      payment.UserID(),
		AmountPenny: payment.AmountPenny(),
		StatusInt:   payment.Status().Value().Int(),
		CreatedAt:   payment.CreatedAt(),
		UpdatedAt:   payment.UpdatedAt(),
	}
}

func recordToPayment(record *paymentRecord) *payment.Payment {
	return payment.From(
		record.ID,
		record.PaymentID,
		record.UserID,
		record.AmountPenny,
		payment.FromStatus(record.StatusInt),
		record.CreatedAt,
		record.UpdatedAt,
	)
}
