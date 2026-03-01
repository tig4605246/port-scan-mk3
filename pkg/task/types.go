package task

type Chunk struct {
	CIDR         string   `json:"cidr"`
	CIDRName     string   `json:"cidr_name"`
	Ports        []string `json:"ports"`
	NextIndex    int      `json:"next_index"`
	ScannedCount int      `json:"scanned_count"`
	TotalCount   int      `json:"total_count"`
	Status       string   `json:"status"`
}

type Task struct {
	ChunkCIDR string
	IP        string
	Port      int
	Index     int
}

func (c Chunk) Remaining() int {
	if c.TotalCount <= c.NextIndex {
		return 0
	}
	return c.TotalCount - c.NextIndex
}
