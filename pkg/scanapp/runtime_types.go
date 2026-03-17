package scanapp

import (
	"github.com/xuxiping/port-scan-mk3/pkg/ratelimit"
	"github.com/xuxiping/port-scan-mk3/pkg/task"
	"github.com/xuxiping/port-scan-mk3/pkg/writer"
)

type targetMeta struct {
	fabName           string
	cidrName          string
	serviceLabel      string
	decision          string
	policyID          string
	reason            string
	executionKey      string
	srcIP             string
	srcNetworkSegment string
}

type chunkRuntime struct {
	ipCidr  string
	ports   []int
	targets []scanTarget
	state   *task.Chunk
	tracker *chunkStateTracker
	bkt     *ratelimit.LeakyBucket
}

type scanTarget struct {
	ip     string
	ipCidr string
	meta   targetMeta
}

type scanTask struct {
	chunkIdx int
	ipCidr   string
	ip       string
	port     int
	meta     targetMeta
}

type scanResult struct {
	chunkIdx int
	record   writer.Record
}
