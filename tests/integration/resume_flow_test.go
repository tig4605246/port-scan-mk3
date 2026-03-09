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

func TestResumeFlow_WhenEnabled_PreservesExpectedTargetCount(t *testing.T) {
	result, err := RunIntegrationScenario(Scenario{Resume: true})
	if err != nil {
		t.Fatalf("scenario failed: %v", err)
	}
	if result.TotalTargets != 4 {
		t.Fatalf("expected 4 total targets, got %d", result.TotalTargets)
	}
	if result.TotalScanned != 4 {
		t.Fatalf("expected 4 scanned targets, got %d", result.TotalScanned)
	}
}

func TestResumeFlow_WhenEnabled_PreservesOperatorVisibleOutcomes(t *testing.T) {
	result, err := RunIntegrationScenario(Scenario{Resume: true})
	if err != nil {
		t.Fatalf("scenario failed: %v", err)
	}
	if result.TotalScanned != result.TotalTargets {
		t.Fatalf("expected scanned/target parity, got %+v", result)
	}
	if result.DuplicateCount != 0 || result.MissingCount != 0 {
		t.Fatalf("expected duplicate/missing free resume outcome, got %+v", result)
	}
}
