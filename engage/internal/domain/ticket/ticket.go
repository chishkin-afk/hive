package ticketdomain

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrEmptyClientID        = errors.New("empty client_id")
	ErrEmptyAssignedTo      = errors.New("empty assigned_to")
	ErrInvalidContent       = errors.New("invalid content")
	ErrThreadIdAlreadySet   = errors.New("thread_id already set")
	ErrAssignedToAlreadySet = errors.New("assigned_to already set")
	ErrInvalidThreadID      = errors.New("invalid thread_id")
	ErrInvalidStatus        = errors.New("invalid status")
	ErrCantUpdateStatus     = errors.New("cant update status")
	ErrTicketAlreadyClosed  = errors.New("ticket already closed")
	ErrZeroCreatedAt        = errors.New("zero created_at")
	ErrZeroClosedAt         = errors.New("zero closed_at")

	ErrTicketNotFound        = errors.New("ticket not found")
	ErrTicketAlreadyExists   = errors.New("ticket already exists")
	ErrTicketAlreadyAssigned = errors.New("ticket already assigned")
	ErrTicketClosed          = errors.New("ticket already closed")
	ErrInvalidClient         = errors.New("invalid client")
	ErrMissingField          = errors.New("required field is missing")
	ErrDeadlock              = errors.New("database deadlock")
)

type Ticket struct {
	id         int64
	clientID   UserID
	threadID   int
	assignedTo UserID
	content    Content
	status     Status
	createdAt  time.Time
	closedAt   *time.Time
}

func New(
	clientID UserID,
	content Content,
) (Ticket, error) {
	if clientID.IsEmpty() {
		return Ticket{}, ErrEmptyClientID
	}
	if !content.IsValid() {
		return Ticket{}, ErrInvalidContent
	}

	return Ticket{
		clientID:  clientID,
		content:   content,
		status:    StatusNew,
		createdAt: time.Now().UTC(),
	}, nil
}

func NewFromStorage(
	id int64,
	clientID UserID,
	threadID int,
	assignedTo UserID,
	content Content,
	status Status,
	createdAt time.Time,
	closedAt *time.Time,
) (Ticket, error) {
	if clientID.IsEmpty() {
		return Ticket{}, ErrEmptyClientID
	}
	if !content.IsValid() {
		return Ticket{}, ErrInvalidContent
	}
	if createdAt.IsZero() {
		return Ticket{}, ErrZeroCreatedAt
	}
	if closedAt != nil && closedAt.IsZero() {
		return Ticket{}, ErrZeroClosedAt
	}

	return Ticket{
		id:         id,
		clientID:   clientID,
		threadID:   threadID,
		assignedTo: assignedTo,
		content:    content,
		status:     status,
		createdAt:  createdAt,
		closedAt:   closedAt,
	}, nil
}

func (t *Ticket) SetThreadID(threadID int) error {
	if t.threadID != 0 {
		return ErrThreadIdAlreadySet
	}
	if threadID == 0 {
		return ErrInvalidThreadID
	}

	t.threadID = threadID
	return nil
}

func (t *Ticket) SetAssignedTo(assignedTo UserID) error {
	if t.assignedTo.ID != 0 {
		return ErrAssignedToAlreadySet
	}
	if assignedTo.IsEmpty() {
		return ErrEmptyAssignedTo
	}

	t.assignedTo = assignedTo
	return nil
}

func (t *Ticket) UpdateStatus(status Status) error {
	if !status.IsValid() {
		return fmt.Errorf("%w: %d", ErrInvalidStatus, status)
	}
	if status < t.status || t.status == StatusClosed {
		return fmt.Errorf("%w: %d", ErrCantUpdateStatus, status)
	}

	t.status = status
	return nil
}

func (t *Ticket) Close() error {
	if t.closedAt != nil {
		return ErrTicketAlreadyClosed
	}

	t.status = StatusClosed

	now := time.Now().UTC()
	t.closedAt = &now

	return nil
}

func (t *Ticket) IsEquals(ticket Ticket) bool {
	if t.threadID == ticket.threadID &&
		t.assignedTo.ID == ticket.assignedTo.ID &&
		t.status == ticket.status {
		return true
	}

	return false
}

func (t *Ticket) ID() int64 {
	return t.id
}

func (t *Ticket) ClientID() UserID {
	return t.clientID
}

func (t *Ticket) ThreadID() int {
	return t.threadID
}

func (t *Ticket) AssignedTo() UserID {
	return t.assignedTo
}

func (t *Ticket) Content() Content {
	return t.content
}

func (t *Ticket) Status() Status {
	return t.status
}

func (t *Ticket) CreatedAt() time.Time {
	return t.createdAt
}

func (t *Ticket) ClosedAt() time.Time {
	if t.closedAt == nil {
		return time.Time{}
	}

	return *t.closedAt
}
