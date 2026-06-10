package clientredis

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/devathh/hive/backend/internal/modules/clients/domain/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientToBytes(t *testing.T) {
	now := time.Now().UTC()
	id := uuid.New()
	userID := uuid.New()
	offerID := uuid.New()

	clientEntity := client.From(id, userID, offerID, 5, now, now)

	bytes, err := clientToBytes(clientEntity)

	require.NoError(t, err)
	assert.NotEmpty(t, bytes)

	var record clientRecord
	err = json.Unmarshal(bytes, &record)
	require.NoError(t, err)

	assert.Equal(t, id, record.ID)
	assert.Equal(t, userID, record.UserID)
	assert.Equal(t, offerID, record.OfferID)
	assert.Equal(t, uint64(5), record.Visits)
	assert.WithinDuration(t, now, record.LastVisit, time.Second)
	assert.WithinDuration(t, now, record.CreatedAt, time.Second)
}

func TestBytesToClient(t *testing.T) {
	now := time.Now().UTC()
	id := uuid.New()
	userID := uuid.New()
	offerID := uuid.New()

	record := clientRecord{
		ID:        id,
		UserID:    userID,
		OfferID:   offerID,
		Visits:    10,
		LastVisit: now,
		CreatedAt: now,
	}

	bytes, err := json.Marshal(record)
	require.NoError(t, err)

	clientEntity, err := bytesToClient(bytes)

	require.NoError(t, err)
	require.NotNil(t, clientEntity)

	assert.Equal(t, id, clientEntity.ID())
	assert.Equal(t, userID, clientEntity.UserID())
	assert.Equal(t, offerID, clientEntity.OfferID())
	assert.Equal(t, uint64(10), clientEntity.Visits())
	assert.WithinDuration(t, now, clientEntity.LastVisit(), time.Second)
	assert.WithinDuration(t, now, clientEntity.CreatedAt(), time.Second)
}

func TestBytesToClientInvalidJSON(t *testing.T) {
	invalidBytes := []byte("invalid json")

	clientEntity, err := bytesToClient(invalidBytes)

	assert.Error(t, err)
	assert.Nil(t, clientEntity)
}

func TestRoundTrip(t *testing.T) {
	now := time.Now().UTC()
	id := uuid.New()
	userID := uuid.New()
	offerID := uuid.New()

	originalClient := client.From(id, userID, offerID, 7, now, now)

	bytes, err := clientToBytes(originalClient)
	require.NoError(t, err)

	restoredClient, err := bytesToClient(bytes)
	require.NoError(t, err)

	assert.Equal(t, originalClient.ID(), restoredClient.ID())
	assert.Equal(t, originalClient.UserID(), restoredClient.UserID())
	assert.Equal(t, originalClient.OfferID(), restoredClient.OfferID())
	assert.Equal(t, originalClient.Visits(), restoredClient.Visits())
	assert.WithinDuration(t, originalClient.LastVisit(), restoredClient.LastVisit(), time.Second)
	assert.WithinDuration(t, originalClient.CreatedAt(), restoredClient.CreatedAt(), time.Second)
}
