package engagedomain

type User struct {
	ID        int64
	FirstName string
	Username  string
}

type Chat struct {
	ID      int64
	Type    string
	Title   string
	IsForum bool
}

type ForumTopic struct {
	ThreadID int
	Name     string
}

type Message struct {
	MessageID       int
	MessageThreadID int
	From            User
	Chat            Chat
	Text            string
}

type InlineKeyboardButton struct {
	Text         string
	Style        string
	URL          string
	CallbackData string
}
