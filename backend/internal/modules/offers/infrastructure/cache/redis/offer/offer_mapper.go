package offerredis

import (
	"encoding/json"
	"time"

	"github.com/devathh/hive/backend/internal/modules/offers/domain/offer"
)

func bytesToDomain(bytes []byte) (*offer.Offer, error) {
	var record offerRecord
	if err := json.Unmarshal(bytes, &record); err != nil {
		return nil, err
	}

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
	), nil
}

func domainToBytes(domain *offer.Offer) ([]byte, error) {
	return json.Marshal(offerRecord{
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
	})
}
