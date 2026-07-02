package tgbindingredis

import (
	"encoding/json"

	"github.com/devathh/hive/backend/internal/modules/auth/domain/tgbinding"
)

func tgBindingToBytes(tgbinding *tgbinding.TgBinding) ([]byte, error) {
	return json.Marshal(&tgBindingRecord{
		ID:        tgbinding.ID(),
		UserID:    tgbinding.UserID(),
		BotURL:    tgbinding.BotURL().String(),
		CreatedAt: tgbinding.CreatedAt(),
	})
}

func bytesToTgBinding(bytes []byte) (*tgbinding.TgBinding, error) {
	var record tgBindingRecord
	if err := json.Unmarshal(bytes, &record); err != nil {
		return nil, err
	}

	return tgbinding.From(
		record.ID,
		record.UserID,
		tgbinding.BotURL(record.BotURL),
		record.CreatedAt,
	), nil
}
