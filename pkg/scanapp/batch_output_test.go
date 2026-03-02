package scanapp

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResolveBatchOutputPaths_WhenNoExistingFiles_UsesBaseTimestampNames(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 1, 30, 45, 0, time.UTC)

	scanPath, openPath, err := resolveBatchOutputPaths(filepath.Join(dir, "scan_results.csv"), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantScan := filepath.Join(dir, "scan_results-20260302T013045Z.csv")
	wantOpen := filepath.Join(dir, "opened_results-20260302T013045Z.csv")
	if scanPath != wantScan {
		t.Fatalf("scan path mismatch: got=%s want=%s", scanPath, wantScan)
	}
	if openPath != wantOpen {
		t.Fatalf("open path mismatch: got=%s want=%s", openPath, wantOpen)
	}
}

func TestResolveBatchOutputPaths_WhenExistingFilesCollide_AppendsIncrementingSuffix(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 1, 30, 45, 0, time.UTC)

	// existing first attempt
	firstScan := filepath.Join(dir, "scan_results-20260302T013045Z.csv")
	firstOpen := filepath.Join(dir, "opened_results-20260302T013045Z.csv")
	if err := os.WriteFile(firstScan, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(firstOpen, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// existing second attempt
	secondScan := filepath.Join(dir, "scan_results-20260302T013045Z-1.csv")
	secondOpen := filepath.Join(dir, "opened_results-20260302T013045Z-1.csv")
	if err := os.WriteFile(secondScan, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secondOpen, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	scanPath, openPath, err := resolveBatchOutputPaths(filepath.Join(dir, "scan_results.csv"), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantScan := filepath.Join(dir, "scan_results-20260302T013045Z-2.csv")
	wantOpen := filepath.Join(dir, "opened_results-20260302T013045Z-2.csv")
	if scanPath != wantScan || openPath != wantOpen {
		t.Fatalf("unexpected paths: scan=%s open=%s", scanPath, openPath)
	}
}
