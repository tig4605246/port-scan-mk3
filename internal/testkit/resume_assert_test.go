package testkit

import (
	"path/filepath"
	"testing"

	"github.com/xuxiping/port-scan-mk3/pkg/state"
	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

func TestAssertResumeChunksEqual(t *testing.T) {
	want := []task.Chunk{
		{CIDR: "10.0.0.0/24", Ports: []string{"22/tcp", "80/tcp"}, NextIndex: 2, ScannedCount: 2, TotalCount: 4, Status: "scanning"},
		{CIDR: "10.0.1.0/24", Ports: []string{"443/tcp"}, NextIndex: 1, ScannedCount: 1, TotalCount: 1, Status: "completed"},
	}
	gotReordered := []task.Chunk{
		{CIDR: "10.0.1.0/24", Ports: []string{"443/tcp"}, NextIndex: 1, ScannedCount: 1, TotalCount: 1, Status: "completed"},
		{CIDR: "10.0.0.0/24", Ports: []string{"80/tcp", "22/tcp"}, NextIndex: 2, ScannedCount: 2, TotalCount: 4, Status: "scanning"},
	}
	if err := AssertResumeChunksEqual(want, gotReordered); err != nil {
		t.Fatalf("expected equal resume chunks, got err=%v", err)
	}

	gotMismatch := []task.Chunk{{CIDR: "10.0.0.0/24", Ports: []string{"22/tcp"}, NextIndex: 3, ScannedCount: 2, TotalCount: 4, Status: "scanning"}}
	if err := AssertResumeChunksEqual(want[:1], gotMismatch); err == nil {
		t.Fatalf("expected mismatch error")
	}
}

func TestLoadResumeState(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "resume_state.json")
	chunks := []task.Chunk{{CIDR: "10.0.0.0/24", Ports: []string{"22/tcp"}, NextIndex: 1, ScannedCount: 1, TotalCount: 2, Status: "scanning"}}
	if err := state.Save(path, chunks); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	got, err := LoadResumeState(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if err := AssertResumeChunksEqual(chunks, got); err != nil {
		t.Fatalf("unexpected comparison error: %v", err)
	}
}
