package offer

import (
	"errors"
	"time"
)

var (
	ErrInvalidSlotDuration = errors.New("slot's duration must be less than 24 hours")
)

type SlotDuration time.Duration

func (sd SlotDuration) Validate() error {
	if sd <= 0 || time.Duration(sd) > 24*time.Hour {
		return ErrInvalidSlotDuration
	}

	return nil
}

func (sd SlotDuration) Minutes() int64 {
	return int64(time.Duration(sd).Minutes())
}
