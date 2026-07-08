package bookingservices

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/devathh/hive/backend/pkg/errs"
	"github.com/google/uuid"
)

const (
	NewBooking = `
📅 Новая запись:

Услуга: %s
Дата: %d %s
Время: %d:%d %s

ID: %s
`
	UpdateBooking = `
📍 Запись была перенесена:

Услуга: %s
Старое время: %d %s %d:%d %s
Новое время: %d %s %d:%d %s

ID: %s
`

	DeleteBooking = `
🗑 Запись была отменена:

Услуга: %s
Дата: %d %s
Время: %d:%d %s

ID: %s
`
)

type baseInfo struct {
	title string
	id    uuid.UUID
}

func (bs *BookingService) bookingCreatedMsg(bi *baseInfo, startTime time.Time) string {
	startTime, zone := bs.convertDate(startTime)
	return fmt.Sprintf(NewBooking,
		bi.title,
		startTime.Day(),
		bs.getMonthRU(startTime.Month().String()),
		startTime.Hour(),
		startTime.Minute(),
		zone,
		bi.id.String(),
	)
}

func (bs *BookingService) bookingUpdatedMsg(bi *baseInfo, newTime, oldTime time.Time) string {
	newTime, newZone := bs.convertDate(newTime)
	oldTime, oldZone := bs.convertDate(oldTime)

	return fmt.Sprintf(UpdateBooking,
		bi.title,
		oldTime.Day(),
		bs.getMonthRU(oldTime.Month().String()),
		oldTime.Hour(),
		oldTime.Minute(),
		oldZone,
		newTime.Day(),
		bs.getMonthRU(newTime.Month().String()),
		newTime.Hour(),
		newTime.Minute(),
		newZone,
		bi.id.String(),
	)
}

func (bs *BookingService) bookingCanceledMsg(bi *baseInfo, startTime time.Time) string {
	startTime, zone := bs.convertDate(startTime)
	return fmt.Sprintf(DeleteBooking,
		bi.title,
		startTime.Day(),
		bs.getMonthRU(startTime.Month().String()),
		startTime.Hour(),
		startTime.Minute(),
		zone,
		bi.id.String(),
	)
}

func (bs *BookingService) notifyUser(ctx context.Context, user *responses.User, msg string) error {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return errs.ErrEmptyMsg
	}

	if !user.Notification.TelegramEnable {
		bs.log.Debug("user's telegram is disabled", slog.String("user_id", user.ID.String()))
		return nil
	}

	return bs.engageClient.SendMessage(ctx, user.TgChatID, msg)
}
