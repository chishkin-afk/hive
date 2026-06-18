package requests

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
	Firstname    *string             `json:"firstname"`
	Lastname     *string             `json:"lastname"`
	PhoneNumber  *string             `json:"phone_number"`
	Password     *string             `json:"password"`
	OnVacation   *bool               `json:"on_vacation"`
	Notification *UpdateNotification `json:"notification"`
}

type UpdateNotification struct {
	TelegramEnable bool `json:"telegram_enable"`
}
