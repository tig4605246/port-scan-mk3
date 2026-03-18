package scanapp

import (
	"testing"
	"time"
)

func TestDashboardState_ProgressUpdates(t *testing.T) {
	current := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	state := newDashboardState(4, func() time.Time { return current })

	state.OnTaskEnqueued("10.0.0.0/24")
	state.OnTaskEnqueued("10.0.0.0/24")
	state.OnResult()

	snap := state.Snapshot()
	if snap.ScannedTasks != 1 {
		t.Fatalf("expected ScannedTasks=1, got %d", snap.ScannedTasks)
	}
	if snap.TotalTasks != 4 {
		t.Fatalf("expected TotalTasks=4, got %d", snap.TotalTasks)
	}
	if snap.Percent != 25 {
		t.Fatalf("expected Percent=25, got %v", snap.Percent)
	}
}

func TestDashboardState_CurrentCIDRAndBucketStatusTransitions(t *testing.T) {
	current := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	state := newDashboardState(2, func() time.Time { return current })

	state.OnTaskEnqueued("10.0.0.0/24")
	snap := state.Snapshot()
	if snap.CurrentCIDR != "10.0.0.0/24" {
		t.Fatalf("expected CurrentCIDR to follow task enqueue, got %q", snap.CurrentCIDR)
	}
	if snap.BucketStatus != "" {
		t.Fatalf("expected empty BucketStatus before updates, got %q", snap.BucketStatus)
	}

	state.OnBucketStatus("10.0.0.0/24", "waiting")
	snap = state.Snapshot()
	if snap.CurrentCIDR != "10.0.0.0/24" || snap.BucketStatus != "waiting" {
		t.Fatalf("unexpected first bucket snapshot: %#v", snap)
	}

	state.OnBucketStatus("10.0.1.0/24", "acquired")
	snap = state.Snapshot()
	if snap.CurrentCIDR != "10.0.1.0/24" || snap.BucketStatus != "acquired" {
		t.Fatalf("unexpected second bucket snapshot: %#v", snap)
	}
}

func TestDashboardState_ControllerStatusMapping(t *testing.T) {
	current := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	state := newDashboardState(1, func() time.Time { return current })

	cases := []struct {
		name         string
		manualPaused bool
		apiPaused    bool
		want         string
	}{
		{name: "running", want: "RUNNING"},
		{name: "api paused", apiPaused: true, want: "PAUSED(API)"},
		{name: "manual paused", manualPaused: true, want: "PAUSED(MANUAL)"},
		{name: "api and manual paused", manualPaused: true, apiPaused: true, want: "PAUSED(API+MANUAL)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state.OnController(tc.manualPaused, tc.apiPaused)
			snap := state.Snapshot()
			if snap.ControllerStatus != tc.want {
				t.Fatalf("expected ControllerStatus=%q, got %q", tc.want, snap.ControllerStatus)
			}
		})
	}
}

func TestDashboardState_SlidingWindowSpeeds(t *testing.T) {
	current := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	state := newDashboardState(10, func() time.Time { return current })

	current = current.Add(-6 * time.Second)
	state.OnTaskEnqueued("10.0.0.0/24")

	current = current.Add(2 * time.Second)
	state.OnTaskEnqueued("10.0.0.0/24")
	state.OnResult()

	current = current.Add(3 * time.Second)
	state.OnTaskEnqueued("10.0.0.0/24")
	state.OnResult()

	current = current.Add(1 * time.Second)
	snap := state.Snapshot()
	if snap.DispatchPerSecond != 0.4 {
		t.Fatalf("expected DispatchPerSecond=0.4, got %v", snap.DispatchPerSecond)
	}
	if snap.ResultsPerSecond != 0.4 {
		t.Fatalf("expected ResultsPerSecond=0.4, got %v", snap.ResultsPerSecond)
	}
}

func TestDashboardState_APIHealthTextAndTimestamps(t *testing.T) {
	current := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	state := newDashboardState(1, func() time.Time { return current })

	okAt := current.Add(2 * time.Second)
	failAt := current.Add(7 * time.Second)

	state.OnPressureSample(81, okAt)
	snap := state.Snapshot()
	if snap.APIHealthText != "ok" {
		t.Fatalf("expected APIHealthText=ok after sample, got %q", snap.APIHealthText)
	}
	if !snap.LastPressureUpdateAt.Equal(okAt) {
		t.Fatalf("expected LastPressureUpdateAt=%v, got %v", okAt, snap.LastPressureUpdateAt)
	}
	if !snap.LastPressureFailureAt.IsZero() {
		t.Fatalf("expected zero LastPressureFailureAt, got %v", snap.LastPressureFailureAt)
	}

	state.OnPressureFailure(3, failAt)
	snap = state.Snapshot()
	if snap.APIHealthText != "fail streak 3" {
		t.Fatalf("expected APIHealthText=fail streak 3 after failure, got %q", snap.APIHealthText)
	}
	if !snap.LastPressureUpdateAt.Equal(okAt) {
		t.Fatalf("expected LastPressureUpdateAt to retain last success timestamp, got %v", snap.LastPressureUpdateAt)
	}
	if !snap.LastPressureFailureAt.Equal(failAt) {
		t.Fatalf("expected LastPressureFailureAt=%v, got %v", failAt, snap.LastPressureFailureAt)
	}
}
