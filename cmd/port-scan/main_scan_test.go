package main

import (
	"bytes"
	"context"
	"net"
	"os"
	"path/filepath"
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

	if err := os.WriteFile(cidrFile, []byte("fab_name,cidr,cidr_name\nfab1,127.0.0.1/32,loopback\n"), 0o644); err != nil {
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

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read output csv: %v", err)
	}
	out := string(data)
	if !strings.Contains(out, "ip,port,status,response_time_ms,fab_name,cidr,cidr_name") {
		t.Fatalf("missing header: %s", out)
	}
	if !strings.Contains(out, "127.0.0.1,"+strconv.Itoa(openPort)+",open") {
		t.Fatalf("missing open row: %s", out)
	}
	if !strings.Contains(out, "127.0.0.1,1,close") && !strings.Contains(out, "127.0.0.1,1,close(timeout)") {
		t.Fatalf("missing close row: %s", out)
	}
}

func TestScanApp_CancelSavesResumeState(t *testing.T) {
	tmp := t.TempDir()
	cidrFile := filepath.Join(tmp, "cidr.csv")
	portFile := filepath.Join(tmp, "ports.csv")
	outFile := filepath.Join(tmp, "out.csv")
	resumeFile := filepath.Join(tmp, "resume_state.json")

	if err := os.WriteFile(cidrFile, []byte("fab_name,cidr,cidr_name\nfab1,127.0.0.1/24,loopback\n"), 0o644); err != nil {
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
