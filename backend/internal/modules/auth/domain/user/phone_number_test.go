package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPhoneNumber_Success(t *testing.T) {
	phone := "+79998887766"

	phoneNumber, err := PhoneNumber(phone).Standart()

	require.NoError(t, err)
	assert.NotEqual(t, phoneNumber.String(), phone)
	assert.Equal(t, phoneNumber.String(), "79998887766")
}

func TestNewPhoneNumber_Invalid(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid_region",
			input: "+3902344556655",
		},
		{
			name:  "invalid_phone",
			input: "777",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := PhoneNumber(tc.input).Standart()

			require.Error(t, err)
		})
	}
}
