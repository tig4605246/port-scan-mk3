package report

import "time"

// ScenarioResult captures the outcome of one e2e scenario.
type ScenarioResult struct {
	Name       string        `json:"name"`
	Status     string        `json:"status"`
	Duration   time.Duration `json:"duration_ns"`
	ExitCode   int           `json:"exit_code"`
	Checklist  []CheckItem   `json:"checklist"`
	Artifacts  []Artifact    `json:"artifacts"`
	LogSnippet string        `json:"log_snippet,omitempty"`
}

// CheckItem is one pass/fail assertion in a scenario checklist.
type CheckItem struct {
	Description string `json:"description"`
	Passed      bool   `json:"passed"`
	Detail      string `json:"detail,omitempty"`
}

// Artifact describes a file produced by an e2e scenario.
type Artifact struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	RowCount int    `json:"row_count,omitempty"`
}

// FullReport aggregates all scenario results into the final e2e report.
type FullReport struct {
	Timestamp     time.Time        `json:"timestamp"`
	TotalDuration time.Duration    `json:"total_duration_ns"`
	Scenarios     []ScenarioResult `json:"scenarios"`
	Summary       Summary          `json:"summary"`
	PassCount     int              `json:"pass_count"`
	FailCount     int              `json:"fail_count"`
}

// ComputeCounts populates PassCount and FailCount from Scenarios.
func (r *FullReport) ComputeCounts() {
	r.PassCount = 0
	r.FailCount = 0
	for _, s := range r.Scenarios {
		if s.Status == "pass" {
			r.PassCount++
			continue
		}
		r.FailCount++
	}
}

// Manifest is the JSON structure written by run_e2e.sh for the Go report tool.
type Manifest struct {
	StartTime int64              `json:"start_time_ns"`
	EndTime   int64              `json:"end_time_ns"`
	Scenarios []ManifestScenario `json:"scenarios"`
}

// ManifestScenario captures per-scenario data from the shell script.
type ManifestScenario struct {
	Name      string   `json:"name"`
	ExitCode  int      `json:"exit_code"`
	StartNs   int64    `json:"start_ns"`
	EndNs     int64    `json:"end_ns"`
	Artifacts []string `json:"artifacts"`
	LogFile   string   `json:"log_file,omitempty"`
}
