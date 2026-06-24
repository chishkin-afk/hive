package paymentservices

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/devathh/hive/backend/internal/common/auth"
	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/modules/payments/domain/payment"
	"github.com/devathh/hive/backend/internal/modules/payments/domain/rate"
	"github.com/devathh/hive/backend/pkg/errs"
	"github.com/google/uuid"
)

const (
	statusSucceeded = "succeeded"
	statusCanceled  = "canceled"
)

type YookassaClient interface {
	CreatePayment(ctx context.Context, userID, rateID uuid.UUID, amountPenny uint64) (*payment.Payment, string, error)
	GetStatus(ctx context.Context, paymentID uuid.UUID) (string, error)
}

type PaymentService struct {
	cfg                *config.Config
	log                *slog.Logger
	ratePersistence    rate.RatePersistenceRepository
	paymentPersistence payment.PaymentPersistenceRepository
	yookassaClient     YookassaClient
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	ratePersistence rate.RatePersistenceRepository,
	paymentPersistence payment.PaymentPersistenceRepository,
	yookassaClient YookassaClient,
) *PaymentService {
	return &PaymentService{
		cfg:                cfg,
		log:                log,
		ratePersistence:    ratePersistence,
		paymentPersistence: paymentPersistence,
		yookassaClient:     yookassaClient,
	}
}

// Create creates an unconfirmed payment in the system (in the pending status)
// and waits for it to be confirmed by the cash register
func (ps *PaymentService) Create(ctx context.Context, req *requests.CreatePayment) (*responses.UnconfirmedPayment, error) {
	userID, err := ps.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	rate, err := ps.getRateFromPersistence(ctx, req.RateID)
	if err != nil {
		return nil, err
	}

	yookassaPayment, url, err := ps.createPayment(ctx, userID, rate)
	if err != nil {
		return nil, err
	}

	unconfirmedPayment, err := ps.storePaymentPersistence(ctx, yookassaPayment)
	if err != nil {
		return nil, err
	}

	return ps.unconfirmedPaymentToDTO(unconfirmedPayment, url), nil
}

// Confirm changes the status of an existing transaction
// to success and adds the subscription-paid days to the user
func (ps *PaymentService) Confirm(ctx context.Context, req *requests.ConfirmPayment) (*responses.Payment, error) {
	updatedPayment, err := ps.confirmPayment(ctx, req)
	if err != nil {
		return nil, err
	}

	return ps.paymentToDTO(updatedPayment), nil
}

// Cancel is a method that changes the status of an existing transaction to canceled
func (ps *PaymentService) Cancel(ctx context.Context, req *requests.CancelPayment) (*responses.Payment, error) {
	updatedPayment, err := ps.cancelPayment(ctx, req)
	if err != nil {
		return nil, err
	}

	return ps.paymentToDTO(updatedPayment), nil
}

// ListByUserID returns all existing (paginated) transactions to the user.
// The user is taken from the token
func (ps *PaymentService) ListByUserID(ctx context.Context, page, limit uint32) (*responses.Payments, error) {
	if err := ps.validatePagination(page, limit); err != nil {
		return nil, err
	}

	userID, err := ps.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	list, totalCount, err := ps.listPaymentsFromPersistence(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}

	return ps.listToDTO(list, totalCount, page, limit), nil
}

