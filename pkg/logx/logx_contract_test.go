package logx

import (
	"bytes"
	"encoding/json"
	"testing"
)

func Test_observability_contract_fields(t *testing.T) {
	buf := &bytes.Buffer{}
	LogJSON(buf, "info", "scan_result", map[string]any{
		"target":           "127.0.0.1",
		"port":             8080,
		"state_transition": "scanned",
		"error_cause":      "none",
	})

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}
	fields, ok := payload["fields"].(map[string]any)
	if !ok {
		t.Fatalf("missing fields object: %#v", payload)
	}
	for _, key := range []string{"target", "port", "state_transition", "error_cause"} {
		if _, ok := fields[key]; !ok {
			t.Fatalf("missing required field %s in %#v", key, fields)
		}
	}
}
