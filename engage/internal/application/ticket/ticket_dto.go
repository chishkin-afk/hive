package ticketapplication

import "time"

type Ticket struct {
	ID                   int64      `json:"id"`
	ClientID             UserID     `json:"client_id"`
	ThreadID             int        `json:"thread_id"`
	AssignedToID         UserID     `json:"assigned_to_id"`
	Content              Content    `json:"content"`
	Status               int        `json:"status"`
	ClientCloseMessageID int        `json:"client_close_message_id"`
	TopicCloseMessageID  int        `json:"topic_close_message_id"`
	CreatedAt            time.Time  `json:"created_at"`
	ClosedAt             *time.Time `json:"closed_at,omitempty"`
}

type UserID struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type Content struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type CreateRequest struct {
	ClientID    int64
	Username    string
	Title       string
	Description string
}

type GetRequest struct {
	ID int64
}

type GetByThreadIDRequest struct {
	ID int
}

type AssignRequest struct {
	TicketID         int64
	ThreadID         int
	AssignToID       int64
	AssignToUsername string
}

type UpdateStatusRequest struct {
	TicketID int64
	Status   int
}

type CloseRequest struct {
	TicketID int64
}
