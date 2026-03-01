package state

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_InvalidJSON(t *testing.T) {
	file := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(file, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(file); err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func TestWithSIGINTCancel_CancelFunc(t *testing.T) {
	ctx, cancel := WithSIGINTCancel(context.Background())
	cancel()
	select {
	case <-ctx.Done():
	default:
		t.Fatal("expected canceled context")
	}
}
