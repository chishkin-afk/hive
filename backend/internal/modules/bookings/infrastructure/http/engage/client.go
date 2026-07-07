package engagehttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
)

type engageHTTP struct {
	cfg    *config.Config
	log    *slog.Logger
	client *http.Client
}

func New(cfg *config.Config, log *slog.Logger, client *http.Client) *engageHTTP {
	return &engageHTTP{
		cfg:    cfg,
		log:    log,
		client: client,
	}
}

func (e *engageHTTP) SendMessage(ctx context.Context, chatID int64, msg string) error {
	req, err := e.createSendRequest(ctx, chatID, msg)
	if err != nil {
		return err
	}

	return e.sendRequest(req)
}

func (e *engageHTTP) sendRequest(req *http.Request) error {
	e.log.Debug("sending request")
	resp, err := e.client.Do(req)
	if err != nil {
		if e.isNormal(err) {
			return err
		}

		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("invalid external code: %s", resp.Status)
	}

	return nil
}

func (e *engageHTTP) createSendRequest(ctx context.Context, chatID int64, msg string) (*http.Request, error) {
	e.log.Debug("marshalling body of sending request", slog.Int64("chat_id", chatID))
	reqBody, err := e.getSendingBody(chatID, msg)
	if err != nil {
		return nil, err
	}

	e.log.Debug("creating request of sending")
	req, err := http.NewRequestWithContext(ctx, "POST", e.getSendingUrl(), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

func (e *engageHTTP) getSendingBody(chatID int64, msg string) ([]byte, error) {
	bytes, err := json.Marshal(&SendMessageRequest{
		ChatID: chatID,
		Text:   msg,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sending body: %w", err)
	}

	return bytes, nil
}

func (e *engageHTTP) getSendingUrl() string {
	return fmt.Sprintf("%s/api/v1/telegram/message/send", e.cfg.Engage.BaseURL)
}

func (e *engageHTTP) isNormal(err error) bool {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return true
	}

	return false
}
