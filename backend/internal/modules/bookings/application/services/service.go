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

var mskLocation *time.Location

func init() {
	var err error
	mskLocation, err = time.LoadLocation("Europe/Moscow")
	if err != nil {
		panic("failed to load MSK timezone: " + err.Error())
	}
}

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
	Notify(ctx context.Context, body string, subject string, to []string) error
	NotifyHTML(ctx context.Context, body string, subject string, to []string) error
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

// Create is a function that saves a booking to the database,
// performing all necessary checks (including overlaps,
// a counter of the user's existing bookings,
// and the trainer's schedule).
func (bs *BookingService) Create(ctx context.Context, req *requests.CreateBooking) (*responses.Booking, error) {
	userID, err := bs.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	offer, coach, err := bs.getOfferAndCoach(ctx, req.OfferID)
	if err != nil {
		return nil, err
	}

	bookingStartTime, err := bs.parseStartTime(offer, &req.StartTime)
	if err != nil {
		return nil, err
	}

	if err := bs.checkPaidUntil(coach, bookingStartTime); err != nil {
		return nil, err
	}

	if coach.OnVacation {
		return nil, errs.NewErrorWithKind(errs.ErrCoachOnVacation, errs.KindPermissionDenied)
	}

	if !offer.IsActive {
		return nil, errs.NewErrorWithKind(errs.ErrOfferIsntActive, errs.KindPermissionDenied)
	}

	schedule, err := bs.getSchedule(ctx, coach.ID, uint8(bookingStartTime.Weekday()))
	if err != nil {
		return nil, err
	}

	if err := bs.checkSchedule(schedule, offer, bookingStartTime); err != nil {
		return nil, err
	}

	savedBooking, err := bs.storeBookingPersistence(ctx, userID, req.Comment, offer, bookingStartTime)
	if err != nil {
		return nil, err
	}

	if err := bs.upsertClient(ctx, savedBooking.OfferID(), req.StartTime); err != nil {
		bs.log.Error("failed to upsert client",
			slog.String("error", err.Error()),
		)
	}

	client, err := bs.authService.GetUserByID(ctx, userID)
	clientName := "Клиент"
	if err == nil {
		clientName = client.Firstname + " " + client.Lastname
	} else {
		bs.log.Error("failed to get client for notifyCreated",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
		)
	}

	go bs.notifyCreated(
		context.Background(),
		coach,
		offer,
		savedBooking,
		clientName,
	)

	return bs.bookingToDTO(savedBooking), nil
}

func (bs *BookingService) checkPaidUntil(user *responses.User, bookingStartTime time.Time) error {
	pu := time.Unix(user.PaidUntil, 0).In(mskLocation)
	if pu.Compare(bookingStartTime) < 0 {
		return errs.NewErrorWithKind(errs.ErrUserNotPaid, errs.KindRequest)
	}

	return nil
}

// GetByID is a method that returns a booking by its ID
// (from the cache or database),
// including a check for read permissions.
func (bs *BookingService) GetByID(ctx context.Context, id uuid.UUID) (*responses.Booking, error) {
	userID, err := bs.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	if cachedBooking, ok := bs.getBookingCache(ctx, id); ok {
		if err := bs.checkReadRights(ctx, cachedBooking, userID); err != nil {
			return nil, err
		}

		return bs.bookingToDTO(cachedBooking), nil
	}

	booking, err := bs.getBookingPersistence(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := bs.checkReadRights(ctx, booking, userID); err != nil {
		return nil, err
	}

	go bs.storeBookingCache(context.Background(), booking)
	return bs.bookingToDTO(booking), nil
}

// ListByOfferID is a method that
// returns a paginated list of bookings
// and checks whether the author is a user.
func (bs *BookingService) ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) (*responses.Bookings, error) {
	userID, err := bs.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := bs.validatePagination(page, limit); err != nil {
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

	return bs.listToDTO(
		list,
		users,
		totalCount,
		page,
		limit,
	), nil
}

func (bs *BookingService) listByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) ([]*booking.Booking, int64, error) {
	list, totalCount, err := bs.bookingPersistence.ListByOfferID(ctx, offerID, page, limit)
	if err != nil {
		bs.log.Error("failed to list bookings by offer's id",
			slog.String("error", err.Error()),
			slog.String("offer_id", offerID.String()),
		)

		return nil, 0, handleRepoError(err)
	}

	return list, totalCount, nil
}

