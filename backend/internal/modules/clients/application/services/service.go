package clientservices

import (
	"context"
	"errors"
	"log/slog"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/devathh/hive/backend/internal/common/auth"
	"github.com/devathh/hive/backend/internal/modules/clients/domain/client"
	"github.com/devathh/hive/backend/pkg/errs"
	"github.com/google/uuid"
)

type ClientService struct {
	log               *slog.Logger
	clientPersistence client.ClientPersistenceRepository
	clientCache       client.ClientCacheRepository
}

func New(
	log *slog.Logger,
	clientPersistence client.ClientPersistenceRepository,
	clientCache client.ClientCacheRepository,
) *ClientService {
	return &ClientService{
		log:               log,
		clientPersistence: clientPersistence,
		clientCache:       clientCache,
	}
}

// Upsert is a method that
// creates client if he doesn't exist or
// updates client if he exists
func (cs *ClientService) Upsert(ctx context.Context, req *requests.UpsertClient) (*responses.Client, error) {
	userID, err := cs.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	updClient, notFound, err := cs.getClientByPairIDs(ctx, userID, req.OfferID)
	if err != nil && !notFound {
		return nil, err
	}

	if !notFound {
		return cs.updateClient(ctx, updClient)
	}

	return cs.createClient(ctx, userID, req.OfferID)
}

// GetByID is a method that
// returns client from db / cache
func (cs *ClientService) GetByID(ctx context.Context, id uuid.UUID) (*responses.Client, error) {
	if cachedClient, ok := cs.getClientFromCache(ctx, id); ok {
		return cs.clientToDTO(cachedClient), nil
	}

	client, err := cs.getClientFromPersistence(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := cs.storeClientCache(ctx, client); err != nil {
		cs.log.Warn("failed to set client's cache", slog.String("error", err.Error()))
	}

	return cs.clientToDTO(client), nil
}

// ListByOfferID is a method that
// returns a list of clients by offer's id
func (cs *ClientService) ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) (*responses.Clients, error) {
	if err := cs.validatePagination(page, limit); err != nil {
		return nil, err
	}

	list, totalCount, err := cs.listClientsByOfferID(ctx, offerID, page, limit)
	if err != nil {
		return nil, err
	}

	return cs.listToDTO(list, totalCount, page, limit), nil
}

// ListByUserID is a method that
// returns a list of clients by user's id
func (cs *ClientService) ListByUserID(ctx context.Context, page, limit uint32) (*responses.Clients, error) {
	userID, err := cs.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := cs.validatePagination(page, limit); err != nil {
		return nil, err
	}

	list, totalCount, err := cs.listClientsByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}

	return cs.listToDTO(list, totalCount, page, limit), nil
}

func (cs *ClientService) getUserID(ctx context.Context) (uuid.UUID, error) {
	raw := ctx.Value(auth.KeyUserID)
	if raw == nil {
		return uuid.Nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
	}

	if id, ok := raw.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
}

func (cs *ClientService) getClientByPairIDs(ctx context.Context, userID, offerID uuid.UUID) (*client.Client, bool, error) {
	foundClient, err := cs.clientPersistence.GetByPairIDs(ctx, userID, offerID)
	if err != nil {
		if errors.Is(err, client.ErrClientNotFound) {
			return nil, true, nil
		}

		cs.log.Error("failed to get client by pair of ids",
			slog.String("user_id", userID.String()),
			slog.String("offer_id", offerID.String()),
			slog.String("error", err.Error()),
		)

		return nil, false, cs.handleRepoError(err)
	}

	return foundClient, false, nil
}

func (cs *ClientService) updateClient(ctx context.Context, updClient *client.Client) (*responses.Client, error) {
	updatedClient, err := cs.clientPersistence.Update(ctx, func(c *client.Client) error {
		c.AddVisit()
		return nil
	}, updClient.ID())
	if err != nil {
		cs.log.Error("failed to update client in db",
			slog.String("client_id", updClient.ID().String()),
			slog.String("error", err.Error()),
		)

		return nil, cs.handleRepoError(err)
	}

	if err := cs.clientCache.Delete(ctx, updatedClient.ID()); err != nil {
		cs.log.Warn("failed to delete client from cache",
			slog.String("error", err.Error()),
			slog.String("client_id", updatedClient.ID().String()),
		)
	}

	return cs.clientToDTO(updatedClient), nil
}

