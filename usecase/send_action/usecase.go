package send_action

import (
	"context"
	"fmt"
	"time"

	"github.com/devem-tech/telegram-bot-platform/platform"
	"github.com/devem-tech/telegram-bot-platform/telegram"
)

const chatActionTTL = 5 * time.Second

type Usecase struct {
	platform.Usecase

	Telegram  *telegram.Client
	Action    telegram.ChatAction
	Namespace string
	Timeout   time.Duration
}

func (u *Usecase) Handle(ctx context.Context, update telegram.Update) error {
	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			default:
				_ = u.Telegram.SendChatAction(ctx, telegram.SendChatActionIn{
					ChatID: update.Message.Chat.ID,
					Action: u.Action,
				})

				time.Sleep(chatActionTTL)
			}
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, u.Timeout)
	defer cancel()

	ch := make(chan error, 1)

	go func() {
		ch <- u.Usecase.Handle(ctx, update)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("%s: %w", u.Namespace, ctx.Err())
	case err := <-ch:
		return err
	}
}
