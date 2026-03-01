package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMainValidate_JSONOutput(t *testing.T) {
	tmp := t.TempDir()
	cidr := filepath.Join(tmp, "cidr.csv")
	port := filepath.Join(tmp, "ports.csv")
	if err := os.WriteFile(cidr, []byte("fab_name,cidr,cidr_name\nfab1,10.0.0.0/30,a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(port, []byte("80/tcp\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := runMain([]string{"validate", "-cidr-file", cidr, "-port-file", port, "-format", "json"}, out, errOut)

	if code != 0 {
		t.Fatalf("exit code=%d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), `"valid":true`) {
		t.Fatalf("expected json output, got %s", out.String())
	}
}
