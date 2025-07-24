package telegram

type Update struct {
	Message *Message
}

type Message struct {
	MessageID MessageID
	From      User
	Date      int
	Chat      Chat
	Text      Text
	Entities  []MessageEntity
}

func (m *Message) IsCommand() bool {
	for _, entity := range m.Entities {
		if entity.IsCommand() {
			return true
		}
	}

	return false
}

type User struct {
	ID        UserID
	IsBot     bool
	FirstName string
	LastName  string
	Username  Username
}

type Chat struct {
	ID    ChatID
	Type  ChatType
	Title string
}

type MessageEntity struct {
	Type MessageEntityType
}

func (e MessageEntity) IsCommand() bool {
	return e.Type == MessageEntityTypeBotCommand
}
