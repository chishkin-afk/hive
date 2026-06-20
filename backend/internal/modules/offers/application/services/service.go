package offerservices

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/devathh/hive/backend/internal/common/auth"
	"github.com/devathh/hive/backend/internal/modules/offers/domain/offer"
	"github.com/devathh/hive/backend/pkg/errs"
	"github.com/google/uuid"
)

type OfferService struct {
	log              *slog.Logger
	offerPersistence offer.OfferPersistenceRepository
	offerCache       offer.OfferCacheRepository
}

func New(
	log *slog.Logger,
	offerPersistence offer.OfferPersistenceRepository,
	offerCache offer.OfferCacheRepository,
) *OfferService {
	return &OfferService{
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
		return nil, err
	}

	offer, err := os.buildNewOffer(req, userID)
	if err != nil {
		return nil, err
	}

	savedOffer, err := os.storeOfferPersistence(ctx, offer)
	if err != nil {
		return nil, err
	}

	return os.offerToDTO(savedOffer), nil
}

// Update is a service's method
// that updates in db the fields of offer
func (os *OfferService) Update(ctx context.Context, req *requests.UpdateOffer) (*responses.Offer, error) {
	userID, err := os.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	updOffer, err := os.getOfferFromPersistence(ctx, req.OfferID)
	if err != nil {
		return nil, err
	}

	if err := os.checkRights(userID, updOffer); err != nil {
		return nil, err
	}

	if err := os.applyChanges(updOffer, req); err != nil {
		return nil, err
	}

	updatedOffer, err := os.updateOfferPersistence(ctx, updOffer)
	if err != nil {
		return nil, err
	}

	return os.offerToDTO(updatedOffer), nil
}

// Delete is a service's method
// that deletes from db the offer
func (os *OfferService) Delete(ctx context.Context, id uuid.UUID) error {
	userID, err := os.getUserID(ctx)
	if err != nil {
		return err
	}

	offer, err := os.getOfferFromPersistence(ctx, id)
	if err != nil {
		return err
	}

	if err := os.checkRights(userID, offer); err != nil {
		return err
	}

	if err := os.deleteOfferPersistence(ctx, offer.ID()); err != nil {
		return err
	}

	return nil
}

// GetByID is a service's method
// that returns the offer by its id
func (os *OfferService) GetByID(ctx context.Context, id uuid.UUID) (*responses.Offer, error) {
	if cachedOffer, ok := os.getOfferFromCache(ctx, id); ok {
		return os.offerToDTO(cachedOffer), nil
	}

	offer, err := os.getOfferFromPersistence(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := os.storeOfferCache(ctx, offer); err != nil {
		os.log.Warn("failed to store offer's cache", slog.String("error", err.Error()))
	}

	return os.offerToDTO(offer), nil
}

// ListByAuthorID is a service's method
// that returns a list of the offers by author's id
func (os *OfferService) ListByAuthorID(ctx context.Context, id uuid.UUID, page, limit uint32) (*responses.Offers, error) {
	if err := os.validatePagination(page, limit); err != nil {
		return nil, err
	}

	list, totalCount, err := os.listOfferByAuthorID(ctx, id, page, limit)
	if err != nil {
		return nil, err
	}

	return os.listToDTO(list, totalCount, page, limit), nil
}

func (os *OfferService) getUserID(ctx context.Context) (uuid.UUID, error) {
	raw := ctx.Value(auth.KeyUserID)
	if raw == nil {
		return uuid.Nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
	}

	if id, ok := raw.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
}

func (os *OfferService) buildNewOffer(req *requests.CreateOffer, userID uuid.UUID) (*offer.Offer, error) {
	offer, err := offer.New(
		userID,
		req.Title,
		req.Description,
		req.Price,
		offer.SlotDuration(time.Duration(req.SlotDurationMinutes)*time.Minute),
		offer.SlotInterval(time.Duration(req.SlotIntervalMinutes)*time.Minute),
	)
	if err != nil {
		return nil, errs.NewErrorWithKind(err, errs.KindRequest)
	}

	return offer, nil
}

func (os *OfferService) storeOfferPersistence(ctx context.Context, offer *offer.Offer) (*offer.Offer, error) {
	savedOffer, err := os.offerPersistence.Save(ctx, offer)
	if err != nil {
		os.log.Error("failed to save offer into db",
			slog.String("error", err.Error()),
			slog.String("offer_id", offer.ID().String()),
		)

		return nil, os.handleRepoError(err)
	}

	return savedOffer, nil
}

func (os *OfferService) offerToDTO(offer *offer.Offer) *responses.Offer {
	return &responses.Offer{
		ID:                  offer.ID(),
		AuthorID:            offer.AuthorID(),
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

func (os *OfferService) getOfferFromPersistence(ctx context.Context, id uuid.UUID) (*offer.Offer, error) {
	updOffer, err := os.offerPersistence.GetByID(ctx, id)
	if err != nil {
		os.log.Error("failed to get offer from persistence",
			slog.String("error", err.Error()),
		)

		return nil, os.handleRepoError(err)
	}

	return updOffer, nil
}

func (os *OfferService) checkRights(userID uuid.UUID, offer *offer.Offer) error {
	if offer.AuthorID() != userID {
		return errs.NewErrorWithKind(errs.ErrNotEnoughRights, errs.KindPermissionDenied)
	}

	return nil
}

func (os *OfferService) applyChanges(updOffer *offer.Offer, req *requests.UpdateOffer) error {
	if req.Title != nil {
		if err := updOffer.ChangeTitle(*req.Title); err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}
	}
	if req.Description != nil {
		if err := updOffer.ChangeDescription(*req.Description); err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}
	}
	if req.Price != nil {
		updOffer.ChangePrice(uint64(*req.Price))
	}
	if req.SlotDurationMinutes != nil {
		slotDuration := offer.SlotDuration(time.Duration(*req.SlotDurationMinutes) * time.Minute)
		if err := updOffer.ChangeSlotDuration(slotDuration); err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}
	}
	if req.SlotIntervalMinutes != nil {
		slotInterval := offer.SlotInterval(time.Duration(*req.SlotIntervalMinutes) * time.Minute)
		if err := updOffer.ChangeSlotInterval(slotInterval); err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}
	}
	if req.IsActive != nil {
		if *req.IsActive {
			if err := updOffer.Enable(); err != nil {
				return errs.NewErrorWithKind(err, errs.KindRequest)
			}
		} else {
			if err := updOffer.Disable(); err != nil {
				return errs.NewErrorWithKind(err, errs.KindRequest)
			}
		}

	}

	return nil
}