// ListByUserID is a method that
// returns a paginated list of bookings
// based on the user's ID.
func (bs *BookingService) ListByUserID(ctx context.Context, page, limit uint32) (*responses.Bookings, error) {
	userID, err := bs.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	if err := bs.validatePagination(page, limit); err != nil {
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
		bs.log.Error("failed to clear booking cache",
			slog.String("error", err.Error()),
		)
	}

	return bs.bookingToDTO(updatedBooking), nil
}

// GetFreeSlots is a method that
// returns a list of all
// available slots for a given day.
func (bs *BookingService) GetFreeSlots(ctx context.Context, offerID uuid.UUID, date time.Time) (*responses.Slots, error) {
	date = bs.normalizeDate(date)

	offer, err := bs.fetchOffer(ctx, offerID)
	if err != nil {
		return nil, err
	}

	schedule, err := bs.fetchScheduleForOffer(ctx, offer, date)
	if err != nil {
		// Если расписания нет на этот день — возвращаем пустой список слотов
		bs.log.Debug("no schedule found for this day",
			slog.String("offer_id", offerID.String()),
			slog.Time("date", date),
		)
		return &responses.Slots{
			Date:  fmt.Sprintf("%d-%02d-%02d", date.Year(), date.Month(), date.Day()),
			Times: []responses.DayTime{},
		}, nil
	}

	dayStart, dayEnd := bs.computeTimeRange(schedule, date)

	bookings, err := bs.listBookingsInRange(ctx, offer.AuthorID, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}

	slots := bs.generateFreeSlots(schedule, offer, bookings)
	nowMSK := time.Now().In(mskLocation)
	// Convert date to MSK midnight
	dateMSK := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, mskLocation)
	filteredSlots := make([]responses.DayTime, 0, len(slots))
	for _, slot := range slots {
		slotTimeMSK := dateMSK.Add(time.Duration(int(slot.Hour))*time.Hour + time.Duration(int(slot.Minute))*time.Minute)
		if slotTimeMSK.After(nowMSK) {
			filteredSlots = append(filteredSlots, slot)
		}
	}

	return bs.slotsToDTO(filteredSlots, date), nil
}

