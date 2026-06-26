package registrationpg

import (
	"time"

	"github.com/google/uuid"
)

type registrationRecord struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Fingerprint string
	CreatedAt   time.Time
}
