package scanapp

import (
	"context"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/speedctrl"
)

func dispatchTasks(ctx context.Context, policy dispatchPolicy, ctrl *speedctrl.Controller, logger *scanLogger, runtimes []*chunkRuntime, taskCh chan<- scanTask) error {
	for idx := range runtimes {
		rt := runtimes[idx]
		ch := rt.state
		if ch.NextIndex >= ch.TotalCount {
			ch.Status = "completed"
			continue
		}
		ch.Status = "scanning"
		for i := ch.NextIndex; i < ch.TotalCount; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ctrl.Gate():
			}

			if err := rt.bkt.Acquire(ctx); err != nil {
				return err
			}

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
			ch.NextIndex = i + 1
			logger.debugf("dispatch cidr=%s target=%s:%d next_index=%d/%d", ch.CIDR, target.ip, port, ch.NextIndex, ch.TotalCount)
			if policy.delay > 0 {
				time.Sleep(policy.delay)
			}
		}
	}
	return nil
}