// GetAllDays is a method that returns a list of days
// within a specified range, indicating
// whether each day is a workday or not.
func (bs *BookingService) GetAllDays(ctx context.Context, offerID uuid.UUID, from, to time.Time) (*responses.Days, error) {
	parsedFrom, parsedTo, err := bs.validateAllDaysRequest(from, to)
	if err != nil {
		return nil, err
	}

	offer, coach, err := bs.getOfferAndCoach(ctx, offerID)
	if err != nil {
		return nil, err
	}

	bookings, err := bs.listBookingsInRange(ctx, offer.AuthorID, from, to)
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

func (bs *BookingService) validateAllDaysRequest(from, to time.Time) (time.Time, time.Time, error) {
	from = time.Date(
		from.Year(),
		from.Month(),
		from.Day(),
		0, 0, 0, 0, mskLocation,
	)

	to = time.Date(
		to.Year(),
		to.Month(),
		to.Day(),
		0, 0, 0, 0, mskLocation,
	)

	diff := to.Sub(from)
	if diff < 24*time.Hour || diff > 14*24*time.Hour {
		return time.Time{}, time.Time{}, errs.NewErrorWithKind(errs.ErrInvalidDays, errs.KindRequest)
	}

	return from, to, nil
}

func (bs *BookingService) slotsToDTO(slots []responses.DayTime, date time.Time) *responses.Slots {
	year, month, day := date.Date()
	return &responses.Slots{
		Date:  fmt.Sprintf("%d-%02d-%02d", year, month, day),
		Times: slots,
	}
}

func (bs *BookingService) generateFreeSlots(schedule *responses.Schedule, offer *responses.Offer, bookings []*booking.Booking) []responses.DayTime {
	slots := make([]responses.DayTime, 0)
	currSlot := bs.allMinutes(&schedule.StartTime)

	endTimeLimit := bs.allMinutes(&schedule.EndTime) - int64(offer.SlotDurationMinutes)

	for currSlot <= endTimeLimit {
		startSlot := int(currSlot)
		endSlot := startSlot + int(offer.SlotDurationMinutes)
		// Конец слота с учётом интервала
		endSlotWithInterval := endSlot + int(offer.SlotIntervalMinutes)

		// Проверяем, не пересекается ли слот + интервал с бронями и перерывами
		isAvailable := bs.checkBookingsWithInterval(offer, bookings, startSlot, endSlotWithInterval) &&
			bs.checkBreaksWithInterval(schedule, startSlot, endSlotWithInterval)

		if isAvailable {
			slots = append(slots, responses.DayTime{
				Hour:   uint8(currSlot / 60),
				Minute: uint8(currSlot % 60),
			})
		}

		// Если слот занят, следующий сдвигается на duration + interval
		// Если свободен — шаг duration
		if !isAvailable {
			currSlot += int64(offer.SlotDurationMinutes + offer.SlotIntervalMinutes)
		} else {
			currSlot += int64(offer.SlotDurationMinutes)
		}
	}

	return slots
}

// checkBookingsWithInterval проверяет, что слот + интервал не пересекается с другими записями
func (bs *BookingService) checkBookingsWithInterval(offer *responses.Offer, bookings []*booking.Booking, startSlot, endSlotWithInterval int) bool {
	for _, booking := range bookings {
		// Пропускаем отменённые записи — они не блокируют слоты
		if booking.IsCancelled() {
			continue
		}

		startTime := booking.StartTime().In(mskLocation)
		otherStartTime := startTime.Hour()*60 + startTime.Minute()
		otherEndTimeWithInterval := otherStartTime + int(booking.OfferDurationMinutes()) + int(booking.OfferIntervalMinutes())

		if intervalsOverlap(
			int64(startSlot),
			int64(endSlotWithInterval),
			int64(otherStartTime),
			int64(otherEndTimeWithInterval),
		) {
			return false
		}
	}

	return true
}

// checkBreaksWithInterval проверяет, что слот + интервал не пересекается с перерывами
func (bs *BookingService) checkBreaksWithInterval(schedule *responses.Schedule, startSlot, endSlotWithInterval int) bool {
	for _, b := range schedule.Breaks {
		bStartTime := int(b.StartTime.Hour)*60 + int(b.StartTime.Minute)
		bEndTime := int(b.EndTime.Hour)*60 + int(b.EndTime.Minute)

		if endSlotWithInterval > int(bStartTime) &&
			startSlot < bEndTime {
			return false
		}
	}

	return true
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
		// Пропускаем отменённые записи — они не блокируют слоты
		if booking.IsCancelled() {
			continue
		}

		startTime := booking.StartTime().In(mskLocation)
		otherStartTime := startTime.Hour()*60 + startTime.Minute()
		otherEndTime := otherStartTime + int(booking.OfferDurationMinutes()) + int(booking.OfferIntervalMinutes())

		if intervalsOverlap(
			int64(startSlot),
			int64(endSlot),
			int64(otherStartTime),
			int64(otherEndTime),
		) {
			return false
		}
	}

	return true
}

func (bs *BookingService) allMinutes(dayTime *responses.DayTime) int64 {
	return int64(dayTime.Hour)*60 + int64(dayTime.Minute)
}

func (bs *BookingService) clearBookingCache(ctx context.Context, id uuid.UUID) error {
	return bs.bookingCache.Delete(ctx, id)
}

func (bs *BookingService) computeTimeRange(schedule *responses.Schedule, date time.Time) (time.Time, time.Time) {
	startDuration := int64(schedule.StartTime.Hour)*60 + int64(schedule.StartTime.Minute)
	dayStart := date.Add(time.Duration(startDuration) * time.Minute)
	endDuration := int64(schedule.EndTime.Hour)*60 + int64(schedule.EndTime.Minute)
	dayEnd := date.Add(time.Duration(endDuration) * time.Minute)

	return dayStart, dayEnd
}

func (bs *BookingService) fetchScheduleForOffer(ctx context.Context, offer *responses.Offer, date time.Time) (*responses.Schedule, error) {
	return bs.scheduleService.GetScheduleByUserID(ctx, offer.AuthorID, uint8(date.Weekday()))
}

