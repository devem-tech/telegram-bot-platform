package send_action

import (
	"context"
	"fmt"
	"time"

	"github.com/devem-tech/telegram-bot-platform/platform"
)

const chatActionTTL = 5 * time.Second

type Usecase[U platform.Update] struct {
	platform.Usecase[U]

	SendChatActionFunc func(ctx context.Context, update U)

	Namespace string
	Timeout   time.Duration
}

func (u *Usecase[U]) Handle(ctx context.Context, update U) error {
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
				u.SendChatActionFunc(ctx, update)

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
