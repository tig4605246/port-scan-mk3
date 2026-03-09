package integration

import (
	"testing"

	"github.com/xuxiping/port-scan-mk3/internal/testkit"
)

func TestScanPipeline_PausesOnPressureAndResumes(t *testing.T) {
	api := testkit.NewMockPressureAPI([]int{20, 95, 95, 30})
	defer api.Close()

	result, err := RunIntegrationScenario(Scenario{
		PressureAPI: api.URL(),
		DisableAPI:  false,
		Threshold:   90,
	})
	if err != nil {
		t.Fatalf("scenario failed: %v", err)
	}
	if result.PauseCount == 0 {
		t.Fatal("expected at least one pause")
	}
	if result.TotalScanned != result.TotalTargets {
		t.Fatalf("scan incomplete: %d/%d", result.TotalScanned, result.TotalTargets)
	}
}

func TestScanPipeline_WithoutPressureAPICompletesWithoutPause(t *testing.T) {
	result, err := RunIntegrationScenario(Scenario{DisableAPI: true})
	if err != nil {
		t.Fatalf("scenario failed: %v", err)
	}
	if result.PauseCount != 0 {
		t.Fatalf("expected no pause events, got %d", result.PauseCount)
	}
	if result.TotalScanned != result.TotalTargets {
		t.Fatalf("scan incomplete: %d/%d", result.TotalScanned, result.TotalTargets)
	}
}

func TestScanPipeline_UsesDefaultThresholdWhenUnset(t *testing.T) {
	api := testkit.NewMockPressureAPI([]int{20, 89, 90, 30})
	defer api.Close()

	result, err := RunIntegrationScenario(Scenario{
		PressureAPI: api.URL(),
		DisableAPI:  false,
	})
	if err != nil {
		t.Fatalf("scenario failed: %v", err)
	}
	if result.PauseCount != 1 {
		t.Fatalf("expected exactly one pause at default threshold, got %d", result.PauseCount)
	}
}

func TestScanPipeline_DefaultScenarioCompletesWithoutLoss(t *testing.T) {
	result, err := RunIntegrationScenario(Scenario{})
	if err != nil {
		t.Fatalf("scenario failed: %v", err)
	}
	if result.TotalTargets != 4 {
		t.Fatalf("expected 4 total targets, got %d", result.TotalTargets)
	}
	if result.TotalScanned != result.TotalTargets {
		t.Fatalf("scan incomplete: %d/%d", result.TotalScanned, result.TotalTargets)
	}
	if result.DuplicateCount != 0 || result.MissingCount != 0 {
		t.Fatalf("unexpected duplicate/missing counts: %+v", result)
	}
}
