package speedcontrol

type EventKind string

const (
	EventGateWaitStart   EventKind = "gate_wait_start"
	EventGateReleased    EventKind = "gate_released"
	EventBucketWaitStart EventKind = "bucket_wait_start"
	EventBucketAcquired  EventKind = "bucket_acquired"
	EventTaskEnqueued    EventKind = "task_enqueued"
)

type Event struct {
	Kind        EventKind `json:"kind"`
	Scenario    string    `json:"scenario,omitempty"`
	CIDR        string    `json:"cidr,omitempty"`
	TaskIndex   int       `json:"task_index,omitempty"`
	TimestampNS int64     `json:"timestamp_ns,omitempty"`
}

type PauseWindow struct {
	StartNS int64  `json:"start_ns"`
	EndNS   int64  `json:"end_ns"`
	Source  string `json:"source,omitempty"`
}

type RuleExpectation struct {
	Name                 string        `json:"name"`
	ExpectedTPS          float64       `json:"expected_tps,omitempty"`
	Tolerance            float64       `json:"tolerance,omitempty"`
	RequirePauseBlocking bool          `json:"require_pause_blocking,omitempty"`
	PauseWindows         []PauseWindow `json:"pause_windows,omitempty"`
	MinImmediateBurst    int           `json:"min_immediate_burst,omitempty"`
	BurstMaxGapNS        int64         `json:"burst_max_gap_ns,omitempty"`

	// Optional formula inputs used when ExpectedTPS is unset.
	BucketRate     float64 `json:"bucket_rate,omitempty"`
	DelaySeconds   float64 `json:"delay_seconds,omitempty"`
	Workers        int     `json:"workers,omitempty"`
	AvgTaskSeconds float64 `json:"avg_task_seconds,omitempty"`
}

type ScenarioMetrics struct {
	TaskEnqueued    int     `json:"task_enqueued"`
	ObservedTPS     float64 `json:"observed_tps"`
	ExpectedTPS     float64 `json:"expected_tps"`
	Tolerance       float64 `json:"tolerance"`
	PauseViolations int     `json:"pause_violations"`
	BurstEvents     int     `json:"burst_events"`
}

type ScenarioVerdict struct {
	Name        string          `json:"name"`
	Pass        bool            `json:"pass"`
	Expected    string          `json:"expected"`
	Observed    string          `json:"observed"`
	Attribution string          `json:"attribution"`
	Explanation string          `json:"explanation"`
	Metrics     ScenarioMetrics `json:"metrics"`
}

type ScenarioRun struct {
	Name        string                 `json:"name"`
	Config      map[string]any         `json:"config,omitempty"`
	Expectation RuleExpectation        `json:"expectation"`
	Events      []Event                `json:"events"`
	Verdict     ScenarioVerdict        `json:"verdict"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}
