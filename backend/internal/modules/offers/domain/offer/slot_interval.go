package offer

import (
	"errors"
	"time"
)

var (
	ErrInvalidSlotInterval = errors.New("slot's interval must be less than 24 hour")
)

type SlotInterval time.Duration

func (si SlotInterval) Validate() error {
	if si <= 0 || time.Duration(si) >= 24*time.Hour {
		return ErrInvalidSlotInterval
	}

	return nil
}

func (si SlotInterval) Minutes() int64 {
	return int64(time.Duration(si).Minutes())
}
