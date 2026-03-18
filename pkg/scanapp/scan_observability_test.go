package scanapp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
	"github.com/xuxiping/port-scan-mk3/pkg/speedctrl"
	"github.com/xuxiping/port-scan-mk3/pkg/task"
	"github.com/xuxiping/port-scan-mk3/pkg/writer"
)

type dashboardSnapshotRecorder struct {
	mu    sync.Mutex
	snaps []dashboardSnapshot
}

func (r *dashboardSnapshotRecorder) Render(_ io.Writer, snap dashboardSnapshot) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.snaps = append(r.snaps, snap)
	return nil
}

func (r *dashboardSnapshotRecorder) snapshots() []dashboardSnapshot {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]dashboardSnapshot, len(r.snaps))
	copy(out, r.snaps)
	return out
}

type sequencePressureFetcher struct {
	mu     sync.Mutex
	values []int
}

func (f *sequencePressureFetcher) Fetch(context.Context) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.values) == 0 {
		return 0, errors.New("no pressure values configured")
	}
	value := f.values[0]
	if len(f.values) > 1 {
		f.values = f.values[1:]
	}
	return value, nil
}

type scriptedPressureResult struct {
	pressure int
	err      error
}

type scriptedPressureFetcher struct {
	mu      sync.Mutex
	results []scriptedPressureResult
}

func (f *scriptedPressureFetcher) Fetch(context.Context) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.results) == 0 {
		return 0, errors.New("no scripted pressure results configured")
	}
	result := f.results[0]
	if len(f.results) > 1 {
		f.results = f.results[1:]
	}
	return result.pressure, result.err
}

type pressureTelemetryRecorder struct {
	mu           sync.Mutex
	samples      []int
	sampleTimes  []time.Time
	failures     []int
	failureTimes []time.Time
}

type controllerTelemetryRecorder struct {
	mu        sync.Mutex
	callbacks int
	statuses  []string
}

func (r *controllerTelemetryRecorder) OnController(manualPaused, apiPaused bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.callbacks++
	r.statuses = append(r.statuses, dashboardControllerStatus(manualPaused, apiPaused))
}

func (r *pressureTelemetryRecorder) OnPressureSample(pressure int, t time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.samples = append(r.samples, pressure)
	r.sampleTimes = append(r.sampleTimes, t)
}

func (r *pressureTelemetryRecorder) OnPressureFailure(streak int, t time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.failures = append(r.failures, streak)
	r.failureTimes = append(r.failureTimes, t)
}

