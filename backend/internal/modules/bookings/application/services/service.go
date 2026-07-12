package bookingservices

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	contextvalue "github.com/devathh/hive/backend/internal/common/context"
	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/modules/bookings/domain/booking"
	"github.com/devathh/hive/backend/pkg/errs"
	"github.com/google/uuid"
)

type engageClient interface {
	SendMessage(ctx context.Context, chatID int64, msg string) error
}

type scheduleService interface {
	GetSchedulesByUserID(ctx context.Context, userID uuid.UUID) (*responses.Schedules, error)
	GetScheduleByUserID(ctx context.Context, userID uuid.UUID, dayOfWeek uint8) (*responses.Schedule, error)
}

type clientService interface {
	Upsert(ctx context.Context, req *requests.UpsertClient) (*responses.Client, error)
}

type authService interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*responses.User, error)
	Lookup(ctx context.Context, req *requests.Lookup) (*responses.LookupUsers, error)
}

type offerService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*responses.Offer, error)
}

type SMTPClient interface {
	Notify(ctx context.Context, body string, to []string) error
}

type BookingService struct {
	cfg                *config.Config
	log                *slog.Logger
	bookingPersistence booking.BookingPersistenceRepository
	bookingCache       booking.BookingCacheRepository

	scheduleService scheduleService
	clientService   clientService
	authService     authService
	offerService    offerService
	engageClient    engageClient

	sc SMTPClient
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
	engageClient engageClient,
	sc SMTPClient,
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
		engageClient:       engageClient,
		sc:                 sc,
	}
}

// Create is a function that saves a Booking to the database,
// but before that, it performs a full validation to check if there are any overlaps,
// if it falls within the deadline, and if the user's subscription is paid.
func (bs *BookingService) Create(ctx context.Context, req *requests.CreateBooking) (*responses.Booking, error) {
	userID, err := bs.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	bookingStartTime, err := bs.validateBookingTime(req.StartTime)
	if err != nil {
		return nil, err
	}

	coach, offer, err := bs.getCoachAndOffer(ctx, req.OfferID)
	if err != nil {
		return nil, err
	}

	if err := bs.checkBookingPrecondition(coach, offer, bookingStartTime); err != nil {
		return nil, err
	}

	bookingToSave, err := bs.buildNewBooking(req, userID, bookingStartTime)
	if err != nil {
		return nil, err
	}

	savedBooking, err := bs.saveBookingAtomically(ctx, offer, userID, bookingToSave)
	if err != nil {
		return nil, err
	}

	if err := bs.upsertClient(ctx, savedBooking.OfferID()); err != nil {
		bs.log.Error("failed to upsert client", slog.String("error", err.Error()))
	}

	go bs.notifyHolders(context.Background(), savedBooking, bs.bookingCreatedMsg(&baseInfo{
		title: offer.Title,
		id:    savedBooking.ID(),
	}, savedBooking.StartTime()))

	return bs.bookingToDTO(savedBooking), nil
}

// GetByID is a method that allows you to retrieve a record
// from the cache or database and return it to the user,
// with permission checking
func (bs *BookingService) GetByID(ctx context.Context, id uuid.UUID) (*responses.Booking, error) {
	userID, err := bs.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	if cachedBooking, ok := bs.getBookingFromCache(ctx, id); ok {
		if err := bs.checkReadRights(ctx, cachedBooking, userID); err != nil {
			return nil, err
		}

		return bs.bookingToDTO(cachedBooking), nil
	}

	booking, err := bs.getBookingFromPersistence(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := bs.checkReadRights(ctx, booking, userID); err != nil {
		return nil, err
	}

	if err := bs.storeBookingCache(ctx, booking); err != nil {
		bs.log.Warn("failed to store booking cache", slog.String("error", err.Error()))
	}

	return bs.bookingToDTO(booking), nil
}

// ListByOfferID returns a list of all bookings (with pagination)
// by offer's id. Accordingly, it checks whether the user is the author
func (bs *BookingService) ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) (*responses.Bookings, error) {
	if err := bs.validatePagination(page, limit); err != nil {
		return nil, err
	}

	userID, err := bs.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := bs.checkAuthor(ctx, offerID, userID); err != nil {
		return nil, err
	}

	list, totalCount, err := bs.listByOfferID(ctx, offerID, page, limit)
	if err != nil {
		return nil, err
	}

	users, err := bs.lookupUsers(ctx, list)
	if err != nil {
		return nil, err
	}

	return bs.listToDTO(list, users, totalCount, page, limit), nil
}

