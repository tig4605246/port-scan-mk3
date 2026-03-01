package report

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSummarizeCSV(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "scan_results.csv")
	data := "ip,ip_cidr,port,status,response_time_ms,fab_name,cidr_name\n" +
		"172.28.0.10,172.28.0.10/32,8080,open,1,fab1,open-target\n" +
		"172.28.0.11,172.28.0.11/32,8080,close,0,fab2,closed-target\n" +
		"172.28.0.12,172.28.0.12/32,8080,close(timeout),0,fab3,timeout-target\n"
	if err := os.WriteFile(csvPath, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := SummarizeCSV(csvPath)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if s.Total != 3 || s.Open != 1 || s.Closed != 1 || s.Timeout != 1 {
		t.Fatalf("unexpected summary: %+v", s)
	}
}
