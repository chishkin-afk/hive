package tgbinding

import (
	"errors"
	"net/url"

	"github.com/google/uuid"
)

var (
	ErrInvalidBotURL = errors.New("invalid bot's url")
)

type BotURL string

func (burl BotURL) String() string {
	return string(burl)
}

func NewBotURL(rawURL string, bindingID uuid.UUID) (BotURL, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", ErrInvalidBotURL
	}

	queries := parsedURL.Query()
	queries.Add("start", bindingID.String())

	parsedURL.RawQuery = queries.Encode()
	return BotURL(parsedURL.String()), nil
}
