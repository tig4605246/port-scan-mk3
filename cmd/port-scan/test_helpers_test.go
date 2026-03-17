package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

type validationResponse struct {
	Valid  bool   `json:"valid"`
	Detail string `json:"detail"`
}

func mustDecodeValidationJSON(t *testing.T, raw string) validationResponse {
	t.Helper()
	var resp validationResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to decode validation json %q: %v", raw, err)
	}
	return resp
}

func mustBatchPairSuffix(t *testing.T, scanPath, openPath string) (string, string) {
	t.Helper()
	return batchSuffix(t, filepath.Base(scanPath), "scan_results-"),
		batchSuffix(t, filepath.Base(openPath), "opened_results-")
}

func batchSuffix(t *testing.T, name, prefix string) string {
	t.Helper()
	if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".csv") {
		t.Fatalf("unexpected batch file name %q", name)
	}
	return strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".csv")
}
