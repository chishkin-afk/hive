package bookingservices

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/devathh/hive/backend/internal/modules/bookings/domain/booking"
	"github.com/google/uuid"
)

const (
	createdNotification = `
📅 Новая запись:

Услуга: %s
Дата: %d %s
Время: %s %s

ID: %s	
`

	updatedNotification = `
📍 Запись была перенесена:

Услуга: %s
Старое время: %d %s %s %s
Новая время: %d %s %s %s

ID: %s
`

	canceledNotification = `
🗑 Запись была отменена:

Услуга: %s
Дата: %d %s
Время: %s %s

ID: %s	
`
)

func (bs *BookingService) notifyHolder(ctx context.Context, msg string, user *responses.User) error {
	if !user.Notification.TelegramEnable {
		bs.log.Warn("tg's notification of user is disabled",
			slog.String("user_id", user.ID.String()),
		)

		return nil
	}

	if err := bs.engageClient.SendMessage(ctx, user.TgChatID, msg); err != nil {
		return err
	}

	return nil
}

func (bs *BookingService) notifyCreated(ctx context.Context, coach *responses.User, offer *responses.Offer, booking *booking.Booking) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	startTime := booking.StartTime().In(mskLocation)

	msg := fmt.Sprintf(createdNotification,
		offer.Title,
		startTime.Day(),
		getMonthRU(startTime.Month()),
		getTimes(startTime),
		"MSK",
		booking.ID().String(),
	)

	bs.notify(
		ctxTimeout,
		coach,
		booking.UserID(),
		msg,
	)
}

func (bs *BookingService) notifyUpdated(
	ctx context.Context,
	coach *responses.User,
	offer *responses.Offer,
	booking *booking.Booking,
	oldTime time.Time,
) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	startTime := booking.StartTime().In(mskLocation)
	oldTime = oldTime.In(mskLocation)

	msg := fmt.Sprintf(updatedNotification,
		offer.Title,
		oldTime.Day(),
		getMonthRU(oldTime.Month()),
		getTimes(oldTime),
		"MSK",
		startTime.Day(),
		getMonthRU(startTime.Month()),
		getTimes(startTime),
		"MSK",
		booking.ID().String(),
	)

	bs.notify(
		ctxTimeout,
		coach,
		booking.UserID(),
		msg,
	)
}

func (bs *BookingService) notifyCanceled(ctx context.Context, coach *responses.User, offer *responses.Offer, booking *booking.Booking) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	startTime := booking.StartTime().In(mskLocation)

	msg := fmt.Sprintf(canceledNotification,
		offer.Title,
		startTime.Day(),
		getMonthRU(startTime.Month()),
		getTimes(startTime),
		"MSK",
		booking.ID().String(),
	)

	bs.notify(
		ctxTimeout,
		coach,
		booking.UserID(),
		msg,
	)
}

func (bs *BookingService) notify(
	ctx context.Context,
	coach *responses.User,
	clientID uuid.UUID,
	msg string,
) {
	if err := bs.notifyHolder(ctx, msg, coach); err != nil {
		bs.log.Error("failed to notify coach",
			slog.String("error", err.Error()),
			slog.String("user_id", coach.ID.String()),
		)
	}

	if coach.ID == clientID {
		return
	}

	client, err := bs.authService.GetUserByID(ctx, clientID)
	if err != nil {
		bs.log.Error("failed to get client from auth",
			slog.String("error", err.Error()),
			slog.String("client_id", clientID.String()),
		)

		return
	}

	if err := bs.notifyHolder(ctx, msg, client); err != nil {
		bs.log.Error("failed to notify client",
			slog.String("error", err.Error()),
			slog.String("user_id", clientID.String()),
		)
	}
}

func getMonthRU(month time.Month) string {
	switch month {
	case time.January:
		return "Января"
	case time.February:
		return "Февраля"
	case time.March:
		return "Марта"
	case time.April:
		return "Апреля"
	case time.May:
		return "Мая"
	case time.June:
		return "Июня"
	case time.July:
		return "Июля"
	case time.August:
		return "Августа"
	case time.September:
		return "Сентября"
	case time.October:
		return "Октября"
	case time.November:
		return "Ноября"
	case time.December:
		return "Декабря"
	}

	return month.String()
}

func getTimes(t time.Time) string {
	return fmt.Sprintf("%.2d:%.2d",
		t.Hour(),
		t.Minute(),
	)
}
