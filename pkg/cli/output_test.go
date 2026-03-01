package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteValidation_HumanAndJSON(t *testing.T) {
	out := &bytes.Buffer{}
	if err := WriteValidation(out, "human", true, "ok"); err != nil {
		t.Fatalf("human write failed: %v", err)
	}
	if !strings.Contains(out.String(), "valid=true") {
		t.Fatalf("unexpected human output: %s", out.String())
	}

	out.Reset()
	if err := WriteValidation(out, "json", false, "bad"); err != nil {
		t.Fatalf("json write failed: %v", err)
	}
	if !strings.Contains(out.String(), `"valid":false`) {
		t.Fatalf("unexpected json output: %s", out.String())
	}
}