func (os *OfferService) updateOfferPersistence(ctx context.Context, updOffer *offer.Offer) (*offer.Offer, error) {
	updatedOffer, err := os.offerPersistence.Update(ctx, updOffer)
	if err != nil {
		os.log.Error("failed to update persistence offer",
			slog.String("error", err.Error()),
			slog.String("offer_id", updOffer.ID().String()),
		)

		return nil, os.handleRepoError(err)
	}

	return updatedOffer, nil
}

func (os *OfferService) deleteOfferPersistence(ctx context.Context, id uuid.UUID) error {
	if err := os.offerPersistence.Delete(ctx, id); err != nil {
		os.log.Error("failed to delete offer from persistence",
			slog.String("error", err.Error()),
			slog.String("offer_id", id.String()),
		)

		return os.handleRepoError(err)
	}

	return nil
}

func (os *OfferService) getOfferFromCache(ctx context.Context, id uuid.UUID) (*offer.Offer, bool) {
	cachedOffer, err := os.offerCache.Get(ctx, id)
	if err != nil {
		if !errors.Is(err, offer.ErrOfferNotFound) {
			os.log.Error("failed to get offer from cache",
				slog.String("error", err.Error()),
				slog.String("offer_id", id.String()),
			)
		}

		return nil, false
	}

	os.log.Debug("offer has been taken from cache", slog.String("offer_id", id.String()))
	return cachedOffer, true
}

func (os *OfferService) storeOfferCache(ctx context.Context, offer *offer.Offer) error {
	return os.offerCache.Set(ctx, offer)
}

func (os *OfferService) validatePagination(page, limit uint32) error {
	if page < 1 || page > 1000 {
		return errs.NewErrorWithKind(errs.ErrInvalidPage, errs.KindRequest)
	}

	if limit < 1 || limit > 100 {
		return errs.NewErrorWithKind(errs.ErrInvalidLimit, errs.KindRequest)
	}

	return nil
}

func (os *OfferService) listOfferByAuthorID(ctx context.Context, id uuid.UUID, page, limit uint32) ([]*offer.Offer, int64, error) {
	list, totalCount, err := os.offerPersistence.ListByAuthorID(ctx, id, page, limit)
	if err != nil {
		os.log.Error("failed to list offers by author id",
			slog.String("author_id", id.String()),
			slog.String("error", err.Error()),
		)

		return nil, 0, os.handleRepoError(err)
	}

	return list, int64(totalCount), nil
}

func (os *OfferService) listToDTO(list []*offer.Offer, totalCount int64, page, limit uint32) *responses.Offers {
	totalPages := (totalCount + int64(limit) - 1) / int64(limit)
	response := responses.Offers{
		List:        make([]*responses.Offer, len(list)),
		TotalPages:  uint32(totalPages),
		HasNextPage: totalPages > int64(page),
		HasPrevPage: page > 1,
	}
	for idx, offer := range list {
		response.List[idx] = os.offerToDTO(offer)
	}

	return &response
}

func (os *OfferService) handleRepoError(err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return errs.NewErrorWithKind(err, errs.KindTimeout)
	case errors.Is(err, offer.ErrOfferNotFound):
		return errs.NewErrorWithKind(err, errs.KindNotFound)
	}

	return errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal)
}
