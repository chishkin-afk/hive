package client

import (
	"time"

	"github.com/chishkin-afk/hive/backend/internal/modules/auth/domain/user"
)

type Client struct {
	user        *user.User
	lastVisitAt *time.Time
}

func New(
	email user.Email,
	firstname string,
	lastname string,
	surname string,
) (*Client, error) {
	u, err := user.New(email, firstname, lastname, surname)
	if err != nil {
		return nil, err
	}

	return &Client{
		user: u,
	}, nil
}

func (c *Client) AddVisit() {
	lastVisitAt := time.Now().UTC()
	c.lastVisitAt = &lastVisitAt
}

func (c *Client) User() *user.User {
	return c.user
}

func (c *Client) LastVisitAt() *time.Time {
	return c.lastVisitAt
}
