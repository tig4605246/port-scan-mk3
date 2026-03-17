package scanapp

import (
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
	"github.com/xuxiping/port-scan-mk3/pkg/ratelimit"
	"github.com/xuxiping/port-scan-mk3/pkg/task"
	"github.com/xuxiping/port-scan-mk3/pkg/writer"
)

type runtimePolicy struct {
	bucketRate     int
	bucketCapacity int
}

func runtimePolicyFromConfig(cfg config.Config) runtimePolicy {
	return runtimePolicy{
		bucketRate:     cfg.BucketRate,
		bucketCapacity: cfg.BucketCapacity,
	}
}

type dispatchPolicy struct {
	delay time.Duration
}

func dispatchPolicyFromConfig(cfg config.Config) dispatchPolicy {
	return dispatchPolicy{delay: cfg.Delay}
}

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
