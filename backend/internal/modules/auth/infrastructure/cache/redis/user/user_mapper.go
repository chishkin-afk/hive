package userredis

import (
	"encoding/json"

	"github.com/devathh/hive/backend/internal/modules/auth/domain/user"
)

func domainToRecord(domain *user.User) ([]byte, error) {
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

func recordToDomain(bytes []byte) (*user.User, error) {
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

func domainToPrepared(domain *user.User, generatedPassword string) ([]byte, error) {
	return json.Marshal(preparedUserRecord{
		User: userRecord{
			ID:             domain.ID(),
			Firstname:      domain.Firstname(),
			Lastname:       domain.Lastname(),
			ConfirmedEmail: domain.ConfirmedEmail().String(),
			PasswordHash:   domain.PasswordHash().String(),
			PaidUntil:      domain.PaidUntil(),
			OnVacation:     domain.OnVacation(),
			TgChatID:       domain.TgChatID(),
			Notification: notification{
				TelegramEnable: domain.Notification().IsTelegramEnable(),
			},
			CreatedAt: domain.CreatedAt(),
			UpdatedAt: domain.UpdatedAt(),
		},
		GeneratedPassword: generatedPassword,
	})
}

func preparedToDomain(bytes []byte) (*user.User, string, error) {
	var prepared preparedUserRecord
	if err := json.Unmarshal(bytes, &prepared); err != nil {
		return nil, "", err
	}

	return user.From(
		prepared.User.ID,
		prepared.User.Firstname,
		prepared.User.Lastname,
		user.Email(prepared.User.ConfirmedEmail),
		user.PasswordHash(prepared.User.PasswordHash),
		prepared.User.PaidUntil,
		prepared.User.OnVacation,
		prepared.User.TgChatID,
		user.NewNotification(prepared.User.Notification.TelegramEnable),
		prepared.User.CreatedAt,
		prepared.User.UpdatedAt,
	), prepared.GeneratedPassword, nil
}
