package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/devem-tech/telegram-bot-platform/pkg/opt"
	"github.com/devem-tech/telegram-bot-platform/pkg/slices"
)

func In(x tgbotapi.Update) Update {
	return Update{
		Message: opt.Nilable(x.Message, func() *Message {
			return &Message{
				MessageID: MessageID(x.Message.MessageID),
				From: opt.Nilable(x.Message.From, func() User {
					return User{
						ID:        UserID(x.Message.From.ID),
						IsBot:     x.Message.From.IsBot,
						FirstName: x.Message.From.FirstName,
						LastName:  x.Message.From.LastName,
						Username:  Username(x.Message.From.UserName),
					}
				}),
				Date: x.Message.Date,
				Chat: opt.Nilable(x.Message.Chat, func() Chat {
					return Chat{
						ID:    ChatID(x.Message.Chat.ID),
						Type:  ChatType(x.Message.Chat.Type),
						Title: x.Message.Chat.Title,
					}
				}),
				Text: Text(x.Message.Text),
				Entities: slices.Map(x.Message.Entities, func(x tgbotapi.MessageEntity) MessageEntity {
					return MessageEntity{
						Type: MessageEntityType(x.Type),
					}
				}),
			}
		}),
	}
}
