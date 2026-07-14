package userredis

import (
	"encoding/json"

	"github.com/devathh/hive/backend/internal/modules/auth/domain/user"
)

func domainToBytes(domain *user.User) ([]byte, error) {
	return json.Marshal(userRecord{
		ID:             domain.ID(),
		Firstname:      domain.Firstname(),
		Lastname:       domain.Lastname(),
		ConfirmedEmail: domain.ConfirmedEmail().String(),
		PhoneNumber:    domain.PhoneNumber().String(),
		PaidUntil:      domain.PaidUntil(),
		OnVacation:     domain.OnVacation(),
		TgChatID:       domain.TgChatID(),
		Notification: notification{
			TelegramEnable: domain.Notification().IsTelegramEnable(),
		},
		TrialUsed: domain.TrialUsed(),
		CreatedAt: domain.CreatedAt(),
		UpdatedAt: domain.UpdatedAt(),
	})
}

func bytesToDomain(bytes []byte) (*user.User, error) {
	var record userRecord
	if err := json.Unmarshal(bytes, &record); err != nil {
		return nil, err
	}

	return user.From(
		record.ID,
		record.Firstname,
		record.Lastname,
		user.Email(record.ConfirmedEmail),
		user.PhoneNumber(record.PhoneNumber),
		"",
		record.PaidUntil,
		record.OnVacation,
		record.TgChatID,
		user.NewNotification(record.Notification.TelegramEnable),
		record.TrialUsed,
		record.CreatedAt,
		record.UpdatedAt,
	), nil
}
