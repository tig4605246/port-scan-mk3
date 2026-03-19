package speedcontrol

import (
	"strings"
	"testing"
	"time"
)

func TestAnalyze_WhenPauseWindowsContainNoEnqueue_PassesPauseBlockingRule(t *testing.T) {
	events := []Event{
		{Kind: EventTaskEnqueued, TimestampNS: time.Unix(0, 0).UnixNano()},
		{Kind: EventTaskEnqueued, TimestampNS: time.Unix(0, 50*time.Millisecond.Nanoseconds()).UnixNano()},
		{Kind: EventTaskEnqueued, TimestampNS: time.Unix(0, 250*time.Millisecond.Nanoseconds()).UnixNano()},
	}

	verdict := Analyze(events, RuleExpectation{
		Name:                 "G1",
		RequirePauseBlocking: true,
		PauseWindows: []PauseWindow{
			{StartNS: time.Unix(0, 100*time.Millisecond.Nanoseconds()).UnixNano(), EndNS: time.Unix(0, 200*time.Millisecond.Nanoseconds()).UnixNano()},
		},
	})

	if !verdict.Pass {
		t.Fatalf("expected pass, got fail: %+v", verdict)
	}
	if verdict.Metrics.PauseViolations != 0 {
		t.Fatalf("expected 0 pause violations, got %+v", verdict.Metrics)
	}
}

func TestAnalyze_WhenSteadyStateWithinTolerance_PassesTPSRule(t *testing.T) {
	events := []Event{
		{Kind: EventTaskEnqueued, TimestampNS: int64(0 * time.Millisecond)},
		{Kind: EventTaskEnqueued, TimestampNS: int64(100 * time.Millisecond)},
		{Kind: EventTaskEnqueued, TimestampNS: int64(200 * time.Millisecond)},
		{Kind: EventTaskEnqueued, TimestampNS: int64(300 * time.Millisecond)},
		{Kind: EventTaskEnqueued, TimestampNS: int64(400 * time.Millisecond)},
		{Kind: EventTaskEnqueued, TimestampNS: int64(500 * time.Millisecond)},
	}

	verdict := Analyze(events, RuleExpectation{
		Name:        "C1",
		ExpectedTPS: 10,
		Tolerance:   0.20,
	})
	if !verdict.Pass {
		t.Fatalf("expected tps check pass, got fail: %+v", verdict)
	}
	if verdict.Metrics.ObservedTPS <= 0 {
		t.Fatalf("expected observed tps > 0, got %+v", verdict.Metrics)
	}
}

func TestAnalyze_WhenExpectationMissed_ReturnsExplainableVerdict(t *testing.T) {
	events := []Event{
		{Kind: EventTaskEnqueued, TimestampNS: int64(0 * time.Millisecond)},
		{Kind: EventTaskEnqueued, TimestampNS: int64(1000 * time.Millisecond)},
	}

	verdict := Analyze(events, RuleExpectation{
		Name:        "C2",
		ExpectedTPS: 10,
		Tolerance:   0.05,
	})

	if verdict.Pass {
		t.Fatalf("expected fail verdict, got pass: %+v", verdict)
	}
	if strings.TrimSpace(verdict.Expected) == "" || strings.TrimSpace(verdict.Observed) == "" {
		t.Fatalf("expected and observed should be non-empty: %+v", verdict)
	}
	if strings.TrimSpace(verdict.Attribution) == "" || strings.TrimSpace(verdict.Explanation) == "" {
		t.Fatalf("attribution and explanation should be non-empty: %+v", verdict)
	}
}

func TestAnalyze_WhenEnqueueAtPauseWindowEnd_DoesNotCountViolation(t *testing.T) {
	end := time.Unix(0, 200*time.Millisecond.Nanoseconds()).UnixNano()
	events := []Event{
		{Kind: EventTaskEnqueued, TimestampNS: end},
	}

	verdict := Analyze(events, RuleExpectation{
		Name:                 "G1_boundary_end",
		RequirePauseBlocking: true,
		PauseWindows: []PauseWindow{
			{
				StartNS: time.Unix(0, 100*time.Millisecond.Nanoseconds()).UnixNano(),
				EndNS:   end,
			},
		},
	})

	if verdict.Metrics.PauseViolations != 0 {
		t.Fatalf("expected no pause violation on window end boundary, got %+v", verdict.Metrics)
	}
	if !verdict.Pass {
		t.Fatalf("expected pass on window end boundary, got %+v", verdict)
	}
}
