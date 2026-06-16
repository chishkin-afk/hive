package bookingservices

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/devathh/hive/backend/internal/common/auth"
	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/modules/bookings/domain/booking"
	"github.com/devathh/hive/backend/pkg/errs"
	"github.com/google/uuid"
)

type scheduleService interface {
	GetScheduleByUserID(ctx context.Context, userID uuid.UUID, dayOfWeek uint8) (*responses.Schedule, error)
}

type clientService interface {
	Upsert(ctx context.Context, req *requests.UpsertClient) (*responses.Client, error)
}

type authService interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*responses.User, error)
}

type offerService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*responses.Offer, error)
}

// TODO: add slot's method

type BookingService struct {
	cfg                *config.Config
	log                *slog.Logger
	bookingPersistence booking.BookingPersistenceRepository
	bookingCache       booking.BookingCacheRepository

	scheduleService scheduleService
	clientService   clientService
	authService     authService
	offerService    offerService
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	bookingPersistence booking.BookingPersistenceRepository,
	bookingCache booking.BookingCacheRepository,
	scheduleService scheduleService,
	clientService clientService,
	authService authService,
	offerService offerService,
) *BookingService {
	return &BookingService{
		cfg:                cfg,
		log:                log,
		bookingPersistence: bookingPersistence,
		bookingCache:       bookingCache,
		scheduleService:    scheduleService,
		clientService:      clientService,
		authService:        authService,
		offerService:       offerService,
	}
}

// Create is a function that saves a Booking to the database,
// but before that, it performs a full validation to check if there are any overlaps,
// if it falls within the deadline, and if the user's subscription is paid.
func (bs *BookingService) Create(ctx context.Context, req *requests.CreateBooking) (*responses.Booking, error) {
	userID, err := bs.getUserID(ctx)
	if err != nil {
		return nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
	}

	bookingToSave, err := booking.New(req.OfferID, userID, req.Comment, time.Unix(req.StartTime, 0))
	if err != nil {
		return nil, errs.NewErrorWithKind(err, errs.KindRequest)
	}

	savedBooking, err := bs.bookingPersistence.Save(ctx, bookingToSave, func(ctx context.Context, bookings []*booking.Booking) error {
		_, offer, err := bs.getCoachOffer(ctx, bookingToSave)
		if err != nil {
			return err
		}

		bookingStartTime, bookingEndTime, err := bs.getBookingTime(bookingToSave.StartTime(), offer)
		if err != nil {
			return err
		}

		// Checks for overlap bookings
		for _, other := range bookings {
			otherStartTime := other.StartTime().Hour()*60 + other.StartTime().Minute()
			otherEndTime := otherStartTime + int(offer.SlotDurationMinutes)

			if bookingEndTime > otherStartTime &&
				bookingStartTime < otherEndTime {
				return errs.NewErrorWithKind(errs.ErrInvalidTimeForBooking, errs.KindRequest)
			}
		}

		schedule, err := bs.scheduleService.GetScheduleByUserID(ctx, offer.AuthorID, uint8(bookingToSave.StartTime().Weekday()))
		if err != nil {
			return err
		}

		if err := bs.checkSchedule(schedule, bookingStartTime, bookingEndTime); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		safeErr, ok := bs.handleRepoError(err)
		if !ok {
			bs.log.Error("failed to save booking into db",
				slog.String("error", err.Error()),
				slog.String("offer_id", req.OfferID.String()),
				slog.String("user_id", userID.String()),
			)
		}

		return nil, safeErr
	}

	if _, err := bs.clientService.Upsert(ctx, &requests.UpsertClient{
		OfferID: savedBooking.OfferID(),
	}); err != nil {
		bs.log.Warn("failed to upsert client into client's service", slog.String("error", err.Error()))
	}

	return bs.bookingToDTO(savedBooking), nil
}

// GetByID is a method that allows you to retrieve a record
// from the cache or database and return it to the user,
// with permission checking
func (bs *BookingService) GetByID(ctx context.Context, id uuid.UUID) (*responses.Booking, error) {
	userID, err := bs.getUserID(ctx)
	if err != nil {
		return nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
	}

	if cachedBooking, err := bs.bookingCache.GetByID(ctx, id); err == nil {
		if err := bs.checkReadRights(ctx, cachedBooking, userID); err != nil {
			return nil, err
		}

		return bs.bookingToDTO(cachedBooking), nil
	} else if !errors.Is(err, booking.ErrBookingNotFound) {
		bs.log.Warn("failed to get booking from cache",
			slog.String("error", err.Error()),
			slog.String("booking_id", id.String()),
		)
	}

	booking, err := bs.bookingPersistence.GetByID(ctx, id)
	if err != nil {
		safeErr, ok := bs.handleRepoError(err)
		if !ok {
			bs.log.Error("failed to get booking from db",
				slog.String("error", err.Error()),
				slog.String("booking_id", id.String()),
			)
		}

		return nil, safeErr
	}

	if err := bs.checkReadRights(ctx, booking, userID); err != nil {
		return nil, err
	}

	if err := bs.bookingCache.Save(ctx, booking); err != nil {
		bs.log.Warn("failed to save booking into cache",
			slog.String("error", err.Error()),
		)
	}

	return bs.bookingToDTO(booking), nil
}

// ListByOfferID returns a list of all bookings (with pagination)
// by offer's id. Accordingly, it checks whether the user is the author
func (bs *BookingService) ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) (*responses.Bookings, error) {
	if page < 1 || page > 1000 {
		return nil, errs.NewErrorWithKind(errs.ErrInvalidPage, errs.KindRequest)
	}

	if limit < 1 || limit > 100 {
		return nil, errs.NewErrorWithKind(errs.ErrInvalidLimit, errs.KindRequest)
	}

	userID, err := bs.getUserID(ctx)
	if err != nil {
		return nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
	}

	offer, err := bs.offerService.GetByID(ctx, offerID)
	if err != nil {
		return nil, err
	}

	if offer.AuthorID != userID {
		return nil, errs.NewErrorWithKind(errs.ErrNotEnoughRights, errs.KindPermissionDenied)
	}

	list, err := bs.bookingPersistence.ListByOfferID(ctx, offerID, page, limit)
	if err != nil {
		safeErr, ok := bs.handleRepoError(err)
		if !ok {
			bs.log.Error("failed to list bookings by offer id",
				slog.String("error", err.Error()),
				slog.String("offer_id", offerID.String()),
			)
		}

		return nil, safeErr
	}

	response := responses.Bookings{
		List: make([]*responses.Booking, len(list)),
	}
	for i, booking := range list {
		response.List[i] = bs.bookingToDTO(booking)
	}

	return &response, nil
}