func (bs *BookingService) normalizeDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, mskLocation)
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

		return nil, handleRepoError(err)
	}

	return updatedBooking, nil
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

			go func() {
				ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				offer, coach, err := bs.getOfferAndCoach(ctxTimeout, booking.OfferID())
				if err != nil {
					bs.log.Error("failed to get offer & coach",
						slog.String("error", err.Error()),
					)
					return
				}

				client, err := bs.authService.GetUserByID(ctxTimeout, booking.UserID())
				clientName := "Клиент"
				if err == nil {
					clientName = client.Firstname + " " + client.Lastname
				} else {
					bs.log.Error("failed to get client for notifyCanceled",
						slog.String("error", err.Error()),
						slog.String("user_id", booking.UserID().String()),
					)
				}

				bs.notifyCanceled(
					ctxTimeout,
					coach,
					offer,
					booking,
					clientName,
				)
			}()
		}
	}
	if req.StartTime != nil {
		return bs.applyNewStartTime(ctx, booking, *req.StartTime)
	}

	return nil
}

func (bs *BookingService) applyNewStartTime(ctx context.Context, booking *booking.Booking, newStartTime requests.Time) error {
	offer, coach, err := bs.getOfferAndCoach(ctx, booking.OfferID())
	if err != nil {
		return err
	}

	bookingStartTime, err := bs.parseStartTime(offer, &newStartTime)
	if err != nil {
		return err
	}

	now := time.Now().In(mskLocation)
	if now.Compare(bookingStartTime) > 1 {
		return errs.NewErrorWithKind(errs.ErrInvalidTimeForBooking, errs.KindRequest)
	}

	schedule, err := bs.getSchedule(ctx, coach.ID, uint8(bookingStartTime.Weekday()))
	if err != nil {
		return err
	}

	if err := bs.checkSchedule(schedule, offer, bookingStartTime); err != nil {
		return err
	}

	startDate := time.Date(
		bookingStartTime.Year(),
		bookingStartTime.Month(),
		bookingStartTime.Day(),
		0, 0, 0, 0, mskLocation,
	)
	endDate := startDate.Add(24 * time.Hour)

	allBookings, err := bs.listBookingsInRange(ctx, offer.AuthorID, startDate, endDate)
	if err != nil {
		return err
	}

	startCurrent := bookingStartTime.Minute() + bookingStartTime.Hour()*60
	endCurrent := startCurrent + int(offer.SlotDurationMinutes) + int(offer.SlotIntervalMinutes)

	for _, existing := range allBookings {
		if existing.ID() == booking.ID() || existing.IsCancelled() {
			continue
		}

		otherTime := existing.StartTime().In(mskLocation)
		startOther := otherTime.Hour()*60 + otherTime.Minute()
		endOther := startOther + int(existing.OfferDurationMinutes()) + int(existing.OfferIntervalMinutes())

		if intervalsOverlap(
			int64(startCurrent),
			int64(endCurrent),
			int64(startOther),
			int64(endOther),
		) {
			return errs.NewErrorWithKind(errs.ErrInvalidTimeForBooking, errs.KindRequest)
		}
	}

	oldTime := booking.StartTime()
	if err := booking.ChangeStartTime(bookingStartTime); err != nil {
		return errs.NewErrorWithKind(err, errs.KindRequest)
	}

	client, err := bs.authService.GetUserByID(ctx, booking.UserID())
	clientName := "Клиент"
	if err == nil {
		clientName = client.Firstname + " " + client.Lastname
	} else {
		bs.log.Error("failed to get client for notifyUpdated",
			slog.String("error", err.Error()),
			slog.String("user_id", booking.UserID().String()),
		)
	}

	go bs.notifyUpdated(
		context.Background(),
		coach,
		offer,
		booking,
		oldTime,
		clientName,
	)

	return nil
}

func (bs *BookingService) listBookingsInRange(ctx context.Context, authorID uuid.UUID, dayStart, dayEnd time.Time) ([]*booking.Booking, error) {
	bookings, err := bs.bookingPersistence.ListAllPerTime(ctx, authorID, dayStart, dayEnd)
	if err != nil {
		bs.log.Error("failed to list all per time",
			slog.String("error", err.Error()),
			slog.String("author_id", authorID.String()),
		)

		return nil, handleRepoError(err)
	}

	return bookings, nil
}

func (bs *BookingService) listByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*booking.Booking, int64, error) {
	list, totalCount, err := bs.bookingPersistence.ListByUserID(ctx, userID, page, limit)
	if err != nil {
		bs.log.Error("failed to list booking by user's id",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
		)

		return nil, 0, handleRepoError(err)
	}

	return list, totalCount, nil
}

