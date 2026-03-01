package writer

import (
	"bytes"
	"strings"
	"testing"
)

func TestCSVWriter_WritesHeaderAndRows(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewCSVWriter(buf)

	r := Record{
		IP: "192.168.1.1", Port: 80, Status: "open", ResponseMS: 12,
		FabName: "fab1", CIDR: "192.168.1.0/24", CIDRName: "office",
	}
	if err := w.Write(r); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if !strings.Contains(buf.String(), "ip,port,status,response_time_ms") {
		t.Fatalf("header missing: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "192.168.1.1,80,open,12") {
		t.Fatalf("row missing: %s", buf.String())
	}
}
