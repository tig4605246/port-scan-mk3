package state

import (
	"path/filepath"
	"testing"

	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

func TestSaveAndLoad_WhenStatePersisted_RestoresChunkFields(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "resume_state.json")

	chunks := []task.Chunk{{CIDR: "10.0.0.0/30", NextIndex: 2, TotalCount: 8, Status: "scanning"}}
	if err := Save(file, chunks); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, err := Load(file)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if got[0].NextIndex != 2 {
		t.Fatalf("next index mismatch: %d", got[0].NextIndex)
	}
}
