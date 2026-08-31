package statecache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrNilPointer     = errors.New("nil pointer")
	ErrRecordNotFound = errors.New("record not found")
)

type StateRepo struct {
	client *redis.Client
	log    *slog.Logger
}

func NewStateRepo(client *redis.Client, log *slog.Logger) (*StateRepo, error) {
	if client == nil {
		return nil, ErrNilPointer
	}

	return &StateRepo{
		client: client,
		log:    log,
	}, nil
}

func (sr *StateRepo) Save(ctx context.Context, state State) error {
	sr.log.Debug("saving state to cache", slog.Int64("clientID", state.ClientID))

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return sr.client.Set(ctx, sr.key(state.ClientID), data, 15*time.Minute).Err()
}

func (sr *StateRepo) Get(ctx context.Context, clientID int64) (State, error) {
	sr.log.Debug("getting state from cache", slog.Int64("clientID", clientID))

	data, err := sr.client.Get(ctx, sr.key(clientID)).Result()
	if err != nil {
		if err == redis.Nil {
			return State{}, ErrRecordNotFound
		}

		return State{}, err
	}

	var state State
	if err = json.Unmarshal([]byte(data), &state); err != nil {
		return State{}, err
	}

	return state, nil
}

func (sr *StateRepo) Delete(ctx context.Context, clientID int64) error {
	sr.log.Debug("deleting state from cache", slog.Int64("clientID", clientID))

	return sr.client.Del(ctx, sr.key(clientID)).Err()
}

func (sr *StateRepo) key(clientID int64) string {
	return fmt.Sprintf("state:%d", clientID)
}