// ListByUserID method that returns a list of all bookings
// with pagination by user's id from the token
func (bs *BookingService) ListByUserID(ctx context.Context, page, limit uint32) (*responses.Bookings, error) {
	if err := bs.validatePagination(page, limit); err != nil {
		return nil, err
	}

	userID, err := bs.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	list, totalCount, err := bs.listByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}

	users, err := bs.lookupUsers(ctx, list)
	if err != nil {
		return nil, err
	}

	return bs.listToDTO(list, users, totalCount, page, limit), nil
}

// Update is a method that updates the fields of a booking.
// This method can only be called by the author of the offer
// or the client of the offer.
func (bs *BookingService) Update(ctx context.Context, req *requests.UpdateBooking) (*responses.Booking, error) {
	userID, err := bs.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	updatedBooking, err := bs.updateBookingStore(ctx, req, userID)
	if err != nil {
		return nil, err
	}

	if err := bs.clearBookingCache(ctx, updatedBooking.ID()); err != nil {
		bs.log.Warn("failed to clear booking cache", slog.String("error", err.Error()))
	}

	return bs.bookingToDTO(updatedBooking), nil
}

func (bs *BookingService) GetFreeSlots(ctx context.Context, offerID uuid.UUID, date time.Time) (*responses.Slots, error) {
	date = bs.normalizeDate(date)

	offer, err := bs.fetchOffer(ctx, offerID)
	if err != nil {
		return nil, err
	}

	schedule, err := bs.fetchScheduleForOffer(ctx, offer, date)
	if err != nil {
		return nil, err
	}

	dayStart, dayEnd := bs.computeTimeRange(schedule, date)

	bookings, err := bs.listBookingsInRange(ctx, offerID, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	return bs.slotsToDTO(
		bs.generateFreeSlots(schedule, offer, bookings),
		date,
	), nil
}

func (bs *BookingService) GetAllDays(ctx context.Context, offerID uuid.UUID, from, to time.Time) (*responses.Days, error) {
	parsedFrom, parsedTo, err := bs.validateAllDaysRequest(from, to)
	if err != nil {
		return nil, err
	}

	bookings, err := bs.listBookingsInRange(ctx, offerID, from, to)
	if err != nil {
		return nil, err
	}

	coach, offer, err := bs.getCoachAndOffer(ctx, offerID)
	if err != nil {
		return nil, err
	}

	schedules, err := bs.getAllSchedules(ctx, coach.ID)
	if err != nil {
		return nil, err
	}

	return bs.getDays(
		schedules,
		offer,
		bookings,
		time.Unix(coach.PaidUntil, 0),
		parsedFrom,
		parsedTo,
	)
}

func (bs *BookingService) getAllSchedules(ctx context.Context, userID uuid.UUID) (*responses.Schedules, error) {
	schedules, err := bs.scheduleService.GetSchedulesByUserID(ctx, userID)
	if err != nil {
		bs.log.Error("failed to get schedules by user's id",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
		)

		return nil, err
	}

	return schedules, nil
}

func (bs *BookingService) getDays(schedules *responses.Schedules, offer *responses.Offer, bookings []*booking.Booking, paidUntil, from, to time.Time) (*responses.Days, error) {
	approxDays := int(to.Sub(from).Hours()/24) + 1
	allDays := responses.Days{
		List: make([]responses.Day, 0, approxDays),
	}

	mpSchedules := make(map[uint8]*responses.Schedule)
	for _, schedule := range schedules.List {
		mpSchedules[schedule.DayOfWeek] = &schedule
	}

	end := to
	if end.Compare(paidUntil) > 0 {
		end = paidUntil
	}

	for !from.After(end) {
		day := responses.Day{
			Time: from.Unix(),
		}

		schedule, ok := mpSchedules[uint8(from.Weekday())]
		if ok && bs.checkFreeSlotsInDay(schedule, offer, bookings, from) {
			day.IsActive = true
		}

		allDays.List = append(allDays.List, day)
		from = from.AddDate(0, 0, 1)
	}

	return &allDays, nil
}

func (bs *BookingService) checkFreeSlotsInDay(schedule *responses.Schedule, offer *responses.Offer, bookings []*booking.Booking, day time.Time) bool {
	accepted := make([]*booking.Booking, 0, len(bookings))
	for _, b := range bookings {
		if b.StartTime().Month() == day.Month() &&
			b.StartTime().Day() == day.Day() {
			accepted = append(accepted, b)
		}
	}

	if len(bs.generateFreeSlots(schedule, offer, accepted)) > 0 {
		return true
	}

	return false
}

func (bs *BookingService) validateAllDaysRequest(from, to time.Time) (time.Time, time.Time, error) {
	from = time.Date(
		from.Year(),
		from.Month(),
		from.Day(),
		0, 0, 0, 0, time.UTC,
	)

	to = time.Date(
		to.Year(),
		to.Month(),
		to.Day(),
		0, 0, 0, 0, time.UTC,
	)

	diff := to.Sub(from)
	if diff < 24*time.Hour || diff > 14*24*time.Hour {
		return time.Time{}, time.Time{}, errs.NewErrorWithKind(errs.ErrInvalidDays, errs.KindRequest)
	}

	return from, to, nil
}

func (bs *BookingService) lookupUsers(ctx context.Context, bookings []*booking.Booking) (*responses.LookupUsers, error) {
	req := requests.Lookup{
		IDs: make([]uuid.UUID, len(bookings)),
	}
	for i, booking := range bookings {
		req.IDs[i] = booking.UserID()
	}

	resp, err := bs.authService.Lookup(ctx, &req)
	if err != nil {
		bs.log.Error("failed to lookup users",
			slog.String("error", err.Error()),
		)

		return nil, errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal)
	}

	return resp, nil
}

