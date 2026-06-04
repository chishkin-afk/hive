package offerpg

import (
	"time"

	"github.com/devathh/hive/backend/internal/modules/offers/domain/offer"
)

func domainToRecord(domain *offer.Offer) *offerRecord {
	return &offerRecord{
		ID:                  domain.ID(),
		AuthorID:            domain.AuthorID(),
		Title:               domain.Title(),
		Description:         domain.Description(),
		Price:               domain.Price(),
		SlotDurationMinutes: domain.SlotDuration().Minutes(),
		SlotIntervalMinutes: domain.SlotInterval().Minutes(),
		IsActive:            domain.IsActive().Bool(),
		CreatedAt:           domain.CreatedAt(),
		UpdatedAt:           domain.UpdatedAt(),
	}
}

func recordToDomain(record *offerRecord) *offer.Offer {
	return offer.From(
		record.ID,
		record.AuthorID,
		record.Title,
		record.Description,
		record.Price,
		offer.SlotDuration(time.Duration(record.SlotDurationMinutes)*time.Minute),
		offer.SlotInterval(time.Duration(record.SlotIntervalMinutes)*time.Minute),
		offer.IsActive(record.IsActive),
		record.CreatedAt,
		record.UpdatedAt,
	)
}
