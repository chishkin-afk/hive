package booking

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDayTime_Success(t *testing.T) {
	var (
		hours   uint8 = 22
		minutes uint8 = 34
	)

	dayTime, err := NewDayTime(hours, minutes)

	require.NoError(t, err)
	assert.Equal(t, dayTime.Hour(), hours)
	assert.Equal(t, dayTime.Minute(), minutes)
	assert.Equal(t, dayTime.AllMinutes(), int(minutes)+(int(hours)*60))
}

func TestNewDayTime_Invalid(t *testing.T) {
	type inputData struct {
		hours   uint8
		minutes uint8
	}

	testCases := []struct {
		name     string
		input    inputData
		expected error
	}{
		{
			name: "too_much_hours",
			input: inputData{
				hours:   24,
				minutes: 34,
			},
			expected: ErrInvalidHour,
		},
		{
			name: "too_much_minutes",
			input: inputData{
				hours:   22,
				minutes: 60,
			},
			expected: ErrInvalidMinute,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewDayTime(tc.input.hours, tc.input.minutes)

			require.Error(t, err)
			assert.EqualError(t, err, tc.expected.Error())
		})
	}
}
