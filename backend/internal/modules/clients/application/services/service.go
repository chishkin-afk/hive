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
		return nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
	}

	updClient, err := cs.clientPersistence.GetByPairIDs(ctx, userID, req.OfferID)
	if errors.Is(err, client.ErrClientNotFound) {
		return cs.createClient(ctx, userID, req.OfferID)
	} else if err != nil {
		safeErr, ok := cs.handleRepoError(err)
		if !ok {
			cs.log.Error("failed to get client by pair of ids",
				slog.String("error", err.Error()),
				slog.String("offer_id", req.OfferID.String()),
				slog.String("user_id", userID.String()),
			)
		}

		return nil, safeErr
	}

	return cs.updateClient(ctx, updClient)
}

// GetByID is a method that
// returns client from db / cache
func (cs *ClientService) GetByID(ctx context.Context, id uuid.UUID) (*responses.Client, error) {
	if cachedClient, err := cs.clientCache.GetByID(ctx, id); err == nil {
		return cs.domainToDTO(cachedClient), nil
	} else if !errors.Is(err, client.ErrClientNotFound) {
		cs.log.Warn("failed to get client from cache by id",
			slog.String("error", err.Error()),
			slog.String("client_id", id.String()),
		)
	}

	client, err := cs.clientPersistence.GetByID(ctx, id)
	if err != nil {
		safeErr, ok := cs.handleRepoError(err)
		if !ok {
			cs.log.Error("failed to get client by id",
				slog.String("error", err.Error()),
				slog.String("client_id", id.String()),
			)
		}

		return nil, safeErr
	}

	if err := cs.clientCache.Save(ctx, client); err != nil {
		cs.log.Warn("failed to save client into cache",
			slog.String("error", err.Error()),
		)
	}
	return cs.domainToDTO(client), nil
}

// ListByOfferID is a method that
// returns a list of clients by offer's id
func (cs *ClientService) ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) (*responses.Clients, error) {
	if page < 1 || page > 1000 {
		return nil, errs.NewErrorWithKind(errs.ErrInvalidPage, errs.KindRequest)
	}

	if limit < 1 || limit > 100 {
		return nil, errs.NewErrorWithKind(errs.ErrInvalidLimit, errs.KindRequest)
	}

	list, totalCount, err := cs.clientPersistence.ListByOfferID(ctx, offerID, page, limit)
	if err != nil {
		safeErr, ok := cs.handleRepoError(err)
		if !ok {
			cs.log.Error("failed to get list of clients by offer id",
				slog.String("error", err.Error()),
				slog.String("offer_id", offerID.String()),
			)
		}

		return nil, safeErr
	}

	totalPages := totalCount / int64(limit)
	response := &responses.Clients{
		List:        make([]*responses.Client, len(list)),
		TotalPages:  uint32(totalPages),
		HasNextPage: totalPages > int64(page),
		HasPrevPage: page > 1,
	}
	for idx, client := range list {
		response.List[idx] = cs.domainToDTO(client)
	}

	return response, nil
}

// ListByUserID is a method that
// returns a list of clients by user's id
func (cs *ClientService) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) (*responses.Clients, error) {
	if page < 1 || page > 1000 {
		return nil, errs.NewErrorWithKind(errs.ErrInvalidPage, errs.KindRequest)
	}

	if limit < 1 || limit > 100 {
		return nil, errs.NewErrorWithKind(errs.ErrInvalidLimit, errs.KindRequest)
	}

	list, totalCount, err := cs.clientPersistence.ListByUserID(ctx, userID, page, limit)
	if err != nil {
		safeErr, ok := cs.handleRepoError(err)
		if !ok {
			cs.log.Error("failed to get list of clients by user id",
				slog.String("error", err.Error()),
				slog.String("user_id", userID.String()),
			)
		}

		return nil, safeErr
	}

	totalPages := totalCount / int64(limit)
	response := &responses.Clients{
		List:        make([]*responses.Client, len(list)),
		TotalPages:  uint32(totalPages),
		HasNextPage: totalPages > int64(page),
		HasPrevPage: page > 1,
	}
	for idx, client := range list {
		response.List[idx] = cs.domainToDTO(client)
	}

	return response, nil
}

func (cs *ClientService) updateClient(ctx context.Context, updClient *client.Client) (*responses.Client, error) {
	updatedClient, err := cs.clientPersistence.Update(ctx, func(c *client.Client) error {
		c.AddVisit()
		return nil
	}, updClient.ID())
	if err != nil {
		safeErr, ok := cs.handleRepoError(err)
		if !ok {
			cs.log.Error("failed to update client in db",
				slog.String("error", err.Error()),
				slog.String("client_id", updClient.ID().String()),
			)
		}

		return nil, safeErr
	}

	if err := cs.clientCache.Delete(ctx, updatedClient.ID()); err != nil {
		cs.log.Warn("failed to delete client from cache",
			slog.String("error", err.Error()),
			slog.String("client_id", updatedClient.ID().String()),
		)
	}

	return cs.domainToDTO(updatedClient), nil
}

func (cs *ClientService) createClient(ctx context.Context, userID, offerID uuid.UUID) (*responses.Client, error) {
	client, err := client.New(userID, offerID)
	if err != nil {
		return nil, errs.NewErrorWithKind(err, errs.KindRequest)
	}

	savedClient, err := cs.clientPersistence.Save(ctx, client)
	if err != nil {
		safeErr, ok := cs.handleRepoError(err)
		if !ok {
			cs.log.Error("failed to safe client into db",
				slog.String("error", err.Error()),
				slog.String("offer_id", offerID.String()),
				slog.String("user_id", userID.String()),
			)
		}

		return nil, safeErr
	}

	return cs.domainToDTO(savedClient), nil
}

func (cs *ClientService) domainToDTO(client *client.Client) *responses.Client {
	return &responses.Client{
		ID:        client.ID().String(),
		UserID:    client.UserID().String(),
		OfferID:   client.OfferID().String(),
		Visits:    client.Visits(),
		LastVisit: client.LastVisit().Unix(),
		CreatedAt: client.CreatedAt().Unix(),
	}
}

func (cs *ClientService) handleRepoError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return errs.NewErrorWithKind(err, errs.KindTimeout), true
	case errors.Is(err, client.ErrClientNotFound):
		return errs.NewErrorWithKind(err, errs.KindNotFound), true
	}

	return errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal), false
}

func (cs *ClientService) getUserID(ctx context.Context) (uuid.UUID, error) {
	raw := ctx.Value(auth.KeyUserID)
	if raw == nil {
		return uuid.Nil, errs.ErrInvalidToken
	}

	if id, ok := raw.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, errs.ErrInvalidToken
}
