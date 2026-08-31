package dtos

type SendMessageRequest struct {
	ChatID   int64  `json:"chat_id"`
	ThreadID int    `json:"thread_id"`
	Text     string `json:"text"`
}

type BindTgRequest struct {
	BindingID string `json:"binding_id"`
	TgChatID  int64  `json:"tg_chat_id"`
}
