package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBreak_Success(t *testing.T) {
	workStart, _ := NewDayTime(8, 20)
	workEnd, _ := NewDayTime(23, 0)

	start, _ := NewDayTime(22, 34)
	end, _ := NewDayTime(22, 39)

	br, err := NewBreak(start, end)

	require.NoError(t, err)
	assert.Equal(t, br.StartTime(), start)
	assert.Equal(t, br.EndTime(), end)
	assert.True(t, br.IsValid(workStart, workEnd))
}

func TestNewBreak_Invalid(t *testing.T) {
	type inputData struct {
		startTime DayTime
		endTime   DayTime
	}

	testCases := []struct {
		name     string
		input    inputData
		expected error
	}{
		{
			name: "invalid_duration",
			input: inputData{
				startTime: DayTime{
					hour: 8,
				},
				endTime: DayTime{
					hour:   8,
					minute: 1,
				},
			},
			expected: ErrInvalidBreak,
		},
		{
			name: "too_large_duration",
			input: inputData{
				startTime: DayTime{
					hour:   0,
					minute: 0,
				},
				endTime: DayTime{
					hour: 24,
				},
			},
			expected: ErrInvalidBreak,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewBreak(tc.input.startTime, tc.input.endTime)

			require.Error(t, err)
			assert.EqualError(t, err, tc.expected.Error())
		})
	}
}

func TestBreakIsValid(t *testing.T) {
	workStart, _ := NewDayTime(8, 0)
	workEnd, _ := NewDayTime(22, 0)

	type inputData struct {
		startTime DayTime
		endTime   DayTime
	}

	testCases := []struct {
		name     string
		input    inputData
		expected bool
	}{
		{
			name: "invalid_start_time",
			input: inputData{
				startTime: DayTime{
					hour:   7,
					minute: 59,
				},
				endTime: DayTime{
					hour:   21,
					minute: 59,
				},
			},
			expected: false,
		},
		{
			name: "invalid_end_time",
			input: inputData{
				startTime: DayTime{
					hour:   8,
					minute: 0,
				},
				endTime: DayTime{
					hour:   22,
					minute: 1,
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			br, err := NewBreak(tc.input.startTime, tc.input.endTime)

			require.NoError(t, err)
			assert.Equal(t, br.StartTime(), tc.input.startTime)
			assert.Equal(t, br.EndTime(), tc.input.endTime)
			assert.Equal(t, br.IsValid(workStart, workEnd), tc.expected)
		})
	}
}
