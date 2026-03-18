package integration

import (
	"sync"
	"testing"

	speedctrlkit "github.com/xuxiping/port-scan-mk3/internal/testkit/speedcontrol"
)

var (
	speedRunsOnce sync.Once
	speedRuns     []speedctrlkit.ScenarioRun
	speedRunsErr  error
)

func loadSpeedRuns(t *testing.T) []speedctrlkit.ScenarioRun {
	t.Helper()
	speedRunsOnce.Do(func() {
		speedRuns, speedRunsErr = speedctrlkit.RunScenarioMatrix()
	})
	if speedRunsErr != nil {
		t.Fatalf("run speed-control scenario matrix: %v", speedRunsErr)
	}
	return speedRuns
}

func findScenario(t *testing.T, name string) speedctrlkit.ScenarioRun {
	t.Helper()
	runs := loadSpeedRuns(t)
	for _, run := range runs {
		if run.Name == name {
			return run
		}
	}
	t.Fatalf("scenario not found: %s", name)
	return speedctrlkit.ScenarioRun{}
}