// TODO: add cache
// GetByID returns a specific transaction by its id.
// It also checks for user authorship
func (ps *PaymentService) GetByID(ctx context.Context, id uuid.UUID) (*responses.Payment, error) {
	userID, err := ps.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	payment, err := ps.getPaymentFromPersistence(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := ps.checkReadRights(userID, payment); err != nil {
		return nil, err
	}

	return ps.paymentToDTO(payment), nil
}

// TODO: add cache
// ListAllRates is a method that returns a list of all rates to the user.
// Tries to get them from the cache or the database
func (ps *PaymentService) ListAllRates(ctx context.Context) (*responses.Rates, error) {
	rates, err := ps.listAllRates(ctx)
	if err != nil {
		return nil, err
	}

	return ps.ratesToDTO(rates), nil
}

// TODO: add cache
// GetRateByID method that returns a specific rate by id
func (ps *PaymentService) GetRateByID(ctx context.Context, id uuid.UUID) (*responses.Rate, error) {
	rate, err := ps.getRateFromPersistence(ctx, id)
	if err != nil {
		return nil, err
	}

	return ps.rateToDTO(rate), nil
}

// TakeTrial is a method that add trial days to user (if trial used is false)
func (ps *PaymentService) TakeTrial(ctx context.Context) error {
	userID, err := ps.getUserID(ctx)
	if err != nil {
		return err
	}

	return ps.takeTrial(ctx, userID)
}

func (ps *PaymentService) getUserID(ctx context.Context) (uuid.UUID, error) {
	raw := ctx.Value(auth.KeyUserID)
	if raw == nil {
		return uuid.Nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
	}

	if id, ok := raw.(uuid.UUID); ok {
		if id == uuid.Nil {
			return uuid.Nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
		}

		return id, nil
	}

	return uuid.Nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
}

func (ps *PaymentService) getRateFromPersistence(ctx context.Context, rateID uuid.UUID) (*rate.Rate, error) {
	rate, err := ps.ratePersistence.GetByID(ctx, rateID)
	if err != nil {
		ps.log.Error("failed to get rate from persistence",
			slog.String("error", err.Error()),
			slog.String("rate_id", rateID.String()),
		)

		return nil, ps.handleRepoError(err)
	}

	ps.log.Debug("rate has been taken from db", slog.String("rate_name", rate.Name()))
	return rate, nil
}

func (ps *PaymentService) createPayment(ctx context.Context, userID uuid.UUID, rate *rate.Rate) (*payment.Payment, string, error) {
	payment, url, err := ps.yookassaClient.CreatePayment(ctx, userID, rate.ID(), rate.AmountPenny())
	if err != nil {
		ps.log.Error("failed to create payment on yookassa",
			slog.String("error", err.Error()),
			slog.String("rate_id", rate.ID().String()),
			slog.String("user_id", userID.String()),
		)

		return nil, "", ps.handleRepoError(err)
	}

	return payment, url, nil
}

func (ps *PaymentService) storePaymentPersistence(ctx context.Context, payment *payment.Payment) (*payment.Payment, error) {
	savedPayment, err := ps.paymentPersistence.Save(ctx, payment)
	if err != nil {
		ps.log.Error("failed to save payment into db",
			slog.String("error", err.Error()),
			slog.String("payment_id", payment.ID().String()),
		)

		return nil, ps.handleRepoError(err)
	}

	return savedPayment, nil
}

func (ps *PaymentService) unconfirmedPaymentToDTO(payment *payment.Payment, url string) *responses.UnconfirmedPayment {
	return &responses.UnconfirmedPayment{
		Payment: ps.paymentToDTO(payment),
		URL:     url,
	}
}

func (ps *PaymentService) confirmPayment(ctx context.Context, req *requests.ConfirmPayment) (*payment.Payment, error) {
	updFunc := func(ctx context.Context, paidUntil time.Time, payment *payment.Payment) (time.Time, error) {
		status, err := ps.getYookassaStatus(ctx, payment.PaymentID())
		if err != nil {
			return time.Time{}, err
		}

		if status != statusSucceeded {
			return time.Time{}, errs.ErrInvalidPaymentStatus
		}

		rate, err := ps.getRateFromPersistence(ctx, req.RateID)
		if err != nil {
			return time.Time{}, err
		}

		newTime := ps.payUntil(paidUntil, rate.PaidDays())
		if err := payment.Confirm(); err != nil {
			return time.Time{}, err
		}

		return newTime, nil
	}

	return ps.updatePayment(ctx, updFunc, req.PaymentID, req.UserID)
}

func (ps *PaymentService) paymentToDTO(payment *payment.Payment) *responses.Payment {
	return &responses.Payment{
		ID:           payment.ID(),
		PaymentID:    payment.PaymentID(),
		UserID:       payment.UserID(),
		AmountPenny:  payment.AmountPenny(),
		StatusString: payment.Status().Value().String(),
		StatusInt:    payment.Status().Value().Int(),
		CreatedAt:    payment.CreatedAt().Unix(),
		UpdatedAt:    payment.UpdatedAt().Unix(),
	}
}

func (ps *PaymentService) cancelPayment(ctx context.Context, req *requests.CancelPayment) (*payment.Payment, error) {
	updFunc := func(ctx context.Context, paidUntil time.Time, payment *payment.Payment) (time.Time, error) {
		status, err := ps.getYookassaStatus(ctx, payment.PaymentID())
		if err != nil {
			return time.Time{}, err
		}

		if status != statusCanceled {
			return time.Time{}, errs.ErrInvalidPaymentStatus
		}

		if err := payment.Cancel(); err != nil {
			return time.Time{}, err
		}

		return paidUntil, nil
	}

	return ps.updatePayment(ctx, updFunc, req.PaymentID, req.UserID)
}

func (ps *PaymentService) validatePagination(page, limit uint32) error {
	if page < 1 || page > 1000 {
		return errs.NewErrorWithKind(errs.ErrInvalidPage, errs.KindRequest)
	}

	if limit < 1 || limit > 100 {
		return errs.NewErrorWithKind(errs.ErrInvalidLimit, errs.KindRequest)
	}

	return nil
}

func (ps *PaymentService) listPaymentsFromPersistence(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*payment.Payment, int64, error) {
	list, totalCount, err := ps.paymentPersistence.ListByUserID(ctx, userID, page, limit)
	if err != nil {
		ps.log.Error("failed to list payments from db",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
		)

		return nil, 0, ps.handleRepoError(err)
	}

	return list, totalCount, nil
}

func (ps *PaymentService) listToDTO(list []*payment.Payment, totalCount int64, page, limit uint32) *responses.Payments {
	totalPages := (totalCount + int64(limit) - 1) / int64(limit)
	response := responses.Payments{
		HasNextPage: page < uint32(totalPages),
		HasPrevPage: page > 1,
		TotalPages:  uint32(totalPages),
		List:        make([]*responses.Payment, len(list)),
	}

	for i, payment := range list {
		response.List[i] = ps.paymentToDTO(payment)
	}

	return &response
}

func (ps *PaymentService) getPaymentFromPersistence(ctx context.Context, id uuid.UUID) (*payment.Payment, error) {
	payment, err := ps.paymentPersistence.GetByID(ctx, id)
	if err != nil {
		ps.log.Error("failed to get payment from persistence",
			slog.String("error", err.Error()),
			slog.String("payment_id", id.String()),
		)

		return nil, ps.handleRepoError(err)
	}

	return payment, nil
}

func (ps *PaymentService) checkReadRights(userID uuid.UUID, payment *payment.Payment) error {
	if userID != payment.UserID() {
		return errs.NewErrorWithKind(errs.ErrNotEnoughRights, errs.KindPermissionDenied)
	}

	return nil
}

func (ps *PaymentService) listAllRates(ctx context.Context) ([]*rate.Rate, error) {
	rates, err := ps.ratePersistence.ListAll(ctx)
	if err != nil {
		ps.log.Error("failed to list all rates",
			slog.String("error", err.Error()),
		)

		return nil, err
	}

	return rates, nil
}

func (ps *PaymentService) ratesToDTO(rates []*rate.Rate) *responses.Rates {
	response := responses.Rates{
		List: make([]*responses.Rate, len(rates)),
	}
	for i, rate := range rates {
		response.List[i] = ps.rateToDTO(rate)
	}

	return &response
}

func (ps *PaymentService) rateToDTO(rate *rate.Rate) *responses.Rate {
	return &responses.Rate{
		ID:          rate.ID(),
		Name:        rate.Name(),
		AmountPenny: rate.AmountPenny(),
		PaidDays:    rate.PaidDays(),
	}
}

func (ps *PaymentService) takeTrial(ctx context.Context, userID uuid.UUID) error {
	updFunc := func(ctx context.Context, paidUntil time.Time, trialUsed bool) (time.Time, error) {
		if trialUsed {
			return time.Time{}, payment.ErrTrialAlreadyUsed
		}

		return ps.payUntil(paidUntil, uint32(ps.cfg.Service.TrialPeriod)), nil
	}

	if err := ps.paymentPersistence.TakeTrial(ctx, userID, updFunc); err != nil {
		ps.log.Error("failed to take trial",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
		)

		return ps.handleRepoError(err)
	}

	return nil
}

func (ps *PaymentService) updatePayment(ctx context.Context, updFunc payment.UpdateFunc, paymentID, userID uuid.UUID) (*payment.Payment, error) {
	updatedPayment, err := ps.paymentPersistence.Update(ctx, paymentID, userID, updFunc)
	if err != nil {
		ps.log.Error("failed to update payment in db",
			slog.String("error", err.Error()),
			slog.String("payment_id", paymentID.String()),
		)

		return nil, ps.handleRepoError(err)
	}

	return updatedPayment, nil
}

func (ps *PaymentService) payUntil(curr time.Time, paidDays uint32) time.Time {
	now := time.Now().UTC()
	if now.Compare(curr) > 0 {
		return now.Add(time.Duration(paidDays) * 24 * time.Hour)
	}

	return curr.Add(time.Duration(paidDays) * 24 * time.Hour)
}

func (ps *PaymentService) getYookassaStatus(ctx context.Context, paymentID uuid.UUID) (string, error) {
	status, err := ps.yookassaClient.GetStatus(ctx, paymentID)
	if err != nil {
		return "", err
	}

	return status, nil
}

func (ps *PaymentService) handleRepoError(err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return errs.NewErrorWithKind(err, errs.KindRequest)
	case errors.Is(err, rate.ErrRateNotFound),
		errors.Is(err, payment.ErrPaymentNotFound):
		return errs.NewErrorWithKind(err, errs.KindNotFound)
	case errors.Is(err, payment.ErrTrialAlreadyUsed),
		errors.Is(err, payment.ErrAlreadyCancelled),
		errors.Is(err, payment.ErrAlreadyConfirmed):
		return errs.NewErrorWithKind(err, errs.KindConflict)
	case errors.Is(err, errs.ErrInvalidPaymentStatus):
		return errs.NewErrorWithKind(err, errs.KindPermissionDenied)
	}

	return errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal)
}
