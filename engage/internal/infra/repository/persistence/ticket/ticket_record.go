package ticketpersistence

import "time"

type Ticket struct {
	ID                 int64
	ClientID           int64
	ClientUsername     string
	ThreadID           int
	AssignedToID       int64
	AssignedToUsername string
	Title              string
	Description        string
	Status             int
	CreatedAt          time.Time
	ClosedAt           *time.Time
}
