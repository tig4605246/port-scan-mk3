package state

import (
	"encoding/json"
	"os"

	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

func Save(path string, chunks []task.Chunk) error {
	data, err := json.MarshalIndent(chunks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func Load(path string) ([]task.Chunk, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var chunks []task.Chunk
	if err := json.Unmarshal(data, &chunks); err != nil {
		return nil, err
	}
	return chunks, nil
}
