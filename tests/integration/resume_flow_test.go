package integration

import "testing"

func TestResumeFlow_CompletesAllTargets(t *testing.T) {
	result, err := RunIntegrationScenario(Scenario{Resume: true})
	if err != nil {
		t.Fatalf("scenario failed: %v", err)
	}
	if result.TotalScanned != result.TotalTargets {
		t.Fatalf("resume incomplete: %d/%d", result.TotalScanned, result.TotalTargets)
	}
	if result.DuplicateCount != 0 {
		t.Fatalf("resume produced duplicate results: %d", result.DuplicateCount)
	}
	if result.MissingCount != 0 {
		t.Fatalf("resume missed results: %d", result.MissingCount)
	}
}