func (bs *BookingService) notifyHolders(ctx context.Context, booking *booking.Booking, msg string) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client, err := bs.authService.GetUserByID(ctx, booking.UserID())
	if err == nil {
		if err = bs.notifyUser(ctx, client, msg); err != nil {
			bs.log.Error("failed to notify client",
				slog.String("error", err.Error()),
			)

			go bs.notifyByMail(context.Background(), msg, client.Email)
		}

		bs.log.Debug("client was be notified",
			slog.String("client_id", client.ID.String()),
		)
	} else {
		bs.log.Error("failed to get client",
			slog.String("error", err.Error()),
		)
	}

	coach, _, err := bs.getCoachAndOffer(ctxTimeout, booking.OfferID())
	if err != nil {
		bs.log.Error("failed to get coach",
			slog.String("error", err.Error()),
		)

		return
	}

	if coach.ID == client.ID {
		return
	}

	if err := bs.notifyUser(ctx, coach, msg); err != nil {
		bs.log.Error("failed to notify coach",
			slog.String("error", err.Error()),
		)
	}
}

func (bs *BookingService) notifyByMail(ctx context.Context, msg string, to string) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := bs.sc.Notify(ctxTimeout, msg, []string{to}); err != nil {
		bs.log.Error("failed to notify client by email",
			slog.String("error", err.Error()),
		)
	}
}

func (bs *BookingService) convertDate(t time.Time) (time.Time, string) {
	moscowLocation, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return t, "UTC"
	}

	return t.In(moscowLocation), "MSK"
}

func (bs *BookingService) getUserID(ctx context.Context) (uuid.UUID, error) {
	raw := ctx.Value(contextvalue.KeyUserID)
	if raw == nil {
		return uuid.Nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
	}

	if id, ok := raw.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
}

func (bs *BookingService) validateBookingTime(startTime int64) (time.Time, error) {
	bookingStartTime := time.Unix(startTime, 0).UTC()
	if bookingStartTime.Compare(time.Now().UTC()) < 0 {
		return time.Time{}, errs.NewErrorWithKind(errs.ErrInvalidTimeForBooking, errs.KindRequest)
	}

	return bookingStartTime, nil
}

// getCoachOffer is a function that receives coach and offer
// and validates them, then returns
func (bs *BookingService) getCoachAndOffer(ctx context.Context, offerID uuid.UUID) (*responses.User, *responses.Offer, error) {
	offer, err := bs.offerService.GetByID(ctx, offerID)
	if err != nil {
		return nil, nil, err
	}

	coach, err := bs.authService.GetUserByID(ctx, offer.AuthorID)
	if err != nil {
		return nil, nil, err
	}

	return coach, offer, nil
}

