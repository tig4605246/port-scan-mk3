package writer

import (
	"bytes"
	"strings"
	"testing"
)

func TestOpenOnlyWriter_WhenMixedStatusesProvided_WritesOnlyOpenRows(t *testing.T) {
	buf := &bytes.Buffer{}
	inner := NewCSVWriter(buf)
	w := NewOpenOnlyWriter(inner)

	if err := w.Write(Record{IP: "10.0.0.1", IPCidr: "10.0.0.0/24", Port: 80, Status: "open"}); err != nil {
		t.Fatalf("write open failed: %v", err)
	}
	if err := w.Write(Record{IP: "10.0.0.2", IPCidr: "10.0.0.0/24", Port: 80, Status: "close"}); err != nil {
		t.Fatalf("write closed failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected header + 1 open row, got %d lines: %s", len(lines), buf.String())
	}
	if !strings.Contains(lines[1], "10.0.0.1,10.0.0.0/24,80,open") {
		t.Fatalf("unexpected open row: %s", lines[1])
	}
}

func TestOpenOnlyWriter_WhenWriteHeaderCalled_WritesCSVHeader(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewOpenOnlyWriter(NewCSVWriter(buf))
	if err := w.WriteHeader(); err != nil {
		t.Fatalf("write header failed: %v", err)
	}
	if !strings.Contains(buf.String(), "ip,ip_cidr,port,status,response_time_ms,fab_name,cidr_name") {
		t.Fatalf("unexpected header: %s", buf.String())
	}
}

func TestOpenOnlyWriter_WhenWriterIsNil_ReturnsNilError(t *testing.T) {
	var w *OpenOnlyWriter
	if err := w.WriteHeader(); err != nil {
		t.Fatalf("expected nil-safe header write, got %v", err)
	}
	if err := w.Write(Record{Status: "open"}); err != nil {
		t.Fatalf("expected nil-safe write, got %v", err)
	}

	w = &OpenOnlyWriter{}
	if err := w.WriteHeader(); err != nil {
		t.Fatalf("expected nil inner header write, got %v", err)
	}
	if err := w.Write(Record{Status: "open"}); err != nil {
		t.Fatalf("expected nil inner write, got %v", err)
	}
}
