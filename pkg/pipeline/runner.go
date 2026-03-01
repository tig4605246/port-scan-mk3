package pipeline

import (
	"context"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/speedctrl"
	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

type DispatchFn func(context.Context, task.Task) error

type Options struct {
	Workers    int
	Delay      time.Duration
	Dispatch   DispatchFn
	Controller *speedctrl.Controller
}

type Runner struct {
	dispatch DispatchFn
	ctrl     *speedctrl.Controller
	delay    time.Duration
}

func NewRunner(opts Options) *Runner {
	delay := opts.Delay
	if delay == 0 {
		delay = 10 * time.Millisecond
	}
	return &Runner{
		dispatch: opts.Dispatch,
		ctrl:     opts.Controller,
		delay:    delay,
	}
}

func (r *Runner) waitGate(ctx context.Context) error {
	for {
		gate := r.ctrl.Gate()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-gate:
			return nil
		}
	}
}

func (r *Runner) Run(ctx context.Context, chunks []task.Chunk) error {
	for _, ch := range chunks {
		for i := ch.NextIndex; i < ch.TotalCount; i++ {
			if err := r.waitGate(ctx); err != nil {
				return err
			}
			if r.dispatch != nil {
				if err := r.dispatch(ctx, task.Task{ChunkCIDR: ch.CIDR, Index: i}); err != nil {
					return err
				}
			}
			time.Sleep(r.delay)
		}
	}
	return nil
}
