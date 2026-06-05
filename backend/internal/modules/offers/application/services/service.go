package offerservices

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/devathh/hive/backend/internal/common/auth"
	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/modules/offers/domain/offer"
	"github.com/devathh/hive/backend/pkg/errs"
	"github.com/google/uuid"
)

type OfferService struct {
	cfg              *config.Config
	log              *slog.Logger
	offerPersistence offer.OfferPersistenceRepository
	offerCache       offer.OfferCacheRepository
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	offerPersistence offer.OfferPersistenceRepository,
	offerCache offer.OfferCacheRepository,
) *OfferService {
	return &OfferService{
		cfg:              cfg,
		log:              log,
		offerPersistence: offerPersistence,
		offerCache:       offerCache,
	}
}

// Create is a service's method
// that saves into db a new offer
func (os *OfferService) Create(ctx context.Context, req *requests.CreateOffer) (*responses.Offer, error) {
	userID, err := os.getUserID(ctx)
	if err != nil {
		return nil, errs.ErrInvalidToken
	}

	offer, err := offer.New(
		userID,
		req.Title,
		req.Description,
		req.Price,
		offer.SlotDuration(time.Duration(req.SlotDurationMinutes)*time.Minute),
		offer.SlotInterval(time.Duration(req.SlotIntervalMinutes)*time.Minute),
	)
	if err != nil {
		return nil, err
	}

	savedOffer, err := os.offerPersistence.Save(ctx, offer)
	if err != nil {
		safeErr, ok := os.handleRepoError(err)
		if !ok {
			os.log.Error("failed to save offer into db",
				slog.String("error", err.Error()),
				slog.String("offer_id", offer.ID().String()),
				slog.String("user_id", userID.String()))
		}

		return nil, safeErr
	}

	return os.offerToDTO(savedOffer), nil
}

// Update is a service's method
// that updates in db the fields of offer
func (os *OfferService) Update(ctx context.Context, req *requests.UpdateOffer) (*responses.Offer, error) {
	userID, err := os.getUserID(ctx)
	if err != nil {
		return nil, errs.ErrInvalidToken
	}

	updOffer, err := os.offerPersistence.GetByID(ctx, req.OfferID)
	if err != nil {
		safeErr, ok := os.handleRepoError(err)
		if !ok {
			os.log.Error("failed to get offer by id",
				slog.String("user_id", userID.String()),
				slog.String("offer_id", req.OfferID.String()),
				slog.String("error", err.Error()))
		}

		return nil, safeErr
	}
	if updOffer.AuthorID() != userID {
		return nil, errs.ErrNotEnoughRights
	}

	if err := os.applyChanges(updOffer, req); err != nil {
		return nil, err
	}

	updatedOffer, err := os.offerPersistence.Update(ctx, updOffer)
	if err != nil {
		safeErr, ok := os.handleRepoError(err)
		if !ok {
			os.log.Error("failed to update offer in db",
				slog.String("error", err.Error()),
				slog.String("user_id", userID.String()))
		}

		return nil, safeErr
	}

	return os.offerToDTO(updatedOffer), nil
}

// Delete is a service's method
// that deletes from db the offer
func (os *OfferService) Delete(ctx context.Context, id uuid.UUID) error {
	userID, err := os.getUserID(ctx)
	if err != nil {
		return errs.ErrInvalidToken
	}

	offer, err := os.offerPersistence.GetByID(ctx, id)
	if err != nil {
		safeErr, ok := os.handleRepoError(err)
		if !ok {
			os.log.Error("failed to get offer by id for delete",
				slog.String("offer_id", id.String()),
				slog.String("user_id", userID.String()),
				slog.String("error", err.Error()))
		}

		return safeErr
	}

	if offer.AuthorID() != userID {
		return errs.ErrNotEnoughRights
	}

	if err := os.offerPersistence.Delete(ctx, offer.ID()); err != nil {
		safeErr, ok := os.handleRepoError(err)
		if !ok {
			os.log.Error("failed to delete offer",
				slog.String("offer_id", offer.ID().String()),
				slog.String("user_id", userID.String()),
				slog.String("error", err.Error()))
		}

		return safeErr
	}

	return nil
}

