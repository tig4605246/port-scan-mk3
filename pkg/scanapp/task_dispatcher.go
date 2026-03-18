package scanapp

import (
	"context"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
	"github.com/xuxiping/port-scan-mk3/pkg/speedctrl"
)

type dispatchPolicy struct {
	delay    time.Duration
	observer dispatchObserver
}

func dispatchPolicyFromConfig(cfg config.Config) dispatchPolicy {
	return dispatchPolicy{
		delay:    cfg.Delay,
		observer: noopDispatchObserver{},
	}
}

func dispatchTasks(ctx context.Context, policy dispatchPolicy, ctrl *speedctrl.Controller, logger *scanLogger, runtimes []*chunkRuntime, taskCh chan<- scanTask) error {
	obs := policy.observer
	if obs == nil {
		obs = noopDispatchObserver{}
	}
	for idx := range runtimes {
		rt := runtimes[idx]
		ch := rt.state
		snap := rt.tracker.Snapshot()
		if snap.NextIndex >= snap.TotalCount {
			rt.tracker.AdvanceNextIndex(snap.NextIndex) // triggers status update
			continue
		}
		rt.tracker.AdvanceNextIndex(snap.NextIndex) // sets status to "scanning"
		for i := snap.NextIndex; i < snap.TotalCount; i++ {
			obs.OnBucketWaitStart(ch.CIDR, i)
			if err := rt.bkt.Acquire(ctx); err != nil {
				return err
			}
			obs.OnBucketAcquired(ch.CIDR, i)

			obs.OnGateWaitStart(ch.CIDR, i)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ctrl.Gate():
			}
			obs.OnGateReleased(ch.CIDR, i)

			target, port, err := indexToRuntimeTarget(rt.targets, rt.ports, i)
			if err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case taskCh <- scanTask{
				chunkIdx: idx,
				ipCidr:   defaultString(target.ipCidr, ch.CIDR),
				ip:       target.ip,
				port:     port,
				meta:     target.meta,
			}:
			}
			obs.OnTaskEnqueued(ch.CIDR, i)
			rt.tracker.AdvanceNextIndex(i + 1)
			logger.debugf("dispatch cidr=%s target=%s:%d next_index=%d/%d", ch.CIDR, target.ip, port, i+1, snap.TotalCount)
			if policy.delay > 0 {
				time.Sleep(policy.delay)
			}
		}
	}
	return nil
}
