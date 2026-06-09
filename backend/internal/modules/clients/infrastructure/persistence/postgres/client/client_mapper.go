package clientpg

import "github.com/devathh/hive/backend/internal/modules/clients/domain/client"

func clientToRecord(client *client.Client) *clientRecord {
	return &clientRecord{
		ID:        client.ID(),
		UserID:    client.UserID(),
		OfferID:   client.OfferID(),
		Visits:    client.Visits(),
		LastVisit: client.LastVisit(),
		CreatedAt: client.CreatedAt(),
	}
}

func recordToClient(record *clientRecord) *client.Client {
	return client.From(
		record.ID,
		record.UserID,
		record.OfferID,
		record.Visits,
		record.LastVisit,
		record.CreatedAt,
	)
}
