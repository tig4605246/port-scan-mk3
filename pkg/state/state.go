package state

import (
	"encoding/json"
	"os"

	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

type PreScanPingState struct {
	Enabled            bool     `json:"enabled"`
	TimeoutMS          int      `json:"timeout_ms"`
	UnreachableIPv4U32 []uint32 `json:"unreachable_ipv4_u32,omitempty"`
}

type Snapshot struct {
	Chunks      []task.Chunk     `json:"chunks"`
	PreScanPing PreScanPingState `json:"pre_scan_ping,omitempty"`
}

// SaveSnapshot writes resume state as the current JSON envelope.
func SaveSnapshot(path string, snap Snapshot) error {
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// LoadSnapshot reads resume state from either the current envelope or legacy chunk array JSON.
func LoadSnapshot(path string) (Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, err
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err == nil {
		return snap, nil
	}

	var chunks []task.Chunk
	if err := json.Unmarshal(data, &chunks); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Chunks: chunks}, nil
}

// Save writes chunk resume state as JSON.
func Save(path string, chunks []task.Chunk) error {
	return SaveSnapshot(path, Snapshot{Chunks: chunks})
}

// Load reads chunk resume state from JSON file.
func Load(path string) ([]task.Chunk, error) {
	snap, err := LoadSnapshot(path)
	if err != nil {
		return nil, err
	}
	return snap.Chunks, nil
}
