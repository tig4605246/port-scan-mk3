package scanapp

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func resolveBatchOutputPaths(outputPath string, now time.Time) (string, string, error) {
	baseDir := filepath.Dir(outputPath)
	if baseDir == "" {
		baseDir = "."
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return "", "", err
	}

	ts := now.UTC().Format("20060102T150405Z")
	for seq := 0; seq < 1_000_000; seq++ {
		suffix := ""
		if seq > 0 {
			suffix = fmt.Sprintf("-%d", seq)
		}
		scanPath := filepath.Join(baseDir, fmt.Sprintf("scan_results-%s%s.csv", ts, suffix))
		openPath := filepath.Join(baseDir, fmt.Sprintf("opened_results-%s%s.csv", ts, suffix))
		if !fileExists(scanPath) && !fileExists(openPath) {
			return scanPath, openPath, nil
		}
	}
	return "", "", fmt.Errorf("failed to allocate unique batch output paths")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
