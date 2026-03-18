package integration

import "testing"

func TestSpeedControlGlobalScenarios(t *testing.T) {
	manual := findScenario(t, "G1_manual_pause")
	if !manual.Verdict.Pass {
		t.Fatalf("G1 should pass: %+v", manual.Verdict)
	}
	if manual.Verdict.Metrics.PauseViolations != 0 {
		t.Fatalf("G1 should have zero pause violations: %+v", manual.Verdict.Metrics)
	}

	orGate := findScenario(t, "G3_or_gate")
	if !orGate.Verdict.Pass {
		t.Fatalf("G3 should pass: %+v", orGate.Verdict)
	}
	if orGate.Verdict.Metrics.PauseViolations != 0 {
		t.Fatalf("G3 should have zero pause violations: %+v", orGate.Verdict.Metrics)
	}
}
