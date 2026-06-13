package schedule

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSchedule_Success(t *testing.T) {
	userID := uuid.New()
	dayOfWeek := uint8(1) // monday
	startTime := DayTime{
		hour:   8,
		minute: 0,
	}
	endTime := DayTime{
		hour:   22,
		minute: 0,
	}
	breaks := []Break{
		{
			startTime: DayTime{
				hour: 12,
			},
			endTime: DayTime{
				hour: 13,
			},
		},
	}

	schedule, err := New(
		userID,
		dayOfWeek,
		startTime,
		endTime,
		breaks,
	)

	require.NoError(t, err)
	require.NotEmpty(t, schedule.id)

	assert.Equal(t, schedule.userID, userID)
	assert.Equal(t, schedule.dayOfWeek, dayOfWeek)
	assert.Equal(t, schedule.startTime, startTime)
	assert.Equal(t, schedule.endTime, endTime)
	assert.Equal(t, schedule.breaks, breaks)
}

func TestNewSchedule_Invalid(t *testing.T) {
	type inputData struct {
		userID    uuid.UUID
		dayOfWeek uint8
		startTime DayTime
		endTime   DayTime
		breaks    []Break
	}

	testCases := []struct {
		name     string
		input    inputData
		expected error
	}{
		{
			name: "invalid_user_id",
			input: inputData{
				userID:    uuid.Nil,
				dayOfWeek: 1,
				startTime: DayTime{
					hour: 8,
				},
				endTime: DayTime{
					hour: 22,
				},
				breaks: []Break{},
			},
			expected: ErrInvalidUserID,
		},
		{
			name: "invalid_day_of_week",
			input: inputData{
				userID:    uuid.New(),
				dayOfWeek: 7,
				startTime: DayTime{
					hour: 8,
				},
				endTime: DayTime{
					hour: 22,
				},
				breaks: []Break{},
			},
			expected: ErrInvalidDayOfWeek,
		},
		{
			name: "invalid_duration",
			input: inputData{
				userID:    uuid.New(),
				dayOfWeek: 1,
				startTime: DayTime{
					hour: 8,
				},
				endTime: DayTime{
					hour:   8,
					minute: 4,
				},
				breaks: []Break{},
			},
			expected: ErrInvalidSchedule,
		},
		{
			name: "invalid_break",
			input: inputData{
				userID:    uuid.New(),
				dayOfWeek: 1,
				startTime: DayTime{
					hour: 8,
				},
				endTime: DayTime{
					hour: 22,
				},
				breaks: []Break{
					{
						startTime: DayTime{
							hour: 7,
						},
						endTime: DayTime{
							hour: 8,
						},
					},
				},
			},
			expected: ErrInvalidBreak,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(
				tc.input.userID,
				tc.input.dayOfWeek,
				tc.input.startTime,
				tc.input.endTime,
				tc.input.breaks,
			)

			require.Error(t, err)
			assert.EqualError(t, err, tc.expected.Error())
		})
	}
}
