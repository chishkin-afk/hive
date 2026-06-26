package registrationpg

import "github.com/devathh/hive/backend/internal/modules/auth/domain/registration"

func registrationToRecord(registration *registration.Registration) *registrationRecord {
	return &registrationRecord{
		ID:          registration.ID(),
		UserID:      registration.UserID(),
		Fingerprint: registration.Fingerprint(),
		CreatedAt:   registration.CreatedAt(),
	}
}

func recordToRegistration(record *registrationRecord) *registration.Registration {
	return registration.From(
		record.ID,
		record.UserID,
		record.Fingerprint,
		record.CreatedAt,
	)
}
