package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/speedctrl"
	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

func TestRunner_WhenDispatchReturnsError_StopsAndReturnsError(t *testing.T) {
	ctrl := speedctrl.NewController()
	calls := 0
	r := NewRunner(Options{
		Delay:      0,
		Controller: ctrl,
		Dispatch: func(context.Context, task.Task) error {
			calls++
			if calls == 2 {
				return errors.New("boom")
			}
			return nil
		},
	})

	err := r.Run(context.Background(), []task.Chunk{{CIDR: "x", TotalCount: 3}})
	if err == nil {
		t.Fatal("expected dispatch error")
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestRunner_WhenContextCanceled_ReturnsContextError(t *testing.T) {
	ctrl := speedctrl.NewController()
	ctrl.SetManualPaused(true)
	r := NewRunner(Options{Controller: ctrl})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := r.Run(ctx, []task.Chunk{{CIDR: "x", TotalCount: 1}}); err == nil {
		t.Fatal("expected context error")
	}
}
