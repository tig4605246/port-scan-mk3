package integration

import "testing"

func TestSpeedControlCombinedScenario(t *testing.T) {
	combined := findScenario(t, "X1_combined_global_pause_with_cidr_bucket")
	if !combined.Verdict.Pass {
		t.Fatalf("X1 should pass: %+v", combined.Verdict)
	}
	if combined.Verdict.Metrics.PauseViolations != 0 {
		t.Fatalf("X1 should have zero pause violations: %+v", combined.Verdict.Metrics)
	}

	cidrSet := make(map[string]struct{})
	for _, e := range combined.Events {
		if e.CIDR != "" {
			cidrSet[e.CIDR] = struct{}{}
		}
	}
	if len(cidrSet) < 2 {
		t.Fatalf("X1 should contain events from at least two CIDRs, got=%d", len(cidrSet))
	}
}
