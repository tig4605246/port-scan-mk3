package integration

import (
	"testing"

	"github.com/xuxiping/port-scan-mk3/internal/testkit"
)

func TestRunIntegrationScenario_DisabledAPI(t *testing.T) {
	res, err := RunIntegrationScenario(Scenario{DisableAPI: true})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.PauseCount != 0 {
		t.Fatalf("expected no pause, got %d", res.PauseCount)
	}
}

func TestRunIntegrationScenario_BadAPI(t *testing.T) {
	_, err := RunIntegrationScenario(Scenario{PressureAPI: "http://127.0.0.1:1", Threshold: 90})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunIntegrationScenario_CustomThresholdAvoidsPause(t *testing.T) {
	api := testkit.NewMockPressureAPI([]int{20, 89, 90, 30})
	defer api.Close()

	res, err := RunIntegrationScenario(Scenario{
		PressureAPI: api.URL(),
		Threshold:   95,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.PauseCount != 0 {
		t.Fatalf("expected no pause above custom threshold, got %d", res.PauseCount)
	}
	if res.TotalScanned != res.TotalTargets {
		t.Fatalf("expected complete scan outcome, got %+v", res)
	}
}

func TestRunIntegrationScenario_ResumeTakesPriorityOverPressureFlow(t *testing.T) {
	res, err := RunIntegrationScenario(Scenario{
		Resume:      true,
		PressureAPI: "http://127.0.0.1:1",
		Threshold:   90,
	})
	if err != nil {
		t.Fatalf("expected resume shortcut to avoid pressure failure, got %v", err)
	}
	if res.TotalScanned != res.TotalTargets || res.DuplicateCount != 0 || res.MissingCount != 0 {
		t.Fatalf("unexpected resume outcome: %+v", res)
	}
}