func (cs *ClientService) createClient(ctx context.Context, userID, offerID uuid.UUID) (*responses.Client, error) {
	client, err := client.New(userID, offerID)
	if err != nil {
		return nil, errs.NewErrorWithKind(err, errs.KindRequest)
	}

	savedClient, err := cs.clientPersistence.Save(ctx, client)
	if err != nil {
		cs.log.Error("failed to save client into db",
			slog.String("user_id", userID.String()),
			slog.String("offer_id", offerID.String()),
		)

		return nil, cs.handleRepoError(err)
	}

	return cs.clientToDTO(savedClient), nil
}

func (cs *ClientService) getClientFromCache(ctx context.Context, id uuid.UUID) (*client.Client, bool) {
	cachedClient, err := cs.clientCache.GetByID(ctx, id)
	if err != nil {
		if !errors.Is(err, client.ErrClientNotFound) {
			cs.log.Warn("failed to get client from cache",
				slog.String("client_id", id.String()),
				slog.String("error", err.Error()),
			)
		}

		return nil, false
	}

	cs.log.Debug("client has been taken from cache", slog.String("client_id", id.String()))
	return cachedClient, true
}

func (cs *ClientService) clientToDTO(client *client.Client) *responses.Client {
	return &responses.Client{
		ID:        client.ID(),
		UserID:    client.UserID(),
		OfferID:   client.OfferID(),
		Visits:    client.Visits(),
		LastVisit: client.LastVisit().Unix(),
		CreatedAt: client.CreatedAt().Unix(),
	}
}

func (cs *ClientService) getClientFromPersistence(ctx context.Context, id uuid.UUID) (*client.Client, error) {
	client, err := cs.clientPersistence.GetByID(ctx, id)
	if err != nil {
		cs.log.Error("failed to get client from persistence by id",
			slog.String("client_id", id.String()),
			slog.String("error", err.Error()),
		)

		return nil, cs.handleRepoError(err)
	}

	return client, nil
}

func (cs *ClientService) storeClientCache(ctx context.Context, client *client.Client) error {
	return cs.clientCache.Save(ctx, client)
}

func (cs *ClientService) validatePagination(page, limit uint32) error {
	if page < 1 || page > 1000 {
		return errs.NewErrorWithKind(errs.ErrInvalidPage, errs.KindRequest)
	}

	if limit < 1 || limit > 100 {
		return errs.NewErrorWithKind(errs.ErrInvalidLimit, errs.KindRequest)
	}

	return nil
}

func (cs *ClientService) listClientsByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) ([]*client.Client, int64, error) {
	list, totalCount, err := cs.clientPersistence.ListByOfferID(ctx, offerID, page, limit)
	if err != nil {
		cs.log.Error("failed to list clients by offer id",
			slog.String("offer_id", offerID.String()),
			slog.String("error", err.Error()),
		)

		return nil, 0, cs.handleRepoError(err)
	}

	return list, totalCount, nil
}

func (cs *ClientService) listToDTO(list []*client.Client, totalCount int64, page, limit uint32) *responses.Clients {
	totalPages := (totalCount + int64(limit) - 1) / int64(limit)
	response := &responses.Clients{
		List:        make([]*responses.Client, len(list)),
		TotalPages:  uint32(totalPages),
		HasNextPage: totalPages > int64(page),
		HasPrevPage: page > 1,
	}
	for idx, client := range list {
		response.List[idx] = cs.clientToDTO(client)
	}

	return response
}

func (cs *ClientService) listClientsByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*client.Client, int64, error) {
	list, totalCount, err := cs.clientPersistence.ListByUserID(ctx, userID, page, limit)
	if err != nil {
		cs.log.Error("failed to list clients by user id",
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)

		return nil, 0, cs.handleRepoError(err)
	}

	return list, totalCount, nil
}

func (cs *ClientService) handleRepoError(err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return errs.NewErrorWithKind(err, errs.KindTimeout)
	case errors.Is(err, client.ErrClientNotFound),
		errors.Is(err, client.ErrOfferNotFound):
		return errs.NewErrorWithKind(err, errs.KindNotFound)
	case errors.Is(err, client.ErrClientAlreadyExists):
		return errs.NewErrorWithKind(err, errs.KindConflict)
	}

	return errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal)
}
