package ticketapplication

import (
	"context"
	"errors"
	"log/slog"
	"time"

	ticketdomain "github.com/braunkc/hive-engage/internal/domain/ticket"
)

type TicketService struct {
	db    ticketdomain.TicketPersistenceRepository
	cache ticketdomain.TicketCacheRepository
	log   *slog.Logger
}

func NewTicketService(
	db ticketdomain.TicketPersistenceRepository,
	cache ticketdomain.TicketCacheRepository,
	log *slog.Logger,
) *TicketService {
	return &TicketService{
		db:    db,
		cache: cache,
		log:   log,
	}
}

func (ts *TicketService) Create(ctx context.Context, req CreateRequest) (Ticket, error) {
	ts.log.Debug("creating ticket",
		slog.Int64("clientID", req.ClientID),
		slog.String("username", req.Username),
		slog.String("title", req.Title),
		slog.String("description", req.Description),
	)

	ticket, err := ticketdomain.New(ticketdomain.UserID{
		ID:       req.ClientID,
		Username: req.Username,
	}, ticketdomain.Content{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		return Ticket{}, err
	}

	ticket, err = ts.db.Create(ctx, ticket)
	if err != nil {
		return Ticket{}, err
	}

	if err = ts.cache.Save(ctx, ticket); err != nil {
		return Ticket{}, err
	}

	return ToDTO(ticket), nil
}

func (ts *TicketService) Get(ctx context.Context, req GetRequest) (Ticket, error) {
	ts.log.Debug("getting ticket",
		slog.Int64("ticketID", req.ID),
	)

	ticket, err := ts.fetch(ctx, req.ID)
	if err != nil {
		return Ticket{}, err
	}

	return ToDTO(ticket), nil
}

func (ts *TicketService) GetOpenByClientID(ctx context.Context, req GetRequest) (Ticket, error) {
	ts.log.Debug("getting open ticket by client id",
		slog.Int64("clientID", req.ID),
	)

	ticket, err := ts.fetchByClientID(ctx, req.ID)
	if err != nil {
		return Ticket{}, err
	}

	if !ticket.ClosedAt().IsZero() {
		return Ticket{}, ticketdomain.ErrTicketNotFound
	}

	return ToDTO(ticket), nil
}

func (ts *TicketService) GetActiveByClientID(ctx context.Context, req GetRequest) (Ticket, error) {
	ts.log.Debug("getting active ticket by client id",
		slog.Int64("clientID", req.ID),
	)

	ticket, err := ts.fetchByClientID(ctx, req.ID)
	if err != nil {
		return Ticket{}, err
	}

	if ticket.Status() != ticketdomain.StatusInProgress {
		return Ticket{}, ticketdomain.ErrTicketNotFound
	}

	return ToDTO(ticket), nil
}

func (ts *TicketService) GetActiveByThreadID(ctx context.Context, req GetByThreadIDRequest) (Ticket, error) {
	ts.log.Debug("getting ticket by thread id",
		slog.Int("clientID", req.ID),
	)

	ticket, err := ts.fetchByThreadID(ctx, req.ID)
	if err != nil {
		return Ticket{}, err
	}

	if ticket.Status() != ticketdomain.StatusInProgress {
		return Ticket{}, ticketdomain.ErrTicketNotFound
	}

	return ToDTO(ticket), nil
}

func (ts *TicketService) Assign(ctx context.Context, req AssignRequest) (Ticket, error) {
	ts.log.Debug("assigning ticket",
		slog.Int64("ticketID", req.TicketID),
		slog.Int64("assignToID", req.AssignToID),
		slog.String("assignToUsername", req.AssignToUsername),
	)

	ticket, err := ts.fetch(ctx, req.TicketID)
	if err != nil {
		return Ticket{}, err
	}

	if err = ticket.SetAssignedTo(ticketdomain.UserID{
		ID:       req.AssignToID,
		Username: req.AssignToUsername,
	}); err != nil {
		return Ticket{}, err
	}

	if err = ticket.SetThreadID(req.ThreadID); err != nil {
		return Ticket{}, err
	}

	if err = ticket.UpdateStatus(ticketdomain.StatusInProgress); err != nil {
		return Ticket{}, err
	}

	ticket, err = ts.save(ctx, ticket)
	if err != nil {
		return Ticket{}, err
	}

	return ToDTO(ticket), nil
}

func (ts *TicketService) Close(ctx context.Context, req CloseRequest) (Ticket, error) {
	ts.log.Debug("closing ticket",
		slog.Int64("ticketID", req.TicketID),
	)

	ticket, err := ts.fetch(ctx, req.TicketID)
	if err != nil {
		return Ticket{}, err
	}

	if err = ticket.Close(); err != nil {
		return Ticket{}, err
	}

	ticket, err = ts.save(ctx, ticket)
	if err != nil {
		return Ticket{}, err
	}

	return ToDTO(ticket), nil
}

func (ts *TicketService) fetch(ctx context.Context, id int64) (ticketdomain.Ticket, error) {
	ticket, err := ts.cache.Get(ctx, id)
	if err == nil {
		return ticket, nil
	}
	if !errors.Is(err, ticketdomain.ErrTicketNotFound) {
		ts.log.Warn("failed to get ticket from cache", slog.Any("err", err))
	}

	ticket, err = ts.db.Get(ctx, id)
	if err != nil {
		return ticketdomain.Ticket{}, err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err = ts.cache.Save(ctx, ticket); err != nil {
			ts.log.Warn("failed to save ticket to cache", slog.Any("err", err))
		}
	}()

	return ticket, nil
}

func (ts *TicketService) fetchByClientID(ctx context.Context, id int64) (ticketdomain.Ticket, error) {
	ticket, err := ts.cache.GetByClientID(ctx, id)
	if err == nil {
		return ticket, nil
	}
	if !errors.Is(err, ticketdomain.ErrTicketNotFound) {
		ts.log.Warn("failed to get ticket from cache", slog.Any("err", err))
	}

	ticket, err = ts.db.GetByClientID(ctx, id)
	if err != nil {
		return ticketdomain.Ticket{}, err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err = ts.cache.Save(ctx, ticket); err != nil {
			ts.log.Warn("failed to save ticket to cache", slog.Any("err", err))
		}
	}()

	return ticket, nil
}

func (ts *TicketService) fetchByThreadID(ctx context.Context, id int) (ticketdomain.Ticket, error) {
	ticket, err := ts.cache.GetByThreadID(ctx, id)
	if err == nil {
		return ticket, nil
	}
	if !errors.Is(err, ticketdomain.ErrTicketNotFound) {
		ts.log.Warn("failed to get ticket from cache", slog.Any("err", err))
	}

	ticket, err = ts.db.GetByThreadID(ctx, id)
	if err != nil {
		return ticketdomain.Ticket{}, err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err = ts.cache.Save(ctx, ticket); err != nil {
			ts.log.Warn("failed to save ticket to cache", slog.Any("err", err))
		}
	}()

	return ticket, nil
}

func (ts *TicketService) save(ctx context.Context, ticket ticketdomain.Ticket) (ticketdomain.Ticket, error) {
	ticket, err := ts.db.Update(ctx, ticket)
	if err != nil {
		return ticketdomain.Ticket{}, err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err = ts.cache.Delete(ctx, ticket.ID()); err != nil {
			ts.log.Warn("failed to delete ticket from cache", slog.Any("err", err))
		}
	}()

	return ticket, nil
}
