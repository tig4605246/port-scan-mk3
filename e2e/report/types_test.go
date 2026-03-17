package report

import (
	"encoding/json"
	"strings"
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

func TestFullReport_WhenMarshaled_UsesSnakeCaseSummaryFields(t *testing.T) {
	r := FullReport{
		Summary: Summary{
			Total:   3,
			Open:    1,
			Closed:  1,
			Timeout: 1,
		},
	}

	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	got := string(raw)
	for _, want := range []string{`"summary":{"total":3`, `"open":1`, `"closed":1`, `"timeout":1`} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected JSON to contain %s, got %s", want, got)
		}
	}
}