func (bs *BookingService) lookupUsers(ctx context.Context, list []*booking.Booking) (*responses.LookupUsers, error) {
	if len(list) == 0 {
		return &responses.LookupUsers{}, nil
	}

	req := requests.Lookup{
		IDs: make([]uuid.UUID, len(list)),
	}

	for i, booking := range list {
		req.IDs[i] = booking.UserID()
	}

	resp, err := bs.authService.Lookup(ctx, &req)
	if err != nil {
		bs.log.Error("failed to lookup users",
			slog.Int("len_of_ids", len(list)),
			slog.String("error", err.Error()),
		)

		return nil, err
	}

	return resp, nil
}

func (bs *BookingService) storeBookingCache(ctx context.Context, booking *booking.Booking) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := bs.bookingCache.Save(ctxTimeout, booking); err != nil {
		bs.log.Error("failed to save booking into cache",
			slog.String("error", err.Error()),
		)
	}
}

func (bs *BookingService) getBookingCache(ctx context.Context, id uuid.UUID) (*booking.Booking, bool) {
	cachedBooking, err := bs.bookingCache.GetByID(ctx, id)
	if err != nil {
		if !errors.Is(err, booking.ErrBookingNotFound) {
			bs.log.Error("failed to get booking from cache",
				slog.String("error", err.Error()),
				slog.String("booking_id", id.String()),
			)
		}

		return nil, false
	}

	bs.log.Debug("booking has been taken from cache",
		slog.String("booking_id", id.String()),
	)

	return cachedBooking, true
}

