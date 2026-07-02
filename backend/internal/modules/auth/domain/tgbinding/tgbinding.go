package tgbinding

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyUserID = errors.New("user's id is empty")
)

type TgBinding struct {
	id        uuid.UUID
	userID    uuid.UUID
	botURL    BotURL
	createdAt time.Time
}

func New(
	userID uuid.UUID,
	rawBotURL string,
) (*TgBinding, error) {
	if userID == uuid.Nil {
		return nil, ErrEmptyUserID
	}

	bindingID := uuid.New()
	botURL, err := NewBotURL(rawBotURL, bindingID)
	if err != nil {
		return nil, err
	}

	return &TgBinding{
		id:        bindingID,
		userID:    userID,
		botURL:    botURL,
		createdAt: time.Now().UTC(),
	}, nil
}

func From(
	id uuid.UUID,
	userID uuid.UUID,
	botURL BotURL,
	createdAt time.Time,
) *TgBinding {
	return &TgBinding{
		id:        id,
		userID:    userID,
		botURL:    botURL,
		createdAt: createdAt,
	}
}

func (tb *TgBinding) ID() uuid.UUID {
	return tb.id
}

func (tb *TgBinding) UserID() uuid.UUID {
	return tb.userID
}

func (tb *TgBinding) BotURL() BotURL {
	return tb.botURL
}

func (tb *TgBinding) CreatedAt() time.Time {
	return tb.createdAt
}
