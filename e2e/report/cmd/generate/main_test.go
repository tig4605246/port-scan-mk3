package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRun_GeneratesFiles(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "out")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := run([]string{"-out", outDir, "-total", "4", "-open", "2", "-closed", "1", "-timeout", "1"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected code 0, got %d stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(outDir, "report.html")); err != nil {
		t.Fatalf("missing report.html: %v", err)
	}
}

func TestRun_FlagError(t *testing.T) {
	code := run([]string{"-bad-flag"}, &bytes.Buffer{}, &bytes.Buffer{})
	if code != 2 {
		t.Fatalf("expected code 2, got %d", code)
	}
}
