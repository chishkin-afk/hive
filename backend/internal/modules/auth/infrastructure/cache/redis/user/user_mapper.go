package userredis

import (
	"encoding/json"

	"github.com/devathh/hive/backend/internal/modules/auth/domain/user"
)

func toBytes(domain *user.User) ([]byte, error) {
	return json.Marshal(userRecord{
		ID:             domain.ID(),
		Firstname:      domain.Firstname(),
		Lastname:       domain.Lastname(),
		ConfirmedEmail: domain.ConfirmedEmail().String(),
		PaidUntil:      domain.PaidUntil(),
		OnVacation:     domain.OnVacation(),
		TgChatID:       domain.TgChatID(),
		Notification: notification{
			TelegramEnable: domain.Notification().IsTelegramEnable(),
		},
		CreatedAt: domain.CreatedAt(),
		UpdatedAt: domain.UpdatedAt(),
	})
}

func toDomain(bytes []byte) (*user.User, error) {
	var record userRecord
	if err := json.Unmarshal(bytes, &record); err != nil {
		return nil, err
	}

	return user.From(
		record.ID,
		record.Firstname,
		record.Lastname,
		user.Email(record.ConfirmedEmail),
		"",
		record.PaidUntil,
		record.OnVacation,
		record.TgChatID,
		user.NewNotification(record.Notification.TelegramEnable),
		record.CreatedAt,
		record.UpdatedAt,
	), nil
}
