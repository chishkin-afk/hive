package notification

import "errors"

var (
	ErrEmailAlreadyEnabled  = errors.New("email is already enabled")
	ErrEmailAlreadyDisabled = errors.New("email is already disabled")
	ErrTgAlreadyEnabled     = errors.New("tg is already enabled")
	ErrTgAlreadyDisabled    = errors.New("tg is already disabled")
	ErrNoChatID             = errors.New("can't enable tg without chat id")
)

type Notification struct {
	tgEnabled    bool
	emailEnabled bool
	tgChatID     int64
}

func New() *Notification {
	return &Notification{
		tgEnabled:    false,
		emailEnabled: false,
		tgChatID:     0,
	}
}

func Restore(
	tgEnabled bool,
	emailEnabled bool,
	tgChatID int64,
) *Notification {
	return &Notification{
		tgEnabled:    tgEnabled,
		emailEnabled: emailEnabled,
		tgChatID:     tgChatID,
	}
}

func (n *Notification) SetChatID(chatID int64) {
	if chatID == 0 {
		n.tgEnabled = false
	}

	n.tgChatID = chatID
}

func (n *Notification) EnableTg() error {
	if n.tgEnabled {
		return ErrTgAlreadyEnabled
	}
	if n.tgChatID == 0 {
		return ErrNoChatID
	}

	n.tgEnabled = true
	return nil
}

func (n *Notification) DisableTg() error {
	if !n.tgEnabled {
		return ErrTgAlreadyDisabled
	}

	n.tgEnabled = false
	return nil
}

func (n *Notification) EnableEmail() error {
	if n.emailEnabled {
		return ErrEmailAlreadyEnabled
	}

	n.emailEnabled = true
	return nil
}

func (n *Notification) DisableEmail() error {
	if !n.emailEnabled {
		return ErrEmailAlreadyDisabled
	}

	n.emailEnabled = false
	return nil
}

func (n *Notification) TgEnabled() bool {
	return n.tgEnabled
}

func (n *Notification) EmailEnabled() bool {
	return n.emailEnabled
}

func (n *Notification) TgChatID() int64 {
	return n.tgChatID
}
