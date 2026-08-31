package ticketcache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	ticketdomain "github.com/braunkc/hive-engage/internal/domain/ticket"
	"github.com/redis/go-redis/v9"
)

var (
	ErrNilPointer = errors.New("nil pointer")
)

type ticketRepo struct {
	client *redis.Client
	log    *slog.Logger
}

func NewTicketRepo(client *redis.Client, log *slog.Logger) (ticketdomain.TicketCacheRepository, error) {
	if client == nil {
		return nil, ErrNilPointer
	}

	return &ticketRepo{
		client: client,
		log:    log,
	}, nil
}

func (tr *ticketRepo) Save(ctx context.Context, ticket ticketdomain.Ticket) error {
	tr.log.Debug("saving ticket to cache", slog.Int64("id", ticket.ID()))

	t := ToRecord(ticket)
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	tx := tr.client.TxPipeline()
	tx.Set(ctx, key(ticket.ID()), data, 15*time.Minute)
	tx.Set(ctx, keyClientID(ticket.ClientID().ID), ticket.ID(), 15*time.Minute)

	if ticket.ThreadID() != 0 {
		tx.Set(ctx, keyThreadID(ticket.ThreadID()), ticket.ID(), 15*time.Minute)
	}

	_, err = tx.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (tr *ticketRepo) Get(ctx context.Context, id int64) (ticketdomain.Ticket, error) {
	tr.log.Debug("getting ticket from cache", slog.Int64("id", id))

	data, err := tr.client.Get(ctx, key(id)).Result()
	if err != nil {
		if err == redis.Nil {
			return ticketdomain.Ticket{}, ticketdomain.ErrTicketNotFound
		}

		return ticketdomain.Ticket{}, err
	}

	var t Ticket
	if err = json.Unmarshal([]byte(data), &t); err != nil {
		return ticketdomain.Ticket{}, err
	}

	return mapToDomain(t)
}

func (tr *ticketRepo) GetByClientID(ctx context.Context, id int64) (ticketdomain.Ticket, error) {
	tr.log.Debug("getting ticket by client id from cache", slog.Int64("id", id))

	ticketIDStr, err := tr.client.Get(ctx, keyClientID(id)).Result()
	if err != nil {
		if err == redis.Nil {
			return ticketdomain.Ticket{}, ticketdomain.ErrTicketNotFound
		}

		return ticketdomain.Ticket{}, err
	}

	ticketID, err := strconv.ParseInt(ticketIDStr, 10, 64)
	if err != nil {
		return ticketdomain.Ticket{}, err
	}

	ticket, err := tr.Get(ctx, ticketID)
	if err != nil {
		return ticketdomain.Ticket{}, err
	}

	return ticket, nil
}

func (tr *ticketRepo) GetByThreadID(ctx context.Context, id int) (ticketdomain.Ticket, error) {
	tr.log.Debug("getting ticket by thread id from cache", slog.Int("id", id))

	ticketIDStr, err := tr.client.Get(ctx, keyThreadID(id)).Result()
	if err != nil {
		if err == redis.Nil {
			return ticketdomain.Ticket{}, ticketdomain.ErrTicketNotFound
		}

		return ticketdomain.Ticket{}, err
	}

	ticketID, err := strconv.ParseInt(ticketIDStr, 10, 64)
	if err != nil {
		return ticketdomain.Ticket{}, err
	}

	ticket, err := tr.Get(ctx, ticketID)
	if err != nil {
		return ticketdomain.Ticket{}, err
	}

	return ticket, nil
}

func (tr *ticketRepo) Delete(ctx context.Context, id int64) error {
	tr.log.Debug("deleting ticket from cache", slog.Int64("id", id))

	ticket, err := tr.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ticketdomain.ErrTicketNotFound) {
			return nil
		}

		return err
	}

	tx := tr.client.TxPipeline()
	tx.Del(ctx, key(id))
	tx.Del(ctx, keyClientID(ticket.ClientID().ID))
	tx.Del(ctx, keyThreadID(ticket.ThreadID()))

	_, err = tx.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func key(id int64) string {
	return fmt.Sprintf("ticket:%d", id)
}

func keyClientID(clientID int64) string {
	return fmt.Sprintf("user_ticket:%d", clientID)
}

func keyThreadID(threadID int) string {
	return fmt.Sprintf("thread_ticket:%d", threadID)
}

func mapToDomain(t Ticket) (ticketdomain.Ticket, error) {
	return ticketdomain.NewFromStorage(
		t.ID,
		ticketdomain.UserID{
			ID:       t.ClientID,
			Username: t.ClientUsername,
		},
		t.ThreadID,
		ticketdomain.UserID{
			ID:       t.AssignedToID,
			Username: t.AssignedToUsername,
		},
		ticketdomain.Content{
			Title:       t.Title,
			Description: t.Description,
		},
		ticketdomain.Status(t.Status),
		t.CreatedAt,
		t.ClosedAt,
	)
}
