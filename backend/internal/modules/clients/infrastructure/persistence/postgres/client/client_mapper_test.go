package clientpg

import (
	"testing"
	"time"

	"github.com/devathh/hive/backend/internal/modules/clients/domain/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientToRecord(t *testing.T) {
	now := time.Now().UTC()
	id := uuid.New()
	userID := uuid.New()
	offerID := uuid.New()

	clientEntity := client.From(id, userID, offerID, 3, now, now)

	record := clientToRecord(clientEntity)

	require.NotNil(t, record)
	assert.Equal(t, id, record.ID)
	assert.Equal(t, userID, record.UserID)
	assert.Equal(t, offerID, record.OfferID)
	assert.Equal(t, uint64(3), record.Visits)
	assert.WithinDuration(t, now, record.LastVisit, time.Second)
	assert.WithinDuration(t, now, record.CreatedAt, time.Second)
}

func TestRecordToClient(t *testing.T) {
	now := time.Now().UTC()
	id := uuid.New()
	userID := uuid.New()
	offerID := uuid.New()

	record := &clientRecord{
		ID:        id,
		UserID:    userID,
		OfferID:   offerID,
		Visits:    15,
		LastVisit: now,
		CreatedAt: now,
	}

	clientEntity := recordToClient(record)

	require.NotNil(t, clientEntity)
	assert.Equal(t, id, clientEntity.ID())
	assert.Equal(t, userID, clientEntity.UserID())
	assert.Equal(t, offerID, clientEntity.OfferID())
	assert.Equal(t, uint64(15), clientEntity.Visits())
	assert.WithinDuration(t, now, clientEntity.LastVisit(), time.Second)
	assert.WithinDuration(t, now, clientEntity.CreatedAt(), time.Second)
}

func TestRecordToClientWithNilRecord(t *testing.T) {
	assert.Panics(t, func() {
		var record *clientRecord = nil
		recordToClient(record)
	})
}

func TestRoundTrip(t *testing.T) {
	now := time.Now().UTC()
	id := uuid.New()
	userID := uuid.New()
	offerID := uuid.New()

	originalClient := client.From(id, userID, offerID, 9, now, now)

	record := clientToRecord(originalClient)
	restoredClient := recordToClient(record)

	assert.Equal(t, originalClient.ID(), restoredClient.ID())
	assert.Equal(t, originalClient.UserID(), restoredClient.UserID())
	assert.Equal(t, originalClient.OfferID(), restoredClient.OfferID())
	assert.Equal(t, originalClient.Visits(), restoredClient.Visits())
	assert.WithinDuration(t, originalClient.LastVisit(), restoredClient.LastVisit(), time.Second)
	assert.WithinDuration(t, originalClient.CreatedAt(), restoredClient.CreatedAt(), time.Second)
}
