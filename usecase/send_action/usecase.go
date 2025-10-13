package send_action

import (
	"context"
	"time"

	"github.com/devem-tech/telegram-bot-platform/platform"
)

const chatActionTTL = 5 * time.Second

type Usecase[U platform.Update] struct {
	platform.Usecase[U]

	Timeout            time.Duration
	OnTimeoutFunc      func(err error) error
	SendChatActionFunc func(ctx context.Context, update U)
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
				if u.SendChatActionFunc != nil {
					u.SendChatActionFunc(ctx, update)
				}

				time.Sleep(chatActionTTL)
			}
		}
	}()

	cancel := func() {}
	if u.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, u.Timeout)
	}

	defer cancel()

	ch := make(chan error, 1)

	go func() {
		ch <- u.Usecase.Handle(ctx, update)
	}()

	select {
	case <-ctx.Done():
		if u.OnTimeoutFunc != nil {
			return u.OnTimeoutFunc(ctx.Err())
		}

		return ctx.Err() //nolint:wrapcheck
	case err := <-ch:
		return err
	}
}
