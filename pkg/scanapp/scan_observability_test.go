package scanapp

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
