package clientredis

import (
	"encoding/json"

	"github.com/devathh/hive/backend/internal/modules/clients/domain/client"
)

func clientToBytes(client *client.Client) ([]byte, error) {
	return json.Marshal(clientRecord{
		ID:        client.ID(),
		UserID:    client.UserID(),
		OfferID:   client.OfferID(),
		Visits:    client.Visits(),
		LastVisit: client.LastVisit(),
		CreatedAt: client.CreatedAt(),
	})
}

func bytesToClient(bytes []byte) (*client.Client, error) {
	var record clientRecord
	if err := json.Unmarshal(bytes, &record); err != nil {
		return nil, err
	}

	return client.From(
		record.ID,
		record.UserID,
		record.OfferID,
		record.Visits,
		record.LastVisit,
		record.CreatedAt,
	), nil
}
