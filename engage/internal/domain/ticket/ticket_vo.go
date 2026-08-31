package ticketdomain

import "strings"

const (
	StatusNew Status = iota
	StatusInProgress
	StatusClosed
)

type UserID struct {
	ID       int64
	Username string
}

func (ui UserID) IsEmpty() bool {
	if ui.ID == 0 {
		return true
	}

	return false
}

type Content struct {
	Title       string
	Description string
}

func (c Content) IsValid() bool {
	if strings.TrimSpace(c.Title) == "" ||
		strings.TrimSpace(c.Description) == "" {
		return false
	}

	return true
}

type Status int

func (s Status) IsValid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusClosed:
		return true
	default:
		return false
	}
}
