package integration

import "testing"

func TestSpeedControlCIDRScenarios(t *testing.T) {
	steady := findScenario(t, "C1_single_cidr_steady_rate")
	if !steady.Verdict.Pass {
		t.Fatalf("C1 should pass: %+v", steady.Verdict)
	}
	if steady.Verdict.Metrics.ObservedTPS <= 0 {
		t.Fatalf("C1 should report observed TPS > 0: %+v", steady.Verdict.Metrics)
	}

	burst := findScenario(t, "C2_single_cidr_burst_then_steady")
	if !burst.Verdict.Pass {
		t.Fatalf("C2 should pass: %+v", burst.Verdict)
	}
	if burst.Verdict.Metrics.BurstEvents < burst.Expectation.MinImmediateBurst {
		t.Fatalf("C2 burst events too low: got=%d want>=%d", burst.Verdict.Metrics.BurstEvents, burst.Expectation.MinImmediateBurst)
	}
}
