package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/braunkc/hive-engage/config"
	"github.com/braunkc/hive-engage/internal/infra/http/dtos"
)

var (
	ErrTgBindingsNotFound = errors.New("tg bindings not found")
)

type HTTPClient struct {
	client  *http.Client
	baseURL string
	log     *slog.Logger
}

func New(cfg config.HTTPAuthConfig, log *slog.Logger) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		baseURL: fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port),
		log:     log,
	}
}

func (h *HTTPClient) BindTg(ctx context.Context, req *dtos.BindTgRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	reqHTTP, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/auth/tgbinding", h.baseURL),
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	reqHTTP.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(reqHTTP)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrTgBindingsNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to bind tg with status: %d", resp.StatusCode)
	}

	return nil
}
