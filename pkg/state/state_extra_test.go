package state

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

func TestLoad_WhenJSONIsInvalid_ReturnsError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(file, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(file); err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func TestLoadSnapshot_WhenLegacyChunkArrayProvided_PreservesCompatibility(t *testing.T) {
	file := filepath.Join(t.TempDir(), "legacy.json")
	wantChunks := []task.Chunk{
		{CIDR: "10.0.0.0/30", NextIndex: 1, TotalCount: 4, Status: "paused"},
	}
	if err := os.WriteFile(file, []byte(`[{"cidr":"10.0.0.0/30","next_index":1,"total_count":4,"status":"paused"}]`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadSnapshot(file)
	if err != nil {
		t.Fatalf("load snapshot failed: %v", err)
	}
	if !reflect.DeepEqual(got.Chunks, wantChunks) {
		t.Fatalf("chunks mismatch: got %+v want %+v", got.Chunks, wantChunks)
	}
	if !reflect.DeepEqual(got.PreScanPing, PreScanPingState{}) {
		t.Fatalf("expected empty pre-scan ping state, got %+v", got.PreScanPing)
	}
}

func TestWithSIGINTCancel_WhenCancelInvoked_CancelsContext(t *testing.T) {
	ctx, cancel := WithSIGINTCancel(context.Background())
	cancel()
	select {
	case <-ctx.Done():
	default:
		t.Fatal("expected canceled context")
	}
}
