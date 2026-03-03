package writer

import (
	"bytes"
	"strings"
	"testing"
)

func TestCSVWriter_WhenWritingFirstRecord_WritesHeaderAndRow(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewCSVWriter(buf)

	r := Record{
		IP: "192.168.1.1", IPCidr: "192.168.1.0/24", Port: 80, Status: "open", ResponseMS: 12,
		FabName: "fab1", CIDRName: "office",
	}
	if err := w.Write(r); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if !strings.Contains(buf.String(), "ip,ip_cidr,port,status,response_time_ms") {
		t.Fatalf("header missing: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "192.168.1.1,192.168.1.0/24,80,open,12") {
		t.Fatalf("row missing: %s", buf.String())
	}
}

func TestCSVWriter_WhenIPCidrMissing_UsesCIDRFallbackAndHeaderWrittenOnce(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewCSVWriter(buf)

	if err := w.WriteHeader(); err != nil {
		t.Fatalf("write header failed: %v", err)
	}
	if err := w.Write(Record{
		IP:      "10.0.0.1",
		CIDR:    "10.0.0.0/24",
		Port:    22,
		Status:  "open",
		FabName: "fab-x",
	}); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if err := w.WriteHeader(); err != nil {
		t.Fatalf("second write header failed: %v", err)
	}
	out := buf.String()
	if strings.Count(out, "ip,ip_cidr,port,status,response_time_ms,fab_name,cidr_name") != 1 {
		t.Fatalf("header should appear once, got: %s", out)
	}
	if !strings.Contains(out, "10.0.0.1,10.0.0.0/24,22,open") {
		t.Fatalf("expected CIDR fallback row, got: %s", out)
	}
}
