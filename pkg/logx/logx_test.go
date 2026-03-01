package logx

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogJSON_EmitsJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	LogJSON(buf, "info", "hello", map[string]any{"k": "v"})
	out := buf.String()
	if !strings.Contains(out, "\"level\":\"info\"") {
		t.Fatalf("missing level: %s", out)
	}
	if !strings.Contains(out, "\"msg\":\"hello\"") {
		t.Fatalf("missing msg: %s", out)
	}
}
