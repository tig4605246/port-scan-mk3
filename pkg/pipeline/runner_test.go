package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/speedctrl"
	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

func TestRunner_WhenControllerPaused_DoesNotDispatchTasks(t *testing.T) {
	ctrl := speedctrl.NewController(speedctrl.WithAPIEnabled(false))
	ctrl.SetManualPaused(true)

	dispatched := 0
	r := NewRunner(Options{
		Workers: 1,
		Dispatch: func(context.Context, task.Task) error {
			dispatched++
			return nil
		},
		Controller: ctrl,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = r.Run(ctx, []task.Chunk{{CIDR: "10.0.0.0/30", TotalCount: 4, Status: "pending"}})
	if dispatched != 0 {
		t.Fatalf("expected zero dispatched while paused, got %d", dispatched)
	}
}
