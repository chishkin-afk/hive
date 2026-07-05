package preparedredis

import (
	"encoding/json"
	"time"

	"github.com/devathh/hive/backend/internal/modules/auth/domain/user"
)

func domainToPreparedBytes(domain *user.User, generatedPassword string, code int) ([]byte, error) {
	return json.Marshal(preparedUserRecord{
		User: userRecord{
			ID:             domain.ID(),
			Firstname:      domain.Firstname(),
			Lastname:       domain.Lastname(),
			ConfirmedEmail: domain.ConfirmedEmail().String(),
			PhoneNumber:    domain.PhoneNumber().String(),
			PasswordHash:   domain.PasswordHash().String(),
			PaidUntil:      domain.PaidUntil(),
			OnVacation:     domain.OnVacation(),
			TgChatID:       domain.TgChatID(),
			Notification: notification{
				TelegramEnable: domain.Notification().IsTelegramEnable(),
			},
			TrialUsed: domain.TrialUsed(),
			CreatedAt: domain.CreatedAt(),
			UpdatedAt: domain.UpdatedAt(),
		},
		Code:              code,
		GeneratedPassword: generatedPassword,
	})
}

func bytesToPrepared(bytes []byte) (*preparedUserRecord, error) {
	var prepared preparedUserRecord
	if err := json.Unmarshal(bytes, &prepared); err != nil {
		return nil, err
	}

	return &prepared, nil
}

func preparedToUser(prepared *preparedUserRecord) *user.User {
	return user.From(
		prepared.User.ID,
		prepared.User.Firstname,
		prepared.User.Lastname,
		user.Email(prepared.User.ConfirmedEmail),
		user.PhoneNumber(prepared.User.PhoneNumber),
		user.PasswordHash(prepared.User.PasswordHash),
		prepared.User.PaidUntil,
		prepared.User.OnVacation,
		prepared.User.TgChatID,
		user.NewNotification(prepared.User.Notification.TelegramEnable),
		prepared.User.TrialUsed,
		prepared.User.CreatedAt,
		prepared.User.UpdatedAt,
	)
}

func lastSendToBytes(at time.Time) ([]byte, error) {
	return json.Marshal(LastSend{
		At: at,
	})
}

func bytesToLastSendRecord(bytes []byte) (*LastSend, error) {
	var record LastSend
	if err := json.Unmarshal(bytes, &record); err != nil {
		return nil, err
	}

	return &record, nil
}