func (bs *BookingService) checkBookingPrecondition(coach *responses.User, offer *responses.Offer, startTime time.Time) error {
	if startTime.Add(
		time.Duration(offer.SlotDurationMinutes+offer.SlotIntervalMinutes)*time.Minute,
	).Day() != startTime.Day() {
		return errs.NewErrorWithKind(errs.ErrInvalidTimeForBooking, errs.KindRequest)
	}

	// Checks for paid until
	if coach.PaidUntil < startTime.Unix() {
		return errs.NewErrorWithKind(errs.ErrUserNotPaid, errs.KindRequest)
	}

	return nil
}

func (bs *BookingService) buildNewBooking(req *requests.CreateBooking, userID uuid.UUID, startTime time.Time) (*booking.Booking, error) {
	booking, err := booking.New(
		req.OfferID,
		userID,
		req.Comment,
		startTime,
	)
	if err != nil {
		return nil, errs.NewErrorWithKind(err, errs.KindRequest)
	}

	return booking, nil
}

func (bs *BookingService) upsertClient(ctx context.Context, offerID uuid.UUID) error {
	_, err := bs.clientService.Upsert(ctx, &requests.UpsertClient{
		OfferID: offerID,
	})

	return err
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

func (bs *BookingService) getBookingFromCache(ctx context.Context, id uuid.UUID) (*booking.Booking, bool) {
	cachedBooking, err := bs.bookingCache.GetByID(ctx, id)
	if err != nil {
		if !errors.Is(err, booking.ErrBookingNotFound) {
			bs.log.Warn("failed to get booking from cache",
				slog.String("booking_id", id.String()),
				slog.String("error", err.Error()),
			)
		}

		return nil, false
	}

	bs.log.Debug("booking has been taken from cache", slog.String("booking_id", id.String()))
	return cachedBooking, true
}

func (bs *BookingService) checkReadRights(ctx context.Context, booking *booking.Booking, userID uuid.UUID) error {
	if booking.UserID() == userID {
		return nil
	}

	offer, err := bs.fetchOffer(ctx, booking.OfferID())
	if err != nil {
		return err
	}

	if offer.AuthorID != userID {
		return errs.NewErrorWithKind(errs.ErrNotEnoughRights, errs.KindPermissionDenied)
	}

	return nil
}

func (bs *BookingService) getBookingFromPersistence(ctx context.Context, id uuid.UUID) (*booking.Booking, error) {
	booking, err := bs.bookingPersistence.GetByID(ctx, id)
	if err != nil {
		bs.log.Error("failed to get booking from db",
			slog.String("error", err.Error()),
			slog.String("booking_id", id.String()),
		)

		return nil, bs.handleRepoError(err)
	}

	return booking, nil
}

func (bs *BookingService) storeBookingCache(ctx context.Context, booking *booking.Booking) error {
	return bs.bookingCache.Save(ctx, booking)
}

func (bs *BookingService) validatePagination(page, limit uint32) error {
	if page < 1 || page > 1000 {
		return errs.NewErrorWithKind(errs.ErrInvalidPage, errs.KindRequest)
	}

	if limit < 1 || limit > 100 {
		return errs.NewErrorWithKind(errs.ErrInvalidLimit, errs.KindRequest)
	}

	return nil
}

func (bs *BookingService) checkAuthor(ctx context.Context, offerID, userID uuid.UUID) error {
	offer, err := bs.offerService.GetByID(ctx, offerID)
	if err != nil {
		return err
	}

	if offer.AuthorID != userID {
		return errs.NewErrorWithKind(errs.ErrNotEnoughRights, errs.KindPermissionDenied)
	}

	return nil
}

func (bs *BookingService) listByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) ([]*booking.Booking, int64, error) {
	list, totalCount, err := bs.bookingPersistence.ListByOfferID(ctx, offerID, page, limit)
	if err != nil {
		bs.log.Error("failed to list the bookings by offer id",
			slog.String("error", err.Error()),
			slog.String("offer_id", offerID.String()),
		)

		return nil, 0, bs.handleRepoError(err)
	}

	return list, totalCount, err
}

func (bs *BookingService) listByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*booking.Booking, int64, error) {
	list, totalCount, err := bs.bookingPersistence.ListByUserID(ctx, userID, page, limit)
	if err != nil {
		bs.log.Error("failed to list the bookings by offer id",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
		)

		return nil, 0, bs.handleRepoError(err)
	}

	return list, totalCount, err
}

