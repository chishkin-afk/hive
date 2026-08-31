package ticketapplication

import (
	"time"

	ticketdomain "github.com/braunkc/hive-engage/internal/domain/ticket"
)

func ToDTO(ticket ticketdomain.Ticket) Ticket {
	var closedAt *time.Time
	if tClosedAt := ticket.ClosedAt(); !tClosedAt.IsZero() {
		closedAt = &tClosedAt
	}

	return Ticket{
		ID: ticket.ID(),
		ClientID: UserID{
			ID:       ticket.ClientID().ID,
			Username: ticket.ClientID().Username,
		},
		ThreadID: ticket.ThreadID(),
		AssignedToID: UserID{
			ID:       ticket.AssignedTo().ID,
			Username: ticket.AssignedTo().Username,
		},
		Content: Content{
			Title:       ticket.Content().Title,
			Description: ticket.Content().Description,
		},
		Status:    int(ticket.Status()),
		CreatedAt: ticket.CreatedAt(),
		ClosedAt:  closedAt,
	}
}