func (bs *BookingService) getBookingPersistence(ctx context.Context, id uuid.UUID) (*booking.Booking, error) {
	booking, err := bs.bookingPersistence.GetByID(ctx, id)
	if err != nil {
		bs.log.Error("failed to get booking persistence",
			slog.String("error", err.Error()),
			slog.String("booking_id", id.String()),
		)

		return nil, handleRepoError(err)
	}

	return booking, nil
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

func (bs *BookingService) fetchOffer(ctx context.Context, offerID uuid.UUID) (*responses.Offer, error) {
	return bs.offerService.GetByID(ctx, offerID)
}

func (bs *BookingService) upsertClient(ctx context.Context, offerID uuid.UUID, visit requests.Time) error {
	_, err := bs.clientService.Upsert(ctx, &requests.UpsertClient{
		OfferID: offerID,
		Visit:   visit,
	})

	return err
}

func (bs *BookingService) storeBookingPersistence(ctx context.Context, userID uuid.UUID, comment string, offer *responses.Offer, bookingStartTime time.Time) (*booking.Booking, error) {
	startDate := time.Date(
		bookingStartTime.Year(),
		bookingStartTime.Month(),
		bookingStartTime.Day(),
		0, 0, 0, 0, mskLocation,
	)
	endDate := startDate.Add(24 * time.Hour)

	createFunc := func(ctx context.Context, bookings []*booking.Booking) (*booking.Booking, error) {
		startCurrent := bookingStartTime.Minute() + bookingStartTime.Hour()*60
		endCurrent := startCurrent + int(offer.SlotDurationMinutes) + int(offer.SlotIntervalMinutes)

		if bs.hasBooking(userID, bookings) {
			return nil, errs.NewErrorWithKind(booking.ErrTooMuchBookings, errs.KindRequest)
		}

		for _, existing := range bookings {
			if existing.IsCancelled() {
				continue
			}

			otherTime := existing.StartTime().In(mskLocation)
			startOther := otherTime.Hour()*60 + otherTime.Minute()
			endOther := startOther + int(existing.OfferDurationMinutes()) + int(existing.OfferIntervalMinutes())

			if intervalsOverlap(
				int64(startCurrent),
				int64(endCurrent),
				int64(startOther),
				int64(endOther),
			) {
				return nil, errs.ErrInvalidTimeForBooking
			}
		}

		return booking.New(
			offer.ID,
			userID,
			comment,
			bookingStartTime,
		)
	}

	booking, err := bs.bookingPersistence.Save(ctx, offer.AuthorID, offer.ID, startDate, endDate, createFunc)
	if err != nil {
		bs.log.Error("failed to store booking persistence",
			slog.String("error", err.Error()),
		)

		return nil, handleRepoError(err)
	}

	return booking, nil
}

func (bs *BookingService) hasBooking(userID uuid.UUID, existingBookings []*booking.Booking) bool {
	for _, booking := range existingBookings {
		if booking.UserID() == userID && !booking.IsCancelled() {
			return true
		}
	}

	return false
}

func (bs *BookingService) checkSchedule(schedule *responses.Schedule, offer *responses.Offer, bookingStartTime time.Time) error {
	startTimeMinutes := bookingStartTime.Minute() + bookingStartTime.Hour()*60
	endTimeMinutes := startTimeMinutes + int(offer.SlotDurationMinutes)

	for _, br := range schedule.Breaks {
		breakStart := int64(br.StartTime.Hour)*60 + int64(br.StartTime.Minute)
		breakEnd := int64(br.EndTime.Hour)*60 + int64(br.EndTime.Minute)

		if intervalsOverlap(
			int64(startTimeMinutes),
			int64(endTimeMinutes),
			breakStart,
			breakEnd,
		) {
			return errs.NewErrorWithKind(errs.ErrInvalidTimeForBooking, errs.KindRequest)
		}
	}

	scheduleStart := int64(schedule.StartTime.Hour)*60 + int64(schedule.StartTime.Minute)
	scheduleEnd := int64(schedule.EndTime.Hour)*60 + int64(schedule.EndTime.Minute)

	if scheduleStart > int64(startTimeMinutes) || scheduleEnd < int64(endTimeMinutes) {
		return errs.NewErrorWithKind(errs.ErrInvalidTimeForBooking, errs.KindRequest)
	}

	return nil
}

func (bs *BookingService) getSchedule(ctx context.Context, userID uuid.UUID, dayOfWeek uint8) (*responses.Schedule, error) {
	schedule, err := bs.scheduleService.GetScheduleByUserID(ctx, userID, dayOfWeek)
	if err != nil {
		return nil, err
	}

	return schedule, nil
}

func (bs *BookingService) getOfferAndCoach(ctx context.Context, offerID uuid.UUID) (*responses.Offer, *responses.User, error) {
	offer, err := bs.offerService.GetByID(ctx, offerID)
	if err != nil {
		return nil, nil, err
	}

	coach, err := bs.authService.GetUserByID(ctx, offer.AuthorID)
	if err != nil {
		return nil, nil, err
	}

	return offer, coach, nil
}

func (bs *BookingService) parseStartTime(offer *responses.Offer, startTime *requests.Time) (time.Time, error) {
	dateTimeStr := fmt.Sprintf("%s %02d:%02d:00", startTime.Date, startTime.StartTime.Hour, startTime.StartTime.Minute)
	bookingStart, err := time.ParseInLocation(time.DateTime, dateTimeStr, mskLocation)
	if err != nil {
		return time.Time{}, errs.NewErrorWithKind(errs.ErrInvalidTimeForBooking, errs.KindRequest)
	}

	now := time.Now().In(mskLocation)
	if now.Compare(bookingStart) > 0 {
		return time.Time{}, errs.NewErrorWithKind(errs.ErrInvalidTimeForBooking, errs.KindRequest)
	}
	endTime := int64(startTime.StartTime.Hour)*60 + int64(startTime.StartTime.Minute)
	endTime += int64(offer.SlotDurationMinutes)

	if endTime > 23*60+59 {
		return time.Time{}, errs.NewErrorWithKind(errs.ErrInvalidTimeForBooking, errs.KindRequest)
	}

	return time.Date(
		bookingStart.Year(),
		bookingStart.Month(),
		bookingStart.Day(),
		int(startTime.StartTime.Hour),
		int(startTime.StartTime.Minute),
		0, 0, mskLocation,
	), nil
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

func (bs *BookingService) validatePagination(page, limit uint32) error {
	if page < 1 || page > 1000 {
		return errs.NewErrorWithKind(errs.ErrInvalidPage, errs.KindRequest)
	}

	if limit < 1 || limit > 100 {
		return errs.NewErrorWithKind(errs.ErrInvalidLimit, errs.KindRequest)
	}

	return nil
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

func handleRepoError(err error) error {
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

func intervalsOverlap(start1, end1, start2, end2 int64) bool {
	return start1 < end2 && start2 < end1
}
