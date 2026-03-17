package main

import (
	"bytes"
	"context"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
	"github.com/xuxiping/port-scan-mk3/pkg/scanapp"
)

func TestRunMain_ScanWritesCSV(t *testing.T) {
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

	if err := os.WriteFile(cidrFile, []byte("fab_name,ip,ip_cidr,cidr_name\nfab1,127.0.0.1,127.0.0.1/32,loopback\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(portFile, []byte(strconv.Itoa(openPort)+"/tcp\n1/tcp\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := runMain([]string{
		"scan",
		"-cidr-file", cidrFile,
		"-port-file", portFile,
		"-output", outFile,
		"-workers", "1",
		"-delay", "0ms",
		"-timeout", "100ms",
		"-disable-api=true",
	}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", code, stderr.String())
	}

	scanOutputPath := mustFindOneMain(t, filepath.Join(tmp, "scan_results-*.csv"))
	data, err := os.ReadFile(scanOutputPath)
	if err != nil {
		t.Fatalf("failed to read output csv: %v", err)
	}
	out := string(data)
	if !strings.Contains(out, "ip,ip_cidr,port,status,response_time_ms,fab_name,cidr_name") {
		t.Fatalf("missing header: %s", out)
	}
	if !strings.Contains(out, "127.0.0.1,127.0.0.1/32,"+strconv.Itoa(openPort)+",open") {
		t.Fatalf("missing open row: %s", out)
	}
	if !strings.Contains(out, "127.0.0.1,127.0.0.1/32,1,close") && !strings.Contains(out, "127.0.0.1,127.0.0.1/32,1,close(timeout)") {
		t.Fatalf("missing close row: %s", out)
	}
}

func TestScanApp_CancelSavesResumeState(t *testing.T) {
	tmp := t.TempDir()
	cidrFile := filepath.Join(tmp, "cidr.csv")
	portFile := filepath.Join(tmp, "ports.csv")
	outFile := filepath.Join(tmp, "out.csv")
	resumeFile := filepath.Join(tmp, "resume_state.json")

	if err := os.WriteFile(cidrFile, []byte("fab_name,ip,ip_cidr,cidr_name\nfab1,127.0.0.1/24,127.0.0.1/24,loopback\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(portFile, []byte("1/tcp\n2/tcp\n3/tcp\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		CIDRFile:         cidrFile,
		PortFile:         portFile,
		Output:           outFile,
		Timeout:          50 * time.Millisecond,
		Delay:            5 * time.Millisecond,
		BucketRate:       1,
		BucketCapacity:   1,
		Workers:          1,
		PressureAPI:      "http://127.0.0.1:1",
		PressureInterval: 10 * time.Second,
		DisableAPI:       true,
		Resume:           "",
		LogLevel:         "error",
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	err := scanapp.Run(ctx, cfg, &bytes.Buffer{}, &bytes.Buffer{}, scanapp.RunOptions{ResumeStatePath: resumeFile})
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if _, statErr := os.Stat(resumeFile); statErr != nil {
		t.Fatalf("expected resume state file, got err=%v", statErr)
	}
}

func TestRunMain_ScanWritesOpenedResultsCSV(t *testing.T) {
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

	code := runMain([]string{
		"scan",
		"-cidr-file", cidrFile,
		"-port-file", portFile,
		"-output", outFile,
		"-workers", "1",
		"-delay", "0ms",
		"-timeout", "100ms",
		"-disable-api=true",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	openOnlyFile := mustFindOneMain(t, filepath.Join(tmp, "opened_results-*.csv"))
	data, err := os.ReadFile(openOnlyFile)
	if err != nil {
		t.Fatalf("failed to read opened_results.csv: %v", err)
	}
	out := string(data)
	if !strings.Contains(out, "ip,ip_cidr,port,status,response_time_ms,fab_name,cidr_name") {
		t.Fatalf("missing header: %s", out)
	}
	if !strings.Contains(out, ",open,") {
		t.Fatalf("expected open row: %s", out)
	}
	if strings.Contains(out, ",close,") || strings.Contains(out, "close(timeout)") {
		t.Fatalf("opened_results.csv must contain open rows only: %s", out)
	}
}

func TestRunMain_WhenScanConfigParseFails_ReturnsExit2AndWritesStderr(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := runMain([]string{"scan", "-cidr-file", "", "-port-file", ""}, stdout, stderr)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d stderr=%s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "-cidr-file and -port-file are required") {
		t.Fatalf("expected parse error on stderr, got %s", stderr.String())
	}
}

func TestRunMain_ScanSuccess_WritesTimestampedBatchPairInRequestedDirectory(t *testing.T) {
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
	requestedOutput := filepath.Join(tmp, "custom-name.csv")

	if err := os.WriteFile(cidrFile, []byte("fab_name,ip,ip_cidr,cidr_name\nfab1,127.0.0.1,127.0.0.1/32,loopback\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(portFile, []byte(strconv.Itoa(openPort)+"/tcp\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := runMain([]string{
		"scan",
		"-cidr-file", cidrFile,
		"-port-file", portFile,
		"-output", requestedOutput,
		"-workers", "1",
		"-delay", "0ms",
		"-timeout", "100ms",
		"-disable-api=true",
	}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout for single-result scan, got %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "[INFO] scan_result") {
		t.Fatalf("expected scan_result log on stderr, got %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "state_transition:completion_summary") {
		t.Fatalf("expected completion summary log on stderr, got %s", stderr.String())
	}

	scanPath := mustFindOneMain(t, filepath.Join(tmp, "scan_results-*.csv"))
	openPath := mustFindOneMain(t, filepath.Join(tmp, "opened_results-*.csv"))
	scanSuffix, openSuffix := mustBatchPairSuffix(t, scanPath, openPath)
	if scanSuffix != openSuffix {
		t.Fatalf("expected matching batch suffixes, got scan=%s open=%s", scanSuffix, openSuffix)
	}
	if _, err := os.Stat(requestedOutput); !os.IsNotExist(err) {
		t.Fatalf("expected requested output path to be used as directory hint only, err=%v", err)
	}
}

func mustFindOneMain(t *testing.T, pattern string) string {
	t.Helper()
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob failed for %s: %v", pattern, err)
	}
	sort.Strings(matches)
	if len(matches) != 1 {
		t.Fatalf("expected exactly one match for %s, got %d (%v)", pattern, len(matches), matches)
	}
	return matches[0]
}
