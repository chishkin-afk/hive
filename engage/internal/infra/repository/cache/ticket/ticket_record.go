package ticketcache

import "time"

type Ticket struct {
	ID                 int64      `json:"id"`
	ClientID           int64      `json:"client_id"`
	ClientUsername     string     `json:"client_username"`
	ThreadID           int        `json:"thread_id"`
	AssignedToID       int64      `json:"assigned_to_id"`
	AssignedToUsername string     `json:"assigned_to_username"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	Status             int        `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
	ClosedAt           *time.Time `json:"closed_at"`
}
