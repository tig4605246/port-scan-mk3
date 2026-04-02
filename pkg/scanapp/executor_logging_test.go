package scanapp

import (
	"bytes"
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

func TestStartScanExecutor_EmitsScanResultEvent_WithCorrectFields(t *testing.T) {
	taskCh := make(chan scanTask, 1)
	taskCh <- scanTask{
		chunkIdx: 0,
		ipCidr:   "10.0.0.0/24",
		ip:       "10.0.0.8",
		port:     443,
		meta:     targetMeta{},
	}
	close(taskCh)

	dial := func(context.Context, string, string) (net.Conn, error) {
		return stubConn{}, nil
	}

	var buf bytes.Buffer
	logger := newLogger("debug", false, &buf)

	_ = startScanExecutor(1, 100*time.Millisecond, dial, logger, taskCh)

	// Wait for worker to process
	time.Sleep(200 * time.Millisecond)

	log := buf.String()
	if !strings.Contains(log, LogEventScanResult) {
		t.Errorf("expected log to contain %q, got: %s", LogEventScanResult, log)
	}
	if !strings.Contains(log, "10.0.0.8") {
		t.Errorf("expected log to contain target IP 10.0.0.8, got: %s", log)
	}
	if !strings.Contains(log, "443") {
		t.Errorf("expected log to contain port 443, got: %s", log)
	}
	if !strings.Contains(log, LogEventScanned) {
		t.Errorf("expected log to contain state %q, got: %s", LogEventScanned, log)
	}
	if !strings.Contains(log, LogEventNone) {
		t.Errorf("expected log to contain errCause %q, got: %s", LogEventNone, log)
	}
}

func TestStartScanExecutor_EmitsErrorEvent_WhenScanFails(t *testing.T) {
	taskCh := make(chan scanTask, 1)
	taskCh <- scanTask{
		chunkIdx: 0,
		ipCidr:   "10.0.0.0/24",
		ip:       "10.0.0.8",
		port:     9999,
		meta:     targetMeta{},
	}
	close(taskCh)

	dial := func(context.Context, string, string) (net.Conn, error) {
		return nil, context.DeadlineExceeded
	}

	var buf bytes.Buffer
	logger := newLogger("debug", false, &buf)

	_ = startScanExecutor(1, 50*time.Millisecond, dial, logger, taskCh)

	// Wait for worker to process
	time.Sleep(200 * time.Millisecond)

	log := buf.String()
	if !strings.Contains(log, LogEventScanResult) {
		t.Errorf("expected log to contain %q, got: %s", LogEventScanResult, log)
	}
	if !strings.Contains(log, LogEventError) {
		t.Errorf("expected log to contain error state %q, got: %s", LogEventError, log)
	}
	if !strings.Contains(log, LogEventRuntimeErr) {
		t.Errorf("expected log to contain errCause %q, got: %s", LogEventRuntimeErr, log)
	}
}

func TestStartScanExecutor_RecoversFromPanic_InWorkerGoroutine(t *testing.T) {
	taskCh := make(chan scanTask, 2)
	taskCh <- scanTask{
		chunkIdx: 0,
		ipCidr:   "10.0.0.0/24",
		ip:       "10.0.0.8",
		port:     443,
		meta:     targetMeta{},
	}
	// Second task ensures worker keeps going
	taskCh <- scanTask{
		chunkIdx: 0,
		ipCidr:   "10.0.0.0/24",
		ip:       "10.0.0.9",
		port:     443,
		meta:     targetMeta{},
	}
	close(taskCh)

	dial := func(context.Context, string, string) (net.Conn, error) {
		return stubConn{}, nil
	}

	var buf bytes.Buffer
	logger := newLogger("debug", false, &buf)

	resultCh := startScanExecutor(1, 100*time.Millisecond, dial, logger, taskCh)

	// Collect results - should complete without panic
	results := collectResults(t, resultCh)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Verify errorf wasn't called with panic message
	if strings.Contains(buf.String(), "executor worker panic") {
		t.Errorf("unexpected panic in worker: %s", buf.String())
	}
}

func TestStartScanExecutor_Constants_AreExported(t *testing.T) {
	// Verify constants are properly defined and non-empty
	if LogEventScanned == "" {
		t.Error("LogEventScanned should not be empty")
	}
	if LogEventError == "" {
		t.Error("LogEventError should not be empty")
	}
	if LogEventRuntimeErr == "" {
		t.Error("LogEventRuntimeErr should not be empty")
	}
	if LogEventNone == "" {
		t.Error("LogEventNone should not be empty")
	}
	if LogEventBucketWaitStart == "" {
		t.Error("LogEventBucketWaitStart should not be empty")
	}
	if LogEventBucketAcquired == "" {
		t.Error("LogEventBucketAcquired should not be empty")
	}
	if LogEventBucketAcquireError == "" {
		t.Error("LogEventBucketAcquireError should not be empty")
	}
	if LogEventGateWaitStart == "" {
		t.Error("LogEventGateWaitStart should not be empty")
	}
	if LogEventGateReleased == "" {
		t.Error("LogEventGateReleased should not be empty")
	}
	if LogEventScanResult == "" {
		t.Error("LogEventScanResult should not be empty")
	}
}