// ListByUserID method that returns a list of all bookings
// with pagination by user's id from the token
func (bs *BookingService) ListByUserID(ctx context.Context, page, limit uint32) (*responses.Bookings, error) {
	if page < 1 || page > 1000 {
		return nil, errs.NewErrorWithKind(errs.ErrInvalidPage, errs.KindRequest)
	}

	if limit < 1 || limit > 100 {
		return nil, errs.NewErrorWithKind(errs.ErrInvalidLimit, errs.KindRequest)
	}

	userID, err := bs.getUserID(ctx)
	if err != nil {
		return nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
	}

	list, err := bs.bookingPersistence.ListByUserID(ctx, userID, page, limit)
	if err != nil {
		safeErr, ok := bs.handleRepoError(err)
		if !ok {
			bs.log.Error("failed to list bookings by offer id",
				slog.String("error", err.Error()),
				slog.String("user_id", userID.String()),
			)
		}

		return nil, safeErr
	}

	response := responses.Bookings{
		List: make([]*responses.Booking, len(list)),
	}
	for i, booking := range list {
		response.List[i] = bs.bookingToDTO(booking)
	}

	return &response, nil
}

// Update is a method that updates the fields of a booking.
// This method can only be called by the author of the offer
// or the client of the offer.
func (bs *BookingService) Update(ctx context.Context, req *requests.UpdateBooking) (*responses.Booking, error) {
	userID, err := bs.getUserID(ctx)
	if err != nil {
		return nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
	}

	updatedBooking, err := bs.bookingPersistence.Update(ctx, req.BookingID, func(ctx context.Context, booking *booking.Booking) error {
		if err := bs.checkReadRights(ctx, booking, userID); err != nil {
			return err
		}

		if err := bs.applyChanges(ctx, booking, req); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		safeErr, ok := bs.handleRepoError(err)
		if !ok {
			bs.log.Error("failed to update booking",
				slog.String("error", err.Error()),
				slog.String("booking_id", req.BookingID.String()),
			)
		}

		return nil, safeErr
	}

	if err := bs.bookingCache.Delete(ctx, updatedBooking.ID()); err != nil {
		bs.log.Warn("failed to delete booking from cache",
			slog.String("error", err.Error()),
			slog.String("booking_id", updatedBooking.ID().String()),
		)
	}

	return bs.bookingToDTO(updatedBooking), nil
}

func (bs *BookingService) applyChanges(ctx context.Context, booking *booking.Booking, req *requests.UpdateBooking) error {
	if req.Comment != nil {
		if err := booking.ChangeComment(*req.Comment); err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}
	}
	if req.IsCancelled != nil {
		if *req.IsCancelled {
			if err := booking.Cancel(); err != nil {
				return errs.NewErrorWithKind(err, errs.KindRequest)
			}
		}
	}
	if req.StartTime != nil {
		return bs.applyNewStartTime(ctx, booking, time.Unix(*req.StartTime, 0))
	}

	return nil
}