func (bs *BookingService) listToDTO(list []*booking.Booking, users *responses.LookupUsers, totalCount int64, page, limit uint32) *responses.Bookings {
	userMap := make(map[uuid.UUID]*responses.LookupUser, len(users.List))
	for _, user := range users.List {
		userMap[user.ID] = &user
	}

	totalPages := (totalCount + int64(limit) - 1) / int64(limit)
	response := responses.Bookings{
		List:        make([]responses.ListElement, 0, len(list)),
		HasNextPage: totalPages > int64(page),
		HasPrevPage: page > 1,
		TotalPages:  uint32(totalPages),
	}
	for _, booking := range list {
		if user, ok := userMap[booking.UserID()]; ok {
			response.List = append(response.List, responses.ListElement{
				Booking: bs.bookingToDTO(booking),
				User: &responses.BookingUser{
					ID:             user.ID,
					Firstname:      user.Firstname,
					Lastname:       user.Lastname,
					ConfirmedEmail: user.ConfirmedEmail,
				},
			})

			continue
		}

		bs.log.Warn("tried to get unexisted user", slog.String("user_id", booking.UserID().String()))
	}

	return &response
}

func (bs *BookingService) updateBookingStore(ctx context.Context, req *requests.UpdateBooking, userID uuid.UUID) (*booking.Booking, error) {
	updFunc := func(ctx context.Context, booking *booking.Booking) error {
		if err := bs.checkReadRights(ctx, booking, userID); err != nil {
			return err
		}

		if err := bs.applyChanges(ctx, booking, req); err != nil {
			return err
		}

		return nil
	}

	updatedBooking, err := bs.bookingPersistence.Update(ctx, req.BookingID, updFunc)
	if err != nil {
		bs.log.Error("failed to update booking in db",
			slog.String("error", err.Error()),
			slog.String("booking_id", req.BookingID.String()),
			slog.String("user_id", userID.String()),
		)

		return nil, bs.handleRepoError(err)
	}

	return updatedBooking, nil
}

func (bs *BookingService) clearBookingCache(ctx context.Context, id uuid.UUID) error {
	return bs.bookingCache.Delete(ctx, id)
}

func (bs *BookingService) normalizeDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}

func (bs *BookingService) fetchOffer(ctx context.Context, offerID uuid.UUID) (*responses.Offer, error) {
	return bs.offerService.GetByID(ctx, offerID)
}

func (bs *BookingService) fetchScheduleForOffer(ctx context.Context, offer *responses.Offer, date time.Time) (*responses.Schedule, error) {
	return bs.scheduleService.GetScheduleByUserID(ctx, offer.AuthorID, uint8(date.Weekday()))
}

func (bs *BookingService) computeTimeRange(schedule *responses.Schedule, date time.Time) (time.Time, time.Time) {
	startDuration := time.Duration(schedule.StartTime.Hour)*time.Hour + time.Duration(schedule.StartTime.Minute)*time.Minute
	dayStart := date.Add(startDuration)
	endDuration := time.Duration(schedule.EndTime.Hour)*time.Hour + time.Duration(schedule.EndTime.Minute)*time.Minute
	dayEnd := date.Add(endDuration)

	return dayStart, dayEnd
}

func (bs *BookingService) listBookingsInRange(ctx context.Context, offerID uuid.UUID, dayStart, dayEnd time.Time) ([]*booking.Booking, error) {
	bookings, err := bs.bookingPersistence.ListAllPerTime(ctx, offerID, dayStart, dayEnd)
	if err != nil {
		bs.log.Error("failed to list all per time",
			slog.String("error", err.Error()),
			slog.String("offer_id", offerID.String()),
		)

		return nil, bs.handleRepoError(err)
	}

	return bookings, nil
}

func (bs *BookingService) slotsToDTO(slots []responses.DayTime, date time.Time) *responses.Slots {
	year, month, day := date.Date()
	return &responses.Slots{
		Date:  fmt.Sprintf("%d-%s-%d", year, month.String(), day),
		Times: slots,
	}
}

