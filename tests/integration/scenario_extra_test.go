package integration

import "testing"

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
