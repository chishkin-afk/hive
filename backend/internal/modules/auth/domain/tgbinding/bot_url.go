package tgbinding

import (
	"errors"
	"net/url"
	"strings"
)

var (
	ErrInvalidBotURL = errors.New("invalid bot's url")
)

type BotURL string

func (burl BotURL) Validate() error {
	if burl == "" {
		return ErrInvalidBotURL
	}

	parsedURL, err := url.Parse(string(burl))
	if err != nil {
		return ErrInvalidBotURL
	}

	queries := parsedURL.Query()
	if strings.TrimSpace(queries.Get("start")) == "" {
		return ErrInvalidBotURL
	}

	return nil
}

func (burl BotURL) String() string {
	return string(burl)
}