func (bs *BookingService) generateFreeSlots(schedule *responses.Schedule, offer *responses.Offer, bookings []*booking.Booking) []responses.DayTime {
	slots := make([]responses.DayTime, 0)
	currSlot := bs.allMinutes(&schedule.StartTime)
	slotInterval := offer.SlotIntervalMinutes
	for currSlot < bs.allMinutes(&schedule.EndTime)-int64(offer.SlotIntervalMinutes+offer.SlotDurationMinutes) {
		slot := responses.DayTime{
			Hour:   uint8(currSlot / 60),
			Minute: uint8(currSlot % 60),
		}

		startSlot := int(currSlot)
		endSlot := int(currSlot) + int(offer.SlotDurationMinutes+offer.SlotIntervalMinutes)

		currSlot += int64(slotInterval)

		if !bs.checkBookings(offer, bookings, startSlot, endSlot) {
			continue
		}

		if !bs.checkBreaks(schedule, startSlot, endSlot) {
			continue
		}

		slots = append(slots, slot)
	}

	return slots
}

func (bs *BookingService) saveBookingAtomically(
	ctx context.Context,
	offer *responses.Offer,
	userID uuid.UUID,
	bookingToSave *booking.Booking,
) (*booking.Booking, error) {
	startDate := time.Date(bookingToSave.StartTime().Year(), bookingToSave.StartTime().Month(), bookingToSave.StartTime().Day(), 0, 0, 0, 0, time.UTC)
	endDate := startDate.Add(24 * time.Hour)

	checkFn := func(ctx context.Context, others []*booking.Booking) (*booking.Booking, error) {
		slotDuration := int(offer.SlotDurationMinutes + offer.SlotIntervalMinutes)
		if bs.hasOverlap(bookingToSave.StartTime(), slotDuration, others) {
			return nil, errs.NewErrorWithKind(errs.ErrInvalidTimeForBooking, errs.KindRequest)
		}

		if bs.hasBooking(userID, others) {
			return nil, errs.NewErrorWithKind(booking.ErrTooMuchBookings, errs.KindRequest)
		}

		schedule, err := bs.scheduleService.GetScheduleByUserID(ctx, offer.AuthorID, uint8(bookingToSave.StartTime().Weekday()))
		if err != nil {
			return nil, err
		}

		startMinutes, endMinutes, err := bs.getBookingTime(bookingToSave.StartTime(), offer)
		if err != nil {
			return nil, err
		}

		if err := bs.checkSchedule(schedule, startMinutes, endMinutes); err != nil {
			return nil, err
		}

		return bookingToSave, nil
	}

	saved, err := bs.bookingPersistence.Save(ctx, offer.ID, startDate, endDate, checkFn)
	if err != nil {
		bs.log.Error("failed to save booking into db",
			slog.String("error", err.Error()),
			slog.String("offer_id", offer.ID.String()),
			slog.String("user_id", userID.String()),
		)

		return nil, bs.handleRepoError(err)
	}

	return saved, nil
}

func (bs *BookingService) hasOverlap(newStart time.Time, durationMinutes int, existingBookings []*booking.Booking) bool {
	newEnd := newStart.Add(time.Duration(durationMinutes) * time.Minute)

	for _, existingBooking := range existingBookings {
		existingEnd := existingBooking.StartTime().Add(time.Duration(durationMinutes) * time.Minute)

		if newStart.Before(existingEnd) && existingBooking.StartTime().Before(newEnd) {
			return true
		}
	}
	return false
}

func (bs *BookingService) hasBooking(userID uuid.UUID, existingBookings []*booking.Booking) bool {
	for _, booking := range existingBookings {
		if booking.UserID() == userID {
			return true
		}
	}

	return false
}

func (bs *BookingService) checkBreaks(schedule *responses.Schedule, startSlot, endSlot int) bool {
	for _, b := range schedule.Breaks {
		bStartTime := int(b.StartTime.Hour)*60 + int(b.StartTime.Minute)
		bEndTime := int(b.EndTime.Hour)*60 + int(b.EndTime.Minute)

		if endSlot > int(bStartTime) &&
			startSlot < bEndTime {
			return false
		}
	}

	return true
}

func (bs *BookingService) checkBookings(offer *responses.Offer, bookings []*booking.Booking, startSlot, endSlot int) bool {
	for _, booking := range bookings {
		otherStartTime := booking.StartTime().Hour()*60 + booking.StartTime().Minute()
		otherEndTime := otherStartTime + int(offer.SlotDurationMinutes+offer.SlotIntervalMinutes)

		if endSlot > otherStartTime &&
			startSlot < otherEndTime {
			return false
		}
	}

	return true
}

