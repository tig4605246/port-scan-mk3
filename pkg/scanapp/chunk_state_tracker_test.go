package scanapp

import (
	"sync"
	"testing"

	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

func TestChunkStateTracker_WhenConcurrentMutations_MaintainsConsistentState(t *testing.T) {
	ch := &task.Chunk{
		CIDR:         "10.0.0.0/24",
		TotalCount:   1000,
		NextIndex:    0,
		ScannedCount: 0,
		Status:       "pending",
	}
	tracker := newChunkStateTracker(ch)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			tracker.AdvanceNextIndex(i + 1)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			tracker.IncrementScanned()
		}
	}()

	wg.Wait()

	snap := tracker.Snapshot()
	if snap.NextIndex != 1000 {
		t.Fatalf("expected NextIndex=1000, got %d", snap.NextIndex)
	}
	if snap.ScannedCount != 1000 {
		t.Fatalf("expected ScannedCount=1000, got %d", snap.ScannedCount)
	}
	if snap.Status != "completed" {
		t.Fatalf("expected Status=completed, got %s", snap.Status)
	}
}

func TestChunkStateTracker_WhenPartialProgress_ReportsScanning(t *testing.T) {
	ch := &task.Chunk{TotalCount: 10, Status: "pending"}
	tracker := newChunkStateTracker(ch)

	tracker.AdvanceNextIndex(5)
	tracker.IncrementScanned()

	snap := tracker.Snapshot()
	if snap.NextIndex != 5 {
		t.Fatalf("expected NextIndex=5, got %d", snap.NextIndex)
	}
	if snap.ScannedCount != 1 {
		t.Fatalf("expected ScannedCount=1, got %d", snap.ScannedCount)
	}
	if snap.Status != "scanning" {
		t.Fatalf("expected Status=scanning, got %s", snap.Status)
	}
}
