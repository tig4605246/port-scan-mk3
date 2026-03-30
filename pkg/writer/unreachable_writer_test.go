package writer

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
)

func TestUnreachableWriter_WriteHeader_WritesFixedHeaderOnce(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewUnreachableWriter(buf)

	if err := w.WriteHeader(); err != nil {
		t.Fatalf("first write header failed: %v", err)
	}
	if err := w.WriteHeader(); err != nil {
		t.Fatalf("second write header failed: %v", err)
	}

	rows, err := csv.NewReader(strings.NewReader(buf.String())).ReadAll()
	if err != nil {
		t.Fatalf("csv parse failed: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	want := []string{
		"ip",
		"ip_cidr",
		"status",
		"reason",
		"fab_name",
		"cidr_name",
		"service_label",
		"decision",
		"matched_policy_id",
		"execution_key",
		"src_ip",
		"src_network_segment",
	}
	for i, col := range want {
		if rows[0][i] != col {
			t.Fatalf("header[%d] mismatch: got=%s want=%s", i, rows[0][i], col)
		}
	}
}
