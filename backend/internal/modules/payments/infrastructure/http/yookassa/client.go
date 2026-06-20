package yookassa

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/modules/payments/domain/payment"
	"github.com/google/uuid"
)

var (
	ErrInvalidResponseCode = errors.New("invalid response code")
)

type yookassaClient struct {
	cfg    *config.Config
	log    *slog.Logger
	client *http.Client
}

func New(cfg *config.Config, log *slog.Logger, client *http.Client) *yookassaClient {
	return &yookassaClient{
		cfg:    cfg,
		log:    log,
		client: client,
	}
}

func (yc *yookassaClient) CreatePayment(ctx context.Context, userID uuid.UUID, paidDays uint64, amountPenny uint64) (*payment.Payment, error) {
	req, err := yc.buildCreateRequest(ctx, yc.createRequestBody(userID, paidDays, amountPenny))
	if err != nil {
		return nil, err
	}

	yc.log.Debug("do http request to create payment", slog.String("user_id", userID.String()))
	resp, err := yc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	response, err := yc.parseCreateResponse(resp)
	if err != nil {
		return nil, err
	}

	return yc.mapResponse(response, amountPenny, userID)
}

func (yc *yookassaClient) GetStatus(ctx context.Context, paymentID uuid.UUID) (string, error) {
	req, err := yc.buildGetRequest(ctx, paymentID)
	if err != nil {
		return "", err
	}

	resp, err := yc.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	return yc.parseGetResponse(resp)
}

func (yc *yookassaClient) checkStatus(code int) error {
	if code > 299 || code < 200 {
		return fmt.Errorf("invalid code from yookassa: %d", code)
	}

	return nil
}

func (yc *yookassaClient) parseGetResponse(resp *http.Response) (string, error) {
	if err := yc.checkStatus(resp.StatusCode); err != nil {
		return "", err
	}

	var response GetPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Status, nil
}

func (yc *yookassaClient) parseCreateResponse(resp *http.Response) (*CreatePaymentResponse, error) {
	if err := yc.checkStatus(resp.StatusCode); err != nil {
		return nil, err
	}

	var response CreatePaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func (yc *yookassaClient) buildGetRequest(ctx context.Context, paymentID uuid.UUID) (*http.Request, error) {
	url := yc.cfg.Yookassa.BaseURL + "/v3/payments/" + paymentID.String()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(yc.cfg.Yookassa.Auth.ShopID,
		yc.cfg.Yookassa.Auth.SecretKey)

	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

func (yc *yookassaClient) buildCreateRequest(ctx context.Context, raw *CreatePaymentRequest) (*http.Request, error) {
	requestBody, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create request: %w", err)
	}

	url := yc.cfg.Yookassa.BaseURL + "/v3/payments"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(yc.cfg.Yookassa.Auth.ShopID,
		yc.cfg.Yookassa.Auth.SecretKey)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Idempotence-Key", yc.getIdempotencyKey())

	return req, nil
}

func (yc *yookassaClient) mapResponse(response *CreatePaymentResponse, amount uint64, userID uuid.UUID) (*payment.Payment, error) {
	return payment.New(
		response.ID,
		userID,
		amount,
	)
}

func (yc *yookassaClient) createRequestBody(userID uuid.UUID, paidDays uint64, amount uint64) *CreatePaymentRequest {
	return &CreatePaymentRequest{
		Amount: Amount{
			Value:    fmt.Sprintf("%.2f", float64(amount)/100.0),
			Currency: yc.cfg.Yookassa.Creates.Currency,
		},
		Capture: yc.cfg.Yookassa.Creates.Capture,
		Confirmation: Confirmation{
			Type:      yc.cfg.Yookassa.Creates.Confirmation.Type,
			ReturnURL: yc.cfg.Yookassa.Creates.Confirmation.ReturnURL,
		},
		Metadata: Metadata{
			UserID:   userID.String(),
			PaidDays: int(paidDays),
		},
		Description: fmt.Sprintf("user:%s", userID.String()),
	}
}

func (yc *yookassaClient) getIdempotencyKey() string {
	return fmt.Sprintf("%s", uuid.NewString())
}
