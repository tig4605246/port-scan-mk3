package state

import (
	"context"
	"os"
	"os/signal"
)

func WithSIGINTCancel(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := signal.NotifyContext(parent, os.Interrupt)
	return ctx, cancel
}