func TestRun_WhenObservabilityJSONEnabled_EmitsProgressAndCompletionEvents(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	openPort, _ := strconv.Atoi(portStr)

	tmp := t.TempDir()
	cidrFile := filepath.Join(tmp, "cidr.csv")
	portFile := filepath.Join(tmp, "ports.csv")
	outFile := filepath.Join(tmp, "scan_results.csv")
	if err := os.WriteFile(cidrFile, []byte("fab_name,ip,ip_cidr,cidr_name\nfab1,127.0.0.1,127.0.0.1/32,loopback\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(portFile, []byte(strconv.Itoa(openPort)+"/tcp\n1/tcp\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		CIDRFile:         cidrFile,
		PortFile:         portFile,
		Output:           outFile,
		Timeout:          100 * time.Millisecond,
		Delay:            0,
		BucketRate:       100,
		BucketCapacity:   100,
		Workers:          1,
		PressureInterval: 5 * time.Second,
		DisableAPI:       true,
		LogLevel:         "info",
		Format:           "json",
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	if err := Run(context.Background(), cfg, stdout, stderr, RunOptions{DisableKeyboard: true, ProgressInterval: 1}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	logs := stderr.String()
	for _, required := range []string{
		"\"target\"",
		"\"port\"",
		"\"state_transition\"",
		"\"error_cause\"",
		"\"state_transition\":\"progress\"",
		"\"state_transition\":\"completion_summary\"",
		"\"success\":true",
	} {
		if !strings.Contains(logs, required) {
			t.Fatalf("missing observability marker %s in logs: %s", required, logs)
		}
	}

	if !strings.Contains(stdout.String(), "progress cidr=") {
		t.Fatalf("expected progress output on stdout, got %s", stdout.String())
	}
}

func TestRun_WhenRichDashboardEnabled_ReceivesLiveTelemetryState(t *testing.T) {
	tmp := t.TempDir()
	cidrFile := filepath.Join(tmp, "cidr.csv")
	portFile := filepath.Join(tmp, "ports.csv")
	outFile := filepath.Join(tmp, "scan_results.csv")

	if err := os.WriteFile(cidrFile, []byte("fab_name,ip,ip_cidr,cidr_name\nfab1,127.0.0.1,127.0.0.1/32,loopback\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(portFile, []byte("1/tcp\n2/tcp\n3/tcp\n4/tcp\n5/tcp\n6/tcp\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		CIDRFile:         cidrFile,
		PortFile:         portFile,
		Output:           outFile,
		Timeout:          100 * time.Millisecond,
		Delay:            10 * time.Millisecond,
		BucketRate:       100,
		BucketCapacity:   100,
		Workers:          1,
		PressureInterval: 10 * time.Millisecond,
		DisableAPI:       false,
		LogLevel:         "error",
		Format:           "human",
	}

	recorder := &dashboardSnapshotRecorder{}
	err := Run(context.Background(), cfg, &bytes.Buffer{}, &bytes.Buffer{}, RunOptions{
		DisableKeyboard: true,
		Dial: func(context.Context, string, string) (net.Conn, error) {
			time.Sleep(25 * time.Millisecond)
			return nil, errors.New("dial refused for test")
		},
		PressureLimit:             90,
		PressureFetcher:           &sequencePressureFetcher{values: []int{95, 95, 20, 20, 20}},
		dashboardTerminalDetector: func(io.Writer) bool { return true },
		dashboardRefreshInterval:  10 * time.Millisecond,
		dashboardRenderer:         recorder,
		ProgressInterval:          1,
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	snaps := recorder.snapshots()
	if len(snaps) == 0 {
		t.Fatal("expected dashboard snapshots during run")
	}

	var (
		sawCIDR            bool
		sawBucketStatus    bool
		sawDispatchRate    bool
		sawResultsRate     bool
		sawControllerState bool
		sawPressure        bool
		sawAPIHealth       bool
	)
	for _, snap := range snaps {
		if snap.CurrentCIDR != "" {
			sawCIDR = true
		}
		switch snap.BucketStatus {
		case "waiting_bucket", "waiting_gate", "enqueued":
			sawBucketStatus = true
		}
		if snap.DispatchPerSecond > 0 {
			sawDispatchRate = true
		}
		if snap.ResultsPerSecond > 0 {
			sawResultsRate = true
		}
		switch snap.ControllerStatus {
		case "RUNNING", "PAUSED(API)", "PAUSED(MANUAL)", "PAUSED(API+MANUAL)":
			sawControllerState = true
		}
		if snap.PressurePercent > 0 && !snap.LastPressureUpdateAt.IsZero() {
			sawPressure = true
		}
		if snap.APIHealthText == "ok" {
			sawAPIHealth = true
		}
	}

	if !sawCIDR {
		t.Fatalf("expected CurrentCIDR to be populated, got snapshots=%#v", snaps)
	}
	if !sawBucketStatus {
		t.Fatalf("expected BucketStatus transition in snapshots, got %#v", snaps)
	}
	if !sawDispatchRate {
		t.Fatalf("expected DispatchPerSecond > 0 in snapshots, got %#v", snaps)
	}
	if !sawResultsRate {
		t.Fatalf("expected ResultsPerSecond > 0 in snapshots, got %#v", snaps)
	}
	if !sawControllerState {
		t.Fatalf("expected controller status snapshots, got %#v", snaps)
	}
	if !sawPressure {
		t.Fatalf("expected pressure samples with timestamp in snapshots, got %#v", snaps)
	}
	if !sawAPIHealth {
		t.Fatalf("expected API health text update in snapshots, got %#v", snaps)
	}
}

func TestPollPressureAPI_WhenJSONLoggerEnabled_EmitsPauseResumeMessages(t *testing.T) {
	values := []int{95, 20}
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		v := values[0]
		if len(values) > 1 {
			values = values[1:]
		}
		_, _ = w.Write([]byte(`{"pressure":` + strconv.Itoa(v) + `}`))
	}))
	defer api.Close()

	ctrl := speedctrl.NewController()
	logOut := &lockedBuffer{}
	logger := newLogger("info", true, logOut)
	errCh := make(chan error, 1)

	ctx, cancel := context.WithCancel(context.Background())
	go pollPressureAPI(ctx, config.Config{
		PressureAPI:      api.URL,
		PressureInterval: 5 * time.Millisecond,
	}, RunOptions{
		PressureLimit: 90,
		PressureHTTP:  &http.Client{Timeout: time.Second},
	}, ctrl, logger, errCh)

	time.Sleep(40 * time.Millisecond)
	cancel()
	time.Sleep(10 * time.Millisecond)

	select {
	case err := <-errCh:
		t.Fatalf("unexpected err: %v", err)
	default:
	}

	logs := logOut.String()
	if !strings.Contains(logs, `"level":"info"`) {
		t.Fatalf("expected json info logs, got %s", logs)
	}
	if !strings.Contains(logs, "掃描已自動暫停") || !strings.Contains(logs, "掃描已自動恢復") {
		t.Fatalf("expected pause/resume messages, got %s", logs)
	}
}

func TestPollPressureAPI_WhenObserverInjected_ReportsSamplesAndFailures(t *testing.T) {
	ctrl := speedctrl.NewController()
	logOut := &lockedBuffer{}
	logger := newLogger("info", false, logOut)
	errCh := make(chan error, 1)
	observer := &pressureTelemetryRecorder{}
	controllerObserver := &controllerTelemetryRecorder{}

	ctx, cancel := context.WithCancel(context.Background())
	go pollPressureAPI(ctx, config.Config{
		PressureInterval: 5 * time.Millisecond,
	}, RunOptions{
		PressureLimit:       90,
		PressureFetcher:     &scriptedPressureFetcher{results: []scriptedPressureResult{{err: errors.New("boom-1")}, {err: errors.New("boom-2")}, {pressure: 42}}},
		pressureObserver:    observer,
		controllerObserver:  controllerObserver,
	}, ctrl, logger, errCh)

	deadline := time.Now().Add(100 * time.Millisecond)
	for time.Now().Before(deadline) {
		observer.mu.Lock()
		done := len(observer.failures) >= 2 && len(observer.samples) >= 1
		observer.mu.Unlock()
		if done {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	time.Sleep(10 * time.Millisecond)

	select {
	case err := <-errCh:
		t.Fatalf("unexpected err: %v", err)
	default:
	}

	observer.mu.Lock()
	defer observer.mu.Unlock()

	if len(observer.failures) < 2 || observer.failures[0] != 1 || observer.failures[1] != 2 {
		t.Fatalf("expected failure streak callbacks [1 2], got %#v", observer.failures)
	}
	if len(observer.samples) == 0 || observer.samples[0] != 42 {
		t.Fatalf("expected first pressure sample callback 42, got %#v", observer.samples)
	}
	if observer.failureTimes[0].IsZero() || observer.failureTimes[1].IsZero() {
		t.Fatalf("expected failure timestamps, got %#v", observer.failureTimes)
	}
	if observer.sampleTimes[0].IsZero() {
		t.Fatalf("expected sample timestamp, got %#v", observer.sampleTimes)
	}

	controllerObserver.mu.Lock()
	defer controllerObserver.mu.Unlock()

	if controllerObserver.callbacks != 0 {
		t.Fatalf("expected no controller callbacks from pressure poll path, got %d with statuses %#v", controllerObserver.callbacks, controllerObserver.statuses)
	}
}

func TestEmitScanResultEvents_WhenProgressStepReached_EmitsProgressSnapshot(t *testing.T) {
	stdout := &bytes.Buffer{}
	logOut := &lockedBuffer{}
	logger := newLogger("info", true, logOut)
	ctrl := speedctrl.NewController()
	ch := &task.Chunk{
		CIDR:         "10.0.0.0/24",
		ScannedCount: 1,
		TotalCount:   4,
	}
	runtimes := []*chunkRuntime{{
		state:   ch,
		tracker: newChunkStateTracker(ch),
	}}
	summary := &resultSummary{written: 2}

	emitScanResultEvents(stdout, logger, ctrl, 2, runtimes, scanResult{
		chunkIdx: 0,
		record: writer.Record{
			IP:         "10.0.0.1",
			IPCidr:     "10.0.0.0/24",
			Port:       80,
			Status:     "open",
			ResponseMS: 7,
		},
	}, summary, false)

	if !strings.Contains(stdout.String(), "progress cidr=10.0.0.0/24 scanned=1/4 paused=false") {
		t.Fatalf("expected progress snapshot on stdout, got %s", stdout.String())
	}
	logs := logOut.String()
	for _, required := range []string{
		"\"state_transition\":\"scanned\"",
		"\"state_transition\":\"progress\"",
		"\"scanned_count\":1",
		"\"total_count\":4",
		"\"completion_rate\":0.25",
	} {
		if !strings.Contains(logs, required) {
			t.Fatalf("missing %s in logs: %s", required, logs)
		}
	}
}

func TestEmitCompletionSummary_WhenResultsMixed_EmitsOutcomeBreakdown(t *testing.T) {
	logOut := &lockedBuffer{}
	logger := newLogger("info", true, logOut)

	emitCompletionSummary(logger, resultSummary{
		written:      3,
		openCount:    1,
		closeCount:   1,
		timeoutCount: 1,
	}, time.Now().Add(-20*time.Millisecond), context.DeadlineExceeded)

	logs := logOut.String()
	for _, required := range []string{
		"\"state_transition\":\"completion_summary\"",
		"\"total_tasks\":3",
		"\"open_count\":1",
		"\"close_count\":1",
		"\"timeout_count\":1",
		"\"success\":false",
		"\"error_cause\":\"deadline_exceeded\"",
	} {
		if !strings.Contains(logs, required) {
			t.Fatalf("missing %s in logs: %s", required, logs)
		}
	}
}
