package scanapp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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
	"github.com/xuxiping/port-scan-mk3/pkg/input"
	"github.com/xuxiping/port-scan-mk3/pkg/speedctrl"
	"github.com/xuxiping/port-scan-mk3/pkg/state"
	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

func TestRun_ResumeFromStateFile(t *testing.T) {
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
	outFile := filepath.Join(tmp, "out.csv")
	resumeFile := filepath.Join(tmp, "resume.json")

	if err := os.WriteFile(cidrFile, []byte("fab_name,ip,ip_cidr,cidr_name\nfab1,127.0.0.1,127.0.0.1/32,loopback\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(portFile, []byte(strconv.Itoa(openPort)+"/tcp\n1/tcp\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	initial := []task.Chunk{{
		CIDR:         "127.0.0.1/32",
		CIDRName:     "loopback",
		Ports:        []string{strconv.Itoa(openPort) + "/tcp", "1/tcp"},
		NextIndex:    1,
		ScannedCount: 1,
		TotalCount:   2,
		Status:       "scanning",
	}}
	if err := state.Save(resumeFile, initial); err != nil {
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
		Resume:           resumeFile,
		LogLevel:         "error",
	}
	if err := Run(context.Background(), cfg, &bytes.Buffer{}, &bytes.Buffer{}, RunOptions{DisableKeyboard: true}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	if !strings.Contains(out, "127.0.0.1,127.0.0.1/32,1,close") && !strings.Contains(out, "127.0.0.1,127.0.0.1/32,1,close(timeout)") {
		t.Fatalf("expected resumed scan row, got: %s", out)
	}
	if strings.Contains(out, "127.0.0.1,127.0.0.1/32,"+strconv.Itoa(openPort)+",open") {
		t.Fatalf("did not expect already-scanned port row, got: %s", out)
	}
}

func TestRun_PressureAPIFailsThreeTimes(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	defer api.Close()

	tmp := t.TempDir()
	cidrFile := filepath.Join(tmp, "cidr.csv")
	portFile := filepath.Join(tmp, "ports.csv")
	outFile := filepath.Join(tmp, "out.csv")
	resumeFile := filepath.Join(tmp, "resume_state.json")

	if err := os.WriteFile(cidrFile, []byte("fab_name,ip,ip_cidr,cidr_name\nfab1,127.0.0.0/24,127.0.0.0/24,loopback\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(portFile, []byte("1/tcp\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		CIDRFile:         cidrFile,
		PortFile:         portFile,
		Output:           outFile,
		Timeout:          20 * time.Millisecond,
		Delay:            0,
		BucketRate:       1,
		BucketCapacity:   1,
		Workers:          1,
		PressureAPI:      api.URL,
		PressureInterval: 5 * time.Millisecond,
		DisableAPI:       false,
		LogLevel:         "error",
	}

	err := Run(context.Background(), cfg, &bytes.Buffer{}, &bytes.Buffer{}, RunOptions{
		DisableKeyboard: true,
		ResumeStatePath: resumeFile,
		PressureHTTP:    &http.Client{Timeout: 500 * time.Millisecond},
	})
	if err == nil {
		t.Fatal("expected api failure error")
	}
	if !strings.Contains(err.Error(), "pressure api failed 3 times") {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(resumeFile); statErr != nil {
		t.Fatalf("expected resume state on fatal api error, got: %v", statErr)
	}
}

func TestFetchPressure(t *testing.T) {
	okAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintln(w, `{"pressure":95}`)
	}))
	defer okAPI.Close()

	n, err := fetchPressure(&http.Client{Timeout: time.Second}, okAPI.URL)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if n != 95 {
		t.Fatalf("unexpected pressure: %d", n)
	}

	strAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintln(w, `{"pressure":"88"}`)
	}))
	defer strAPI.Close()

	n, err = fetchPressure(&http.Client{Timeout: time.Second}, strAPI.URL)
	if err != nil || n != 88 {
		t.Fatalf("unexpected parse result n=%d err=%v", n, err)
	}

	badAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	defer badAPI.Close()
	if _, err := fetchPressure(&http.Client{Timeout: time.Second}, badAPI.URL); err == nil {
		t.Fatal("expected status error")
	}
}

