package scanapp

import (
	"context"
	"testing"

	"github.com/xuxiping/port-scan-mk3/pkg/ratelimit"
	"github.com/xuxiping/port-scan-mk3/pkg/speedctrl"
	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

type dispatchObsEvent struct {
	kind      string
	cidr      string
	taskIndex int
}

type recordingDispatchObserver struct {
	eventsList []dispatchObsEvent
}

func (r *recordingDispatchObserver) OnGateWaitStart(cidr string, taskIndex int) {
	r.eventsList = append(r.eventsList, dispatchObsEvent{kind: "gate_wait_start", cidr: cidr, taskIndex: taskIndex})
}

func (r *recordingDispatchObserver) OnGateReleased(cidr string, taskIndex int) {
	r.eventsList = append(r.eventsList, dispatchObsEvent{kind: "gate_released", cidr: cidr, taskIndex: taskIndex})
}

func (r *recordingDispatchObserver) OnBucketWaitStart(cidr string, taskIndex int) {
	r.eventsList = append(r.eventsList, dispatchObsEvent{kind: "bucket_wait_start", cidr: cidr, taskIndex: taskIndex})
}

func (r *recordingDispatchObserver) OnBucketAcquired(cidr string, taskIndex int) {
	r.eventsList = append(r.eventsList, dispatchObsEvent{kind: "bucket_acquired", cidr: cidr, taskIndex: taskIndex})
}

func (r *recordingDispatchObserver) OnTaskEnqueued(cidr string, taskIndex int) {
	r.eventsList = append(r.eventsList, dispatchObsEvent{kind: "task_enqueued", cidr: cidr, taskIndex: taskIndex})
}

func TestDispatchTasks_WhenObserverInjected_EmitsOrderedEvents(t *testing.T) {
	ctrl := speedctrl.NewController()
	logger := newLogger("error", false, nil)

	bucket := ratelimit.NewLeakyBucket(100, 1)
	defer bucket.Close()

	ch := &task.Chunk{CIDR: "10.0.0.0/24", TotalCount: 1, Status: "pending"}
	rt := &chunkRuntime{
		ipCidr: "10.0.0.0/24",
		ports:  []int{80},
		targets: []scanTarget{
			{ip: "10.0.0.1", ipCidr: "10.0.0.0/24"},
		},
		state:   ch,
		tracker: newChunkStateTracker(ch),
		bkt:     bucket,
	}

	taskCh := make(chan scanTask, 1)
	obs := &recordingDispatchObserver{}
	err := dispatchTasks(context.Background(), dispatchPolicy{delay: 0, observer: obs}, ctrl, logger, []*chunkRuntime{rt}, taskCh)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []dispatchObsEvent{
		{kind: "bucket_wait_start", cidr: "10.0.0.0/24", taskIndex: 0},
		{kind: "bucket_acquired", cidr: "10.0.0.0/24", taskIndex: 0},
		{kind: "gate_wait_start", cidr: "10.0.0.0/24", taskIndex: 0},
		{kind: "gate_released", cidr: "10.0.0.0/24", taskIndex: 0},
		{kind: "task_enqueued", cidr: "10.0.0.0/24", taskIndex: 0},
	}
	if len(obs.eventsList) != len(want) {
		t.Fatalf("unexpected event count: got=%d want=%d events=%#v", len(obs.eventsList), len(want), obs.eventsList)
	}
	for i := range want {
		if obs.eventsList[i] != want[i] {
			t.Fatalf("event[%d] mismatch: got=%#v want=%#v", i, obs.eventsList[i], want[i])
		}
	}
}
