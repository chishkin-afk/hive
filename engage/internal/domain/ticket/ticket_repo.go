package ticketdomain

import (
	"context"
)

type TicketPersistenceRepository interface {
	Create(ctx context.Context, ticket Ticket) (Ticket, error)
	Get(ctx context.Context, id int64) (Ticket, error)
	GetByClientID(ctx context.Context, id int64) (Ticket, error)
	GetByThreadID(ctx context.Context, id int) (Ticket, error)
	Update(ctx context.Context, ticket Ticket) (Ticket, error)
}

type TicketCacheRepository interface {
	Save(ctx context.Context, ticket Ticket) error
	Get(ctx context.Context, id int64) (Ticket, error)
	GetByClientID(ctx context.Context, id int64) (Ticket, error)
	GetByThreadID(ctx context.Context, id int) (Ticket, error)
	Delete(ctx context.Context, id int64) error
}
