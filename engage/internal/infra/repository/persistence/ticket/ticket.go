package ticketpersistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	ticketdomain "github.com/braunkc/hive-engage/internal/domain/ticket"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNilPointer = errors.New("nil pointer")
)

type ticketRepo struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewTicketRepo(db *pgxpool.Pool, log *slog.Logger) (ticketdomain.TicketPersistenceRepository, error) {
	if db == nil {
		return nil, ErrNilPointer
	}

	return &ticketRepo{
		db:  db,
		log: log,
	}, nil
}

func (tr *ticketRepo) Create(ctx context.Context, ticket ticketdomain.Ticket) (ticketdomain.Ticket, error) {
	tr.log.Debug("creating ticket at db", slog.Int64("client_id", ticket.ClientID().ID))

	t := ToRecord(ticket)
	query, args, err := sq.Insert("tickets").
		Columns(
			"client_id",
			"client_username",
			"thread_id",
			"assigned_to_id",
			"assigned_to_username",
			"title",
			"description",
			"status",
			"created_at",
			"closed_at",
		).
		Values(
			t.ClientID,
			t.ClientUsername,
			t.ThreadID,
			t.AssignedToID,
			t.AssignedToUsername,
			t.Title,
			t.Description,
			t.Status,
			t.CreatedAt,
			t.ClosedAt,
		).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return ticketdomain.Ticket{}, err
	}

	err = tr.db.QueryRow(ctx, query, args...).Scan(&t.ID)
	if err != nil {
		return ticketdomain.Ticket{}, handlePgError(err)
	}

	return tr.mapToDomain(t)
}

func (tr *ticketRepo) Get(ctx context.Context, id int64) (ticketdomain.Ticket, error) {
	tr.log.Debug("getting ticket from db", slog.Int64("id", id))

	query, args, err := sq.
		Select(
			"id",
			"client_id",
			"client_username",
			"thread_id",
			"assigned_to_id",
			"assigned_to_username",
			"title",
			"description",
			"status",
			"created_at",
			"closed_at",
		).
		From("tickets").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return ticketdomain.Ticket{}, err
	}

	row := tr.db.QueryRow(ctx, query, args...)

	t, err := scanTicket(row)
	if err != nil {
		return ticketdomain.Ticket{}, handlePgError(err)
	}

	return tr.mapToDomain(t)
}

func (tr *ticketRepo) GetByClientID(ctx context.Context, id int64) (ticketdomain.Ticket, error) {
	tr.log.Debug("getting ticket by client id from db", slog.Int64("id", id))

	query, args, err := sq.
		Select(
			"id",
			"client_id",
			"client_username",
			"thread_id",
			"assigned_to_id",
			"assigned_to_username",
			"title",
			"description",
			"status",
			"created_at",
			"closed_at",
		).
		From("tickets").
		Where(sq.And{sq.Eq{"client_id": id}, sq.Eq{"closed_at": nil}}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return ticketdomain.Ticket{}, err
	}

	row := tr.db.QueryRow(ctx, query, args...)

	t, err := scanTicket(row)
	if err != nil {
		return ticketdomain.Ticket{}, handlePgError(err)
	}

	return tr.mapToDomain(t)
}

func (tr *ticketRepo) GetByThreadID(ctx context.Context, id int) (ticketdomain.Ticket, error) {
	tr.log.Debug("getting ticket by thread id from db", slog.Int("id", id))

	query, args, err := sq.
		Select(
			"id",
			"client_id",
			"client_username",
			"thread_id",
			"assigned_to_id",
			"assigned_to_username",
			"title",
			"description",
			"status",
			"created_at",
			"closed_at",
		).
		From("tickets").
		Where(sq.Eq{"thread_id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return ticketdomain.Ticket{}, err
	}

	row := tr.db.QueryRow(ctx, query, args...)

	t, err := scanTicket(row)
	if err != nil {
		return ticketdomain.Ticket{}, handlePgError(err)
	}

	return tr.mapToDomain(t)
}

func (tr *ticketRepo) Update(ctx context.Context, ticket ticketdomain.Ticket) (ticketdomain.Ticket, error) {
	tr.log.Debug("updating ticket at db", slog.Int64("id", ticket.ID()))

	t := ToRecord(ticket)
	query, args, err := sq.Update("tickets").
		Set("client_id", t.ClientID).
		Set("client_username", t.ClientUsername).
		Set("thread_id", t.ThreadID).
		Set("assigned_to_id", t.AssignedToID).
		Set("assigned_to_username", t.AssignedToUsername).
		Set("title", t.Title).
		Set("description", t.Description).
		Set("status", t.Status).
		Set("closed_at", t.ClosedAt).
		Where(sq.Eq{"id": t.ID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return ticketdomain.Ticket{}, err
	}

	_, err = tr.db.Exec(ctx, query, args...)
	if err != nil {
		return ticketdomain.Ticket{}, handlePgError(err)
	}

	return ticket, nil
}

func scanTicket(row pgx.Row) (Ticket, error) {
	var t Ticket
	err := row.Scan(
		&t.ID, &t.ClientID, &t.ClientUsername, &t.ThreadID, &t.AssignedToID, &t.AssignedToUsername,
		&t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.ClosedAt,
	)
	if err != nil {
		return Ticket{}, err
	}
	return t, nil
}

func (tr *ticketRepo) mapToDomain(t Ticket) (ticketdomain.Ticket, error) {
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

func handlePgError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return ticketdomain.ErrTicketNotFound
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.Code {
	case "23505":
		return fmt.Errorf("%w: %s", ticketdomain.ErrTicketAlreadyExists, pgErr.Detail)
	case "23502":
		return fmt.Errorf("%w: %s", ticketdomain.ErrMissingField, pgErr.ColumnName)
	case "23503":
		return fmt.Errorf("%w: %s", ticketdomain.ErrInvalidClient, pgErr.Detail)
	case "40P01":
		return fmt.Errorf("%w: %s", ticketdomain.ErrDeadlock, pgErr.Message)
	default:
		return fmt.Errorf("database error (code %s): %w", pgErr.Code, err)
	}
}