func (bs *BookingService) allMinutes(dayTime *responses.DayTime) int64 {
	return int64(dayTime.Hour)*60 + int64(dayTime.Minute)
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

			go bs.notifyHolders(
				context.Background(),
				booking,
				bs.getBookingCanceledMsg(context.Background(), booking),
			)
		}
	}
	if req.StartTime != nil {
		return bs.applyNewStartTime(ctx, booking, time.Unix(*req.StartTime, 0))
	}

	return nil
}

func (bs *BookingService) getBookingCanceledMsg(ctx context.Context, booking *booking.Booking) string {
	ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	offer, err := bs.offerService.GetByID(ctxTimeout, booking.OfferID())
	if err != nil {
		bs.log.Error("failed to get title of offer",
			slog.String("error", err.Error()),
		)

		return ""
	}

	return bs.bookingCanceledMsg(&baseInfo{
		title: offer.Title,
		id:    booking.ID(),
	}, booking.StartTime())
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

	startDate := time.Date(newStartTime.Year(), newStartTime.Month(), newStartTime.Day(), 0, 0, 0, 0, time.UTC)
	endDate := startDate.Add(24 * time.Hour)

	allBookings, err := bs.listBookingsInRange(ctx, booking.OfferID(), startDate, endDate)
	if err != nil {
		return err
	}

	for _, other := range allBookings {
		otherStartTime := other.StartTime().Hour()*60 + other.StartTime().Minute()
		otherEndTime := otherStartTime + int(offer.SlotDurationMinutes+offer.SlotIntervalMinutes)

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

	go bs.notifyHolders(context.Background(), booking, bs.bookingUpdatedMsg(
		&baseInfo{
			title: offer.Title,
			id:    booking.ID(),
		},
		newStartTime,
		booking.StartTime(),
	))

	return booking.ChangeStartTime(newStartTime)
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
			return errs.ErrInvalidTimeForBooking
		}
	}

	// Checks for start & end times of coach's schedule
	startTime := int(schedule.StartTime.Hour)*60 + int(schedule.StartTime.Minute)
	endTime := int(schedule.EndTime.Hour)*60 + int(schedule.EndTime.Minute)

	if startTime > bStartTime || endTime < bEndTime {
		return errs.ErrInvalidTimeForBooking
	}

	return nil
}

// getBookingTime is a method that simply creates
// the start and end times of the recording in minutes (relative)
func (bs *BookingService) getBookingTime(startTime time.Time, offer *responses.Offer) (int, int, error) {
	bookingStartTime := startTime.Hour()*60 + startTime.Minute()
	bookingEndTime := int(offer.SlotDurationMinutes) + bookingStartTime
	if bookingStartTime >= bookingEndTime {
		return 0, 0, errs.ErrInvalidTimeForBooking
	}

	return bookingStartTime, bookingEndTime, nil
}

func (bs *BookingService) getMonthRU(month string) string {
	switch month {
	case "January":
		return "Января"
	case "February":
		return "Февраля"
	case "March":
		return "Марта"
	case "April":
		return "Апреля"
	case "May":
		return "Мая"
	case "June":
		return "Июня"
	case "July":
		return "Июля"
	case "August":
		return "Августа"
	case "September":
		return "Сентября"
	case "October":
		return "Октября"
	case "November":
		return "Ноября"
	case "December":
		return "Декабря"
	}

	return month
}

func (bs *BookingService) handleRepoError(err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return errs.NewErrorWithKind(err, errs.KindTimeout)
	case errors.Is(err, booking.ErrBookingNotFound):
		return errs.NewErrorWithKind(err, errs.KindNotFound)
	case errors.Is(err, booking.ErrBookingAlreadyExists):
		return errs.NewErrorWithKind(err, errs.KindConflict)
	case errors.Is(err, booking.ErrInvalidOfferID),
		errors.Is(err, errs.ErrInvalidTimeForBooking):
		return errs.NewErrorWithKind(err, errs.KindRequest)
	case errors.Is(err, booking.ErrTooMuchBookings),
		errors.Is(err, errs.ErrUserNotPaid):
		return errs.NewErrorWithKind(err, errs.KindPermissionDenied)
	}

	return errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal)
}