func TestParsePortRows(t *testing.T) {
	ports, err := parsePortRows([]string{"80/tcp", "443/tcp"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(ports) != 2 || ports[0] != 80 || ports[1] != 443 {
		t.Fatalf("unexpected ports: %#v", ports)
	}

	if _, err := parsePortRows([]string{"53/udp"}); err == nil {
		t.Fatal("expected invalid protocol error")
	}
}

func TestBuildRuntime_DefaultPortsFromInput(t *testing.T) {
	_, ipNet, _ := net.ParseCIDR("10.0.0.0/30")
	records := []input.CIDRRecord{{
		FabName:  "fab1",
		CIDR:     "10.0.0.0/30",
		CIDRName: "x",
		Net:      ipNet,
	}}
	chunks := []task.Chunk{{
		CIDR:       "10.0.0.0/30",
		CIDRName:   "x",
		Ports:      nil,
		TotalCount: 0,
	}}
	ports := []input.PortSpec{{Number: 80, Proto: "tcp", Raw: "80/tcp"}}

	rts, err := buildRuntime(chunks, records, ports, config.Config{
		BucketRate:     10,
		BucketCapacity: 10,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(rts) != 1 {
		t.Fatalf("unexpected runtime len: %d", len(rts))
	}
	if rts[0].state.TotalCount != 4 {
		t.Fatalf("unexpected total count: %d", rts[0].state.TotalCount)
	}
}

func TestShouldSaveOnDispatchErr(t *testing.T) {
	if shouldSaveOnDispatchErr(nil) {
		t.Fatal("expected false for nil err")
	}
	if !shouldSaveOnDispatchErr(context.Canceled) {
		t.Fatal("expected true for context canceled")
	}
	if !shouldSaveOnDispatchErr(context.DeadlineExceeded) {
		t.Fatal("expected true for deadline exceeded")
	}
	if shouldSaveOnDispatchErr(errors.New("other")) {
		t.Fatal("expected false for other err")
	}
}

func TestLogger_TextAndJSON(t *testing.T) {
	textOut := &bytes.Buffer{}
	l := newLogger("debug", false, textOut)
	l.debugf("x=%d", 1)
	if !strings.Contains(textOut.String(), "[DEBUG] x=1") {
		t.Fatalf("unexpected text log: %s", textOut.String())
	}

	jsonOut := &bytes.Buffer{}
	l = newLogger("info", true, jsonOut)
	l.infof("hello")
	if !strings.Contains(jsonOut.String(), `"level":"info"`) {
		t.Fatalf("unexpected json log: %s", jsonOut.String())
	}
}

func TestPollPressureAPI_PauseResumeTransition(t *testing.T) {
	values := []int{95, 20}
	var mu sync.Mutex
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		v := values[0]
		if len(values) > 1 {
			values = values[1:]
		}
		mu.Unlock()
		_, _ = fmt.Fprintf(w, `{"pressure":%d}`, v)
	}))
	defer api.Close()

	cfg := config.Config{
		PressureAPI:      api.URL,
		PressureInterval: 5 * time.Millisecond,
	}
	ctrl := speedctrl.NewController()
	logOut := &lockedBuffer{}
	logger := newLogger("info", false, logOut)
	errCh := make(chan error, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go pollPressureAPI(ctx, cfg, RunOptions{PressureLimit: 90, PressureHTTP: &http.Client{Timeout: time.Second}}, ctrl, logger, errCh)

	time.Sleep(40 * time.Millisecond)
	cancel()
	time.Sleep(10 * time.Millisecond)

	select {
	case err := <-errCh:
		t.Fatalf("unexpected err: %v", err)
	default:
	}
	if ctrl.IsPaused() {
		t.Fatal("expected resumed after pressure drop")
	}
	if !strings.Contains(logOut.String(), "掃描已自動暫停") || !strings.Contains(logOut.String(), "掃描已自動恢復") {
		t.Fatalf("expected pause/resume logs, got: %s", logOut.String())
	}
}

func TestStartManualPauseMonitor_LogsStateChange(t *testing.T) {
	ctrl := speedctrl.NewController()
	out := &lockedBuffer{}
	logger := newLogger("info", false, out)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startManualPauseMonitor(ctx, ctrl, logger)
	time.Sleep(50 * time.Millisecond)
	ctrl.SetManualPaused(true)
	time.Sleep(250 * time.Millisecond)
	ctrl.SetManualPaused(false)
	time.Sleep(250 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)

	logs := out.String()
	if !strings.Contains(logs, "掃描已手動暫停") || !strings.Contains(logs, "掃描已手動恢復") {
		t.Fatalf("expected manual pause/resume logs, got: %s", logs)
	}
}

func TestRun_ScansOnlyIPsListedByIPColumn(t *testing.T) {
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

	if err := os.WriteFile(cidrFile, []byte(
		"fab_name,ip,ip_cidr,cidr_name\n"+
			"fab1,127.0.0.1,127.0.0.0/30,subset\n",
	), 0o644); err != nil {
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
		LogLevel:         "error",
	}
	if err := Run(context.Background(), cfg, &bytes.Buffer{}, &bytes.Buffer{}, RunOptions{DisableKeyboard: true}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	outBytes, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(outBytes)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected header + 2 rows for 1 listed ip x 2 ports, got %d lines: %s", len(lines), string(outBytes))
	}
	if strings.Contains(string(outBytes), "127.0.0.2") || strings.Contains(string(outBytes), "127.0.0.3") {
		t.Fatalf("unexpected non-listed ip in output: %s", string(outBytes))
	}
}

func TestRun_WritesOpenedResultsCSV(t *testing.T) {
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
	openOnlyPath := filepath.Join(tmp, "opened_results.csv")

	if err := os.WriteFile(cidrFile, []byte(
		"fab_name,ip,ip_cidr,cidr_name\n"+
			"fab1,127.0.0.1,127.0.0.1/32,loopback\n",
	), 0o644); err != nil {
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
		LogLevel:         "error",
	}
	if err := Run(context.Background(), cfg, &bytes.Buffer{}, &bytes.Buffer{}, RunOptions{DisableKeyboard: true}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	openOnlyBytes, err := os.ReadFile(openOnlyPath)
	if err != nil {
		t.Fatalf("read opened_results.csv failed: %v", err)
	}
	openOnly := string(openOnlyBytes)
	if !strings.Contains(openOnly, ",open,") {
		t.Fatalf("expected at least one open record, got: %s", openOnly)
	}
	if strings.Contains(openOnly, ",close,") || strings.Contains(openOnly, "close(timeout)") {
		t.Fatalf("opened_results.csv must include open records only, got: %s", openOnly)
	}
}
