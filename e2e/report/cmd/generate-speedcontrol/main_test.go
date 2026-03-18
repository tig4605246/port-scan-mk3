package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRun_GeneratesSpeedControlReportArtifacts(t *testing.T) {
	outDir := t.TempDir()
	stderr := &bytes.Buffer{}

	code := run([]string{"-out", outDir}, bytes.NewBuffer(nil), stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d stderr=%s", code, stderr.String())
	}

	for _, name := range []string{"report.md", "report.html", "raw_metrics.json"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("missing artifact %s: %v", name, err)
		}
	}
}
