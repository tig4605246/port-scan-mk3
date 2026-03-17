package scanapp

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
	"github.com/xuxiping/port-scan-mk3/pkg/speedctrl"
	"github.com/xuxiping/port-scan-mk3/pkg/task"
	"github.com/xuxiping/port-scan-mk3/pkg/writer"
)

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
	}, summary)

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