// GetByID is a service's method
// that returns the offer by its id
func (os *OfferService) GetByID(ctx context.Context, id uuid.UUID) (*responses.Offer, error) {
	if cachedOffer, err := os.offerCache.Get(ctx, id); err == nil {
		return os.offerToDTO(cachedOffer), nil
	} else if !errors.Is(err, offer.ErrOfferNotFound) {
		os.log.Warn("failed to get offer from cache",
			slog.String("offer_id", id.String()),
			slog.String("error", err.Error()))
	}

	offer, err := os.offerPersistence.GetByID(ctx, id)
	if err != nil {
		safeErr, ok := os.handleRepoError(err)
		if !ok {
			os.log.Error("failed to get offer by id",
				slog.String("offer_id", id.String()),
				slog.String("error", err.Error()))
		}

		return nil, safeErr
	}

	if err := os.offerCache.Set(ctx, offer); err != nil {
		os.log.Warn("failed to set offer into cache",
			slog.String("error", err.Error()))
	}

	return os.offerToDTO(offer), nil
}

// TODO: add pagination
// ListByAuthorID is a service's method
// that returns a list of the offers by author's id
func (os *OfferService) ListByAuthorID(ctx context.Context, id uuid.UUID) (*responses.Offers, error) {
	list, err := os.offerPersistence.ListByAuthorID(ctx, id)
	if err != nil {
		safeErr, ok := os.handleRepoError(err)
		if !ok {
			os.log.Error("failed to list offers by author id",
				slog.String("error", err.Error()),
				slog.String("author_id", id.String()))
		}

		return nil, safeErr
	}

	response := responses.Offers{
		List: make([]*responses.Offer, len(list)),
	}
	for idx, offer := range list {
		response.List[idx] = os.offerToDTO(offer)
	}

	return &response, nil
}

func (os *OfferService) applyChanges(updOffer *offer.Offer, req *requests.UpdateOffer) error {
	if req.Title != nil {
		if err := updOffer.ChangeTitle(*req.Title); err != nil {
			return err
		}
	}
	if req.Description != nil {
		if err := updOffer.ChangeDescription(*req.Description); err != nil {
			return err
		}
	}
	if req.Price != nil {
		updOffer.ChangePrice(uint64(*req.Price))
	}
	if req.SlotDurationMinutes != nil {
		slotDuration := offer.SlotDuration(time.Duration(*req.SlotDurationMinutes) * time.Minute)
		if err := updOffer.ChangeSlotDuration(slotDuration); err != nil {
			return err
		}
	}
	if req.SlotIntervalMinutes != nil {
		slotInterval := offer.SlotInterval(time.Duration(*req.SlotIntervalMinutes) * time.Minute)
		if err := updOffer.ChangeSlotInterval(slotInterval); err != nil {
			return err
		}
	}
	if req.IsActive != nil {
		if *req.IsActive {
			return updOffer.Enable()
		}

		return updOffer.Disable()
	}

	return nil
}

func (os *OfferService) offerToDTO(offer *offer.Offer) *responses.Offer {
	return &responses.Offer{
		ID:                  offer.ID().String(),
		AuthorID:            offer.AuthorID().String(),
		Title:               offer.Title(),
		Description:         offer.Description(),
		Price:               offer.Price(),
		SlotDurationMinutes: uint32(offer.SlotDuration().Minutes()),
		SlotIntervalMinutes: uint32(offer.SlotInterval().Minutes()),
		IsActive:            offer.IsActive().Bool(),
		CreatedAt:           offer.CreatedAt().Unix(),
		UpdatedAt:           offer.UpdatedAt().Unix(),
	}
}

func (os *OfferService) handleRepoError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return err, true
	case errors.Is(err, offer.ErrOfferNotFound):
		return err, true
	}

	return err, false
}

func (os *OfferService) getUserID(ctx context.Context) (uuid.UUID, error) {
	raw := ctx.Value(auth.KeyUserID)
	if raw == nil {
		return uuid.Nil, errs.ErrInvalidToken
	}

	if id, ok := raw.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, errs.ErrInvalidToken
}
