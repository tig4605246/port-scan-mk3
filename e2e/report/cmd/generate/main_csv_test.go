package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_FromCSV(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "scan_results.csv")
	outDir := filepath.Join(dir, "out")
	data := "ip,ip_cidr,port,status,response_time_ms,fab_name,cidr_name\n" +
		"172.28.0.10,172.28.0.10/32,8080,open,1,fab1,open-target\n" +
		"172.28.0.11,172.28.0.11/32,8080,close,0,fab2,closed-target\n"
	if err := os.WriteFile(csvPath, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	stderr := &bytes.Buffer{}
	code := run([]string{"-out", outDir, "-csv", csvPath}, &bytes.Buffer{}, stderr)
	if code != 0 {
		t.Fatalf("expected 0, got %d stderr=%s", code, stderr.String())
	}

	txt, err := os.ReadFile(filepath.Join(outDir, "report.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(txt), "Open=1") || !strings.Contains(string(txt), "Closed=1") {
		t.Fatalf("unexpected report: %s", string(txt))
	}
}
