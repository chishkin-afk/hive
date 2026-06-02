package requests

import "github.com/google/uuid"

type Register struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
	Password  string `json:"password"` // can be empty
}

type ConfirmUser struct {
	UserID uuid.UUID `json:"user_id"`
	Code   int       `json:"code"`
}
