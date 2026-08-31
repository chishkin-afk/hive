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

// TODO: сделать уведомление пользователей по почте

const (
	createdNotification = `
📅 Новая запись:

Клиент: %s
Услуга: %s
Дата: %d %s
Время: %s %s
Комментарий: %s
`

	updatedNotification = `
📍 Запись была перенесена:

Клиент: %s
Услуга: %s
Старое время: %d %s %s %s
Новое время: %d %s %s %s
Комментарий: %s
`

	canceledNotification = `
🗑 Запись была отменена:

Клиент: %s
Услуга: %s
Дата: %d %s
Время: %s %s
Комментарий: %s
`
)

func (bs *BookingService) notifyMail(ctx context.Context, msg string, subject string, email []string) error {
	return bs.sc.Notify(ctx, msg, subject, email)
}

func (bs *BookingService) notifyMailHTML(ctx context.Context, htmlBody string, subject string, email []string) error {
	return bs.sc.NotifyHTML(ctx, htmlBody, subject, email)
}

func (bs *BookingService) buildNotificationHTML(content string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Montserrat, Arial, sans-serif; background-color: #E6E6E6; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; background-color: #FAFAFA; border-radius: 12px; box-shadow: 0 4px 16px rgba(0,0,0,0.12); }
        .header { padding: 10px 0; border-bottom: 2px solid #FFB61A; }
        .content { padding: 20px 0; }
        .content pre { font-family: Montserrat, Arial, sans-serif; font-size: 16px; color: #1A1A1A; white-space: pre-wrap; word-wrap: break-word; margin: 0; line-height: 1.8; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header"></div>
        <div class="content">
            <pre>%s</pre>
        </div>
    </div>
</body>
</html>
`, content)
}

func (bs *BookingService) notifyHolder(ctx context.Context, msg string, subject string, user *responses.User) error {
	var tgErr, mailErr error

	// Всегда отправляем на почту в HTML
	htmlBody := bs.buildNotificationHTML(msg)
	mailErr = bs.notifyMailHTML(ctx, htmlBody, subject, []string{user.Email})
	if mailErr != nil {
		bs.log.Error("failed to send email notification",
			slog.String("error", mailErr.Error()),
			slog.String("user_id", user.ID.String()),
		)
	}

	// Отправляем в Telegram, если включено
	if user.Notification.TelegramEnable {
		if err := bs.engageClient.SendMessage(ctx, user.TgChatID, msg); err != nil {
			bs.log.Error("failed to send telegram notification",
				slog.String("error", err.Error()),
				slog.String("user_id", user.ID.String()),
			)
			tgErr = err
		}
	} else {
		bs.log.Warn("tg's notification of user is disabled",
			slog.String("user_id", user.ID.String()),
		)
	}

	// Возвращаем ошибку, если хотя бы один канал не сработал
	if tgErr != nil && mailErr != nil {
		return fmt.Errorf("failed to send via both channels: tg: %w, mail: %w", tgErr, mailErr)
	}
	if tgErr != nil {
		return tgErr
	}
	if mailErr != nil {
		return mailErr
	}

	return nil
}

func (bs *BookingService) notifyCreated(ctx context.Context, coach *responses.User, offer *responses.Offer, booking *booking.Booking, clientName string) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	startTime := booking.StartTime().In(mskLocation)
	comment := booking.Comment()
	if comment == "" {
		comment = "—"
	}

	msg := fmt.Sprintf(createdNotification,
		clientName,
		offer.Title,
		startTime.Day(),
		getMonthRU(startTime.Month()),
		getTimes(startTime),
		"MSK",
		comment,
	)

	subject := fmt.Sprintf("Новая запись: %s", offer.Title)
	bs.notify(
		ctxTimeout,
		coach,
		booking.UserID(),
		msg,
		subject,
	)
}

func (bs *BookingService) notifyUpdated(
	ctx context.Context,
	coach *responses.User,
	offer *responses.Offer,
	booking *booking.Booking,
	oldTime time.Time,
	clientName string,
) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	startTime := booking.StartTime().In(mskLocation)
	oldTime = oldTime.In(mskLocation)
	comment := booking.Comment()
	if comment == "" {
		comment = "—"
	}

	msg := fmt.Sprintf(updatedNotification,
		clientName,
		offer.Title,
		oldTime.Day(),
		getMonthRU(oldTime.Month()),
		getTimes(oldTime),
		"MSK",
		startTime.Day(),
		getMonthRU(startTime.Month()),
		getTimes(startTime),
		"MSK",
		comment,
	)

	subject := fmt.Sprintf("Запись перенесена: %s", offer.Title)
	bs.notify(
		ctxTimeout,
		coach,
		booking.UserID(),
		msg,
		subject,
	)
}

func (bs *BookingService) notifyCanceled(ctx context.Context, coach *responses.User, offer *responses.Offer, booking *booking.Booking, clientName string) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	startTime := booking.StartTime().In(mskLocation)
	comment := booking.Comment()
	if comment == "" {
		comment = "—"
	}

	msg := fmt.Sprintf(canceledNotification,
		clientName,
		offer.Title,
		startTime.Day(),
		getMonthRU(startTime.Month()),
		getTimes(startTime),
		"MSK",
		comment,
	)

	subject := fmt.Sprintf("Запись отменена: %s", offer.Title)
	bs.notify(
		ctxTimeout,
		coach,
		booking.UserID(),
		msg,
		subject,
	)
}

func (bs *BookingService) notify(
	ctx context.Context,
	coach *responses.User,
	clientID uuid.UUID,
	msg string,
	subject string,
) {
	if err := bs.notifyHolder(ctx, msg, subject, coach); err != nil {
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

	if err := bs.notifyHolder(ctx, msg, subject, client); err != nil {
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
