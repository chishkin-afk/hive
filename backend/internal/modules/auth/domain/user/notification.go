package user

// Notification is a value object of User model
type Notification struct {
	telegramEnable bool
}

func NewNotification(telegramEnable bool) Notification {
	return Notification{
		telegramEnable: telegramEnable,
	}
}

func (n Notification) IsTelegramEnable() bool {
	return n.telegramEnable
}
