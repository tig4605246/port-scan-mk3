package report

import (
	"testing"
	"time"
)

func TestFullReport_WhenScenariosAdded_ComputesPassFailCounts(t *testing.T) {
	r := FullReport{
		Timestamp:     time.Now(),
		TotalDuration: 5 * time.Second,
		Scenarios: []ScenarioResult{
			{Name: "normal", Status: "pass"},
			{Name: "api_5xx", Status: "pass"},
			{Name: "api_timeout", Status: "fail"},
		},
	}
	r.ComputeCounts()
	if r.PassCount != 2 || r.FailCount != 1 {
		t.Fatalf("expected pass=2 fail=1, got pass=%d fail=%d", r.PassCount, r.FailCount)
	}
}
