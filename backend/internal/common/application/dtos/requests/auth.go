package requests

import "github.com/google/uuid"

type PrepareUser struct {
	Firstname   string `json:"firstname"`
	Lastname    string `json:"lastname"` // can be empty
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"` // can be empty
	Password    string `json:"password"`     // can be empty
}

type ConfirmUser struct {
	Email string `json:"email"`
	Code  int    `json:"code"`
}

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUser struct {
	Firstname    *string            `json:"firstname"`
	Lastname     *string            `json:"lastname"`
	PhoneNumber  *string            `json:"phone_number"`
	Password     *string            `json:"password"`
	OnVacation   *bool              `json:"on_vacation"`
	Notification UpdateNotification `json:"notification"`
}

type UpdateNotification struct {
	TelegramEnable *bool `json:"telegram_enable"`
}

type UpdateTgChatID struct {
	TgChatID int64 `json:"tg_chat_id"`
}

type BindTG struct {
	BindingID uuid.UUID `json:"binding_id"`
	TgChatID  int64     `json:"tg_chat_id"`
}

type ResendCode struct {
	Email string `json:"email"`
}

type Lookup struct {
	IDs []uuid.UUID `json:"ids"`
}

type ForgetPassword struct {
	Email string `json:"email"`
}

type ChangePassword struct {
	ResetToken  string `json:"reset_token"`
	NewPassword string `json:"new_password"`
}
