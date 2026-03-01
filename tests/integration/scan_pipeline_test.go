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