func (bs *BookingService) applyNewStartTime(ctx context.Context, booking *booking.Booking, newStartTime time.Time) error {
	offer, err := bs.offerService.GetByID(ctx, booking.OfferID())
	if err != nil {
		return err
	}

	bookingStartTime, bookingEndTime, err := bs.getBookingTime(newStartTime, offer)
	if err != nil {
		return err
	}

	allBookings, err := bs.bookingPersistence.ListAllByOfferID(ctx, booking.OfferID())
	if err != nil {
		safeErr, ok := bs.handleRepoError(err)
		if !ok {
			bs.log.Error("failed to get list of all the future bookings",
				slog.String("error", err.Error()),
				slog.String("offer_id", booking.OfferID().String()),
			)
		}

		return safeErr
	}

	for _, other := range allBookings {
		otherStartTime := other.StartTime().Hour()*60 + other.StartTime().Minute()
		otherEndTime := otherStartTime + int(offer.SlotDurationMinutes)

		if bookingEndTime > otherStartTime &&
			bookingStartTime < otherEndTime {
			return errs.NewErrorWithKind(errs.ErrInvalidTimeForBooking, errs.KindRequest)
		}
	}

	schedule, err := bs.scheduleService.GetScheduleByUserID(ctx, offer.AuthorID, uint8(newStartTime.Weekday()))
	if err != nil {
		return err
	}

	if err := bs.checkSchedule(schedule, bookingStartTime, bookingEndTime); err != nil {
		return err
	}

	return booking.ChangeStartTime(newStartTime)
}

func (bs *BookingService) checkReadRights(ctx context.Context, booking *booking.Booking, userID uuid.UUID) error {
	if booking.UserID() == userID {
		return nil
	}

	offer, err := bs.offerService.GetByID(ctx, booking.OfferID())
	if err != nil {
		return err
	}

	if offer.AuthorID != userID {
		return errs.NewErrorWithKind(errs.ErrNotEnoughRights, errs.KindPermissionDenied)
	}

	return nil
}

// checkSchedule is a mini method that validates whether
// a booking falls within the author's specified period
// or if it falls on a break.
func (bs *BookingService) checkSchedule(schedule *responses.Schedule, bStartTime, bEndTime int) error {
	// Checks for breaks of coach's schedule
	for _, br := range schedule.Breaks {
		brStartTime := int(br.StartTime.Hour)*60 + int(br.StartTime.Minute)
		brEndTime := int(br.EndTime.Hour)*60 + int(br.EndTime.Minute)

		if bEndTime > brStartTime &&
			bStartTime < brEndTime {
			return errs.NewErrorWithKind(errs.ErrInvalidTimeForBooking, errs.KindRequest)
		}
	}

	// Checks for start & end times of coach's schedule
	startTime := int(schedule.StartTime.Hour)*60 + int(schedule.StartTime.Minute)
	endTime := int(schedule.EndTime.Hour)*60 + int(schedule.EndTime.Minute)

	if startTime > bStartTime || endTime < bEndTime {
		return errs.NewErrorWithKind(errs.ErrInvalidTimeForBooking, errs.KindRequest)
	}

	return nil
}

// getBookingTime is a method that simply creates
// the start and end times of the recording in minutes (relative)
func (bs *BookingService) getBookingTime(startTime time.Time, offer *responses.Offer) (int, int, error) {
	bookingStartTime := startTime.Hour()*60 + startTime.Minute()
	bookingEndTime := int(offer.SlotDurationMinutes) + bookingStartTime
	if bookingStartTime >= bookingEndTime {
		return 0, 0, errs.NewErrorWithKind(errs.ErrInvalidTimeForBooking, errs.KindRequest)
	}

	return bookingStartTime, bookingEndTime, nil
}

// getCoachOffer is a function that receives coach and offer
// and validates them, then returns
func (bs *BookingService) getCoachOffer(ctx context.Context, booking *booking.Booking) (*responses.User, *responses.Offer, error) {
	offer, err := bs.offerService.GetByID(ctx, booking.OfferID())
	if err != nil {
		return nil, nil, err
	}

	coach, err := bs.authService.GetUserByID(ctx, offer.AuthorID)
	if err != nil {
		return nil, nil, err
	}

	// Checks for paid period of coach
	if coach.PaidUntil < booking.StartTime().Unix() {
		return nil, nil, errs.NewErrorWithKind(errs.ErrUserNotPaid, errs.KindPermissionDenied)
	}

	return coach, offer, nil
}

func (bs *BookingService) bookingToDTO(booking *booking.Booking) *responses.Booking {
	return &responses.Booking{
		ID:          booking.ID(),
		OfferID:     booking.OfferID(),
		UserID:      booking.UserID(),
		Comment:     booking.Comment(),
		StartTime:   booking.StartTime().Unix(),
		IsCancelled: booking.IsCancelled(),
		CreatedAt:   booking.CreatedAt().Unix(),
		UpdatedAt:   booking.UpdatedAt().Unix(),
	}
}

func (bs *BookingService) getUserID(ctx context.Context) (uuid.UUID, error) {
	raw := ctx.Value(auth.KeyUserID)
	if raw == nil {
		return uuid.Nil, errs.ErrInvalidToken
	}

	if id, ok := raw.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, errs.ErrInvalidToken
}

func (bs *BookingService) handleRepoError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return errs.NewErrorWithKind(err, errs.KindTimeout), true
	case errors.Is(err, booking.ErrBookingNotFound):
		return errs.NewErrorWithKind(err, errs.KindNotFound), true
	}

	return errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal), false
}
