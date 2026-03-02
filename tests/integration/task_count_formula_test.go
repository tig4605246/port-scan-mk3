package integration

import "testing"

func TestTaskCountFormula_UniqueExpandedIPsTimesPorts(t *testing.T) {
	result, err := RunIntegrationScenario(Scenario{DisableAPI: true})
	if err != nil {
		t.Fatalf("scenario failed: %v", err)
	}

	// Baseline fixture model: unique expanded IPs=2, ports=2 => expected task count=4.
	uniqueExpandedIPs := 2
	portCount := 2
	want := uniqueExpandedIPs * portCount
	if result.TotalTargets != want {
		t.Fatalf("task formula mismatch: want=%d got=%d", want, result.TotalTargets)
	}
	if result.TotalScanned != result.TotalTargets {
		t.Fatalf("scanned/target mismatch: %d/%d", result.TotalScanned, result.TotalTargets)
	}
}
