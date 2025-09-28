package telegram

import (
	"context"
	"fmt"
	"io"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/devem-tech/telegram-bot-platform/pkg/opt"
)

type ParseMode string

const (
	ParseModeMarkdownV2 ParseMode = tgbotapi.ModeMarkdownV2
	ParseModeModeHTML   ParseMode = tgbotapi.ModeHTML
)

type ChatAction string

const (
	ChatActionTyping      ChatAction = tgbotapi.ChatTyping
	ChatActionUploadPhoto ChatAction = tgbotapi.ChatUploadPhoto
	ChatActionUploadVideo ChatAction = tgbotapi.ChatUploadVideo
)

type InputFile struct {
	Name string
	Data io.Reader
}

type Client struct {
	client *tgbotapi.BotAPI
}

func New(
	token string,
) *Client {
	client, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	return &Client{
		client: client,
	}
}

type UpdateCh = <-chan tgbotapi.Update

func (c *Client) Updates() UpdateCh {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return c.client.GetUpdatesChan(u)
}

func (c *Client) Me(_ context.Context) Username {
	return Username(c.client.Self.UserName)
}

type SendMessageIn struct {
	ChatID           ChatID
	Text             Text
	ReplyToMessageID *MessageID
	ParseMode        *ParseMode
}

func (c *Client) SendMessage(_ context.Context, in SendMessageIn) error {
	message := tgbotapi.NewMessage(int64(in.ChatID), string(in.Text))
	message.ReplyToMessageID = opt.Nilable(in.ReplyToMessageID, func() int {
		return int(*in.ReplyToMessageID)
	})
	message.ParseMode = opt.Nilable(in.ParseMode, func() string {
		return string(*in.ParseMode)
	})

	if _, err := c.client.Send(message); err != nil {
		return fmt.Errorf("telegram: %w", err)
	}

	return nil
}

type SendChatActionIn struct {
	ChatID ChatID
	Action ChatAction
}

func (c *Client) SendChatAction(_ context.Context, in SendChatActionIn) error {
	chatAction := tgbotapi.NewChatAction(int64(in.ChatID), string(in.Action))

	if _, err := c.client.Send(chatAction); err != nil {
		return fmt.Errorf("telegram: %w", err)
	}

	return nil
}

type SendVideoIn struct {
	ChatID              ChatID
	Video               InputFile
	DisableNotification bool
	Caption             string
	ParseMode           ParseMode
	SupportsStreaming   bool
}

func (c *Client) SendVideo(_ context.Context, in SendVideoIn) error {
	video := tgbotapi.NewVideo(int64(in.ChatID), tgbotapi.FileReader{
		Name:   in.Video.Name,
		Reader: in.Video.Data,
	})

	video.DisableNotification = in.DisableNotification
	video.Caption = in.Caption
	video.ParseMode = string(in.ParseMode)
	video.SupportsStreaming = in.SupportsStreaming

	if _, err := c.client.Send(video); err != nil {
		return fmt.Errorf("telegram: %w", err)
	}

	return nil
}

type DeleteMessageIn struct {
	ChatID    ChatID
	MessageID MessageID
}

func (c *Client) DeleteMessage(_ context.Context, in DeleteMessageIn) error {
	deleteMessage := tgbotapi.NewDeleteMessage(int64(in.ChatID), int(in.MessageID))

	if _, err := c.client.Send(deleteMessage); err != nil {
		return fmt.Errorf("telegram: %w", err)
	}

	return nil
}

type SendMediaGroupIn struct {
	ChatID ChatID
	Media  []InputFile
}

func (c *Client) SendMediaGroup(_ context.Context, in SendMediaGroupIn) error {
	files := make([]any, len(in.Media))

	for i, media := range in.Media {
		// Note: currently, sending only photos is supported.
		files[i] = tgbotapi.NewInputMediaPhoto(tgbotapi.FileReader{
			Name:   media.Name,
			Reader: media.Data,
		})
	}

	mediaGroup := tgbotapi.NewMediaGroup(int64(in.ChatID), files)

	if _, err := c.client.Send(mediaGroup); err != nil {
		if err.Error() == "json: cannot unmarshal array into Go value of type tgbotapi.Message" {
			return nil
		}

		return fmt.Errorf("telegram: %w", err)
	}

	return nil
}
