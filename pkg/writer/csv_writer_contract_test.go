package writer

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
)

func TestCSVWriter_WhenMetadataMissing_PreservesHeaderOrderAndEmptyMetadata(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewCSVWriter(buf)

	if err := w.Write(Record{
		IP:         "127.0.0.1",
		IPCidr:     "127.0.0.0/24",
		Port:       8080,
		Status:     "open",
		ResponseMS: 3,
		// FabName/CIDRName intentionally empty
	}); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	r := csv.NewReader(strings.NewReader(buf.String()))
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("csv parse failed: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	wantHeader := []string{
		"ip",
		"ip_cidr",
		"port",
		"status",
		"response_time_ms",
		"fab_name",
		"cidr_name",
		"service_label",
		"decision",
		"policy_id",
		"reason",
		"execution_key",
		"src_ip",
		"src_network_segment",
	}
	for i, col := range wantHeader {
		if rows[0][i] != col {
			t.Fatalf("header[%d] mismatch: got=%s want=%s", i, rows[0][i], col)
		}
	}

	for i := 5; i < len(wantHeader); i++ {
		if rows[1][i] != "" {
			t.Fatalf("expected empty metadata field %s, got %q", wantHeader[i], rows[1][i])
		}
	}
}
