package telegram

type ChatID int64

type ChatType string

const (
	ChatTypePrivate ChatType = "private"
)

type MessageID int

type MessageEntityType string

const (
	MessageEntityTypeBotCommand MessageEntityType = "bot_command"
)

type Text string

type UserID int64

type Username string
