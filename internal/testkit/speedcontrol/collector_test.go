package speedcontrol

import (
	"fmt"
	"sync"
	"testing"
)

func TestCollector_WhenEventsRecorded_KeepsOrderAndFields(t *testing.T) {
	c := NewCollector("G1")

	c.Record(Event{Kind: EventGateWaitStart, CIDR: "10.0.0.0/24", TaskIndex: 0, TimestampNS: 100})
	c.Record(Event{Kind: EventTaskEnqueued, CIDR: "10.0.0.0/24", TaskIndex: 0, TimestampNS: 120})

	got := c.Events()
	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}
	if got[0].Kind != EventGateWaitStart || got[1].Kind != EventTaskEnqueued {
		t.Fatalf("unexpected event order: %#v", got)
	}
	if got[0].CIDR != "10.0.0.0/24" || got[0].TaskIndex != 0 || got[0].TimestampNS != 100 {
		t.Fatalf("unexpected first event fields: %#v", got[0])
	}
}

func TestCollector_WhenEventsReturned_ModifyingCopyDoesNotMutateCollector(t *testing.T) {
	c := NewCollector("G2")
	c.Record(Event{Kind: EventBucketAcquired, CIDR: "10.0.0.0/24", TaskIndex: 1, TimestampNS: 200})

	got := c.Events()
	got[0].CIDR = "mutated"
	got[0].TaskIndex = 99

	again := c.Events()
	if again[0].CIDR != "10.0.0.0/24" || again[0].TaskIndex != 1 {
		t.Fatalf("collector state was mutated through returned slice: %#v", again[0])
	}
}

func TestCollector_WhenRecordedConcurrently_KeepsAllEvents(t *testing.T) {
	c := NewCollector("G3")

	const total = 64
	var wg sync.WaitGroup
	wg.Add(total)
	for i := 0; i < total; i++ {
		i := i
		go func() {
			defer wg.Done()
			c.Record(Event{
				Kind:        EventTaskEnqueued,
				CIDR:        "10.0.0.0/24",
				TaskIndex:   i,
				TimestampNS: int64(1000 + i),
				Scenario:    fmt.Sprintf("S-%d", i%2),
			})
		}()
	}
	wg.Wait()

	got := c.Events()
	if len(got) != total {
		t.Fatalf("expected %d events, got %d", total, len(got))
	}
}
