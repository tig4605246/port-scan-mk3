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
