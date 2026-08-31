package ticketcache

import (
	"time"

	ticketdomain "github.com/braunkc/hive-engage/internal/domain/ticket"
)

func ToRecord(ticket ticketdomain.Ticket) Ticket {
	var closedAt *time.Time
	if !ticket.ClosedAt().IsZero() {
		t := ticket.ClosedAt()
		closedAt = &t
	}

	return Ticket{
		ID:                 ticket.ID(),
		ClientID:           ticket.ClientID().ID,
		ClientUsername:     ticket.ClientID().Username,
		ThreadID:           ticket.ThreadID(),
		AssignedToID:       ticket.AssignedTo().ID,
		AssignedToUsername: ticket.AssignedTo().Username,
		Title:              ticket.Content().Title,
		Description:        ticket.Content().Description,
		Status:             int(ticket.Status()),
		CreatedAt:          ticket.CreatedAt(),
		ClosedAt:           closedAt,
	}
}
