package userpg

import (
	"encoding/json"

	"github.com/devathh/hive/backend/internal/modules/auth/domain/user"
)

func toRecord(domain *user.User) (*userRecord, error) {
	bytes, err := json.Marshal(notification{
		TelegramEnable: domain.Notification().IsTelegramEnable(),
	})
	if err != nil {
		return nil, err
	}

	return &userRecord{
		ID:             domain.ID(),
		Firstname:      domain.Firstname(),
		Lastname:       domain.Lastname(),
		ConfirmedEmail: domain.ConfirmedEmail().String(),
		PasswordHash:   domain.PasswordHash().String(),
		PaidUntil:      domain.PaidUntil(),
		OnVacation:     domain.OnVacation(),
		TgChatID:       domain.TgChatID(),
		Notification:   bytes,
		CreatedAt:      domain.CreatedAt(),
		UpdatedAt:      domain.UpdatedAt(),
	}, nil
}

func toDomain(record *userRecord) (*user.User, error) {
	var notification notification
	if err := json.Unmarshal(record.Notification, &notification); err != nil {
		return nil, err
	}

	return user.From(
		record.ID,
		record.Firstname,
		record.Lastname,
		user.Email(record.ConfirmedEmail),
		user.PasswordHash(record.PasswordHash),
		record.PaidUntil,
		record.OnVacation,
		record.TgChatID,
		user.NewNotification(notification.TelegramEnable),
		record.CreatedAt,
		record.UpdatedAt,
	), nil
}
