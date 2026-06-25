package paymentredis

import (
	"encoding/json"

	"github.com/devathh/hive/backend/internal/modules/payments/domain/payment"
)

func paymentToBytes(payment *payment.Payment) ([]byte, error) {
	return json.Marshal(paymentRecord{
		ID:          payment.ID(),
		PaymentID:   payment.PaymentID(),
		UserID:      payment.UserID(),
		AmountPenny: payment.AmountPenny(),
		StatusInt:   payment.Status().Value().Int(),
		CreatedAt:   payment.CreatedAt(),
		UpdatedAt:   payment.UpdatedAt(),
	})
}

func bytesToPayment(bytes []byte) (*payment.Payment, error) {
	var record paymentRecord
	if err := json.Unmarshal(bytes, &record); err != nil {
		return nil, err
	}

	return payment.From(
		record.ID,
		record.PaymentID,
		record.UserID,
		record.AmountPenny,
		payment.FromStatus(record.StatusInt),
		record.CreatedAt,
		record.UpdatedAt,
	), nil
}
