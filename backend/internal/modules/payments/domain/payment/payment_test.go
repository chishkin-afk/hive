package payment

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPayment_Success(t *testing.T) {
	var (
		paymentID          = uuid.New()
		userID             = uuid.New()
		amountPenny uint64 = 39999
	)

	now := time.Now().UTC()
	payment, err := New(
		paymentID,
		userID,
		amountPenny,
	)

	require.NoError(t, err)

	assert.NotEmpty(t, payment.ID())
	assert.Equal(t, payment.PaymentID(), paymentID)
	assert.Equal(t, payment.UserID(), userID)
	assert.Equal(t, payment.AmountPenny(), amountPenny)
	assert.False(t, payment.IsConfirmed())
	assert.WithinDuration(t, now, payment.CreatedAt(), 100*time.Millisecond)
	assert.Nil(t, payment.ConfirmedAt())
}

func TestNewPayment_Invalid(t *testing.T) {
	type inputData struct {
		paymentID   uuid.UUID
		userID      uuid.UUID
		amountPenny uint64
	}

	testCases := []struct {
		name     string
		data     inputData
		expected error
	}{
		{
			name: "invalid_payment_id",
			data: inputData{
				paymentID:   uuid.Nil,
				userID:      uuid.New(),
				amountPenny: 39999,
			},
			expected: ErrEmptyPaymentID,
		},
		{
			name: "invalid_user_id",
			data: inputData{
				paymentID:   uuid.New(),
				userID:      uuid.Nil,
				amountPenny: 39999,
			},
			expected: ErrEmptyUserID,
		},
		{
			name: "invalid_amount_penny",
			data: inputData{
				paymentID:   uuid.New(),
				userID:      uuid.New(),
				amountPenny: 1,
			},
			expected: ErrInvalidAmount,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(
				tc.data.paymentID,
				tc.data.userID,
				tc.data.amountPenny,
			)

			require.Error(t, err)
			assert.EqualError(t, err, tc.expected.Error())
		})
	}
}
