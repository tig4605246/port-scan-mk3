package testkit

import (
	"fmt"
	"sort"

	"github.com/xuxiping/port-scan-mk3/pkg/state"
	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

// LoadResumeState loads resume chunks from a state file.
func LoadResumeState(path string) ([]task.Chunk, error) {
	return state.Load(path)
}

// AssertResumeChunksEqual compares chunk state semantically, ignoring chunk order.
func AssertResumeChunksEqual(want, got []task.Chunk) error {
	if len(want) != len(got) {
		return fmt.Errorf("resume chunk count mismatch: want=%d got=%d", len(want), len(got))
	}

	wantMap := make(map[string]task.Chunk, len(want))
	for _, ch := range want {
		wantMap[ch.CIDR] = ch
	}
	for _, actual := range got {
		expected, ok := wantMap[actual.CIDR]
		if !ok {
			return fmt.Errorf("unexpected resume chunk cidr=%s", actual.CIDR)
		}
		if err := compareChunk(expected, actual); err != nil {
			return fmt.Errorf("cidr=%s: %w", actual.CIDR, err)
		}
	}
	return nil
}

func compareChunk(want, got task.Chunk) error {
	if want.NextIndex != got.NextIndex {
		return fmt.Errorf("next_index mismatch: want=%d got=%d", want.NextIndex, got.NextIndex)
	}
	if want.ScannedCount != got.ScannedCount {
		return fmt.Errorf("scanned_count mismatch: want=%d got=%d", want.ScannedCount, got.ScannedCount)
	}
	if want.TotalCount != got.TotalCount {
		return fmt.Errorf("total_count mismatch: want=%d got=%d", want.TotalCount, got.TotalCount)
	}
	if want.Status != got.Status {
		return fmt.Errorf("status mismatch: want=%s got=%s", want.Status, got.Status)
	}
	if !samePorts(want.Ports, got.Ports) {
		return fmt.Errorf("ports mismatch: want=%v got=%v", want.Ports, got.Ports)
	}
	return nil
}

func samePorts(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aa := append([]string(nil), a...)
	bb := append([]string(nil), b...)
	sort.Strings(aa)
	sort.Strings(bb)
	for i := range aa {
		if aa[i] != bb[i] {
			return false
		}
	}
	return true
}
