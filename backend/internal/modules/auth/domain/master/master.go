package master

import "github.com/chishkin-afk/hive/backend/internal/modules/auth/domain/user"

type Master struct {
	user *user.User
}

func New(
	email user.Email,
	firstname string,
	lastname string,
	surname string,
) (*Master, error) {
	u, err := user.New(email, firstname, lastname, surname)
	if err != nil {
		return nil, err
	}

	return &Master{
		user: u,
	}, nil
}

func (m *Master) User() *user.User {
	return m.user
}
