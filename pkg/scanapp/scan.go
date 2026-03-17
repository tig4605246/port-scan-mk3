package scanapp

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
	"github.com/xuxiping/port-scan-mk3/pkg/input"
	"github.com/xuxiping/port-scan-mk3/pkg/speedctrl"
	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

const (
	defaultResumeStateFile = "resume_state.json"
	defaultPressureLimit   = 90
)

// DialFunc abstracts TCP dialing for tests and runtime customization.
type DialFunc func(context.Context, string, string) (net.Conn, error)

// RunOptions customizes runtime behaviors that are not exposed as CLI flags.
type RunOptions struct {
	Dial             DialFunc
	ResumeStatePath  string
	PressureLimit    int
	DisableKeyboard  bool
	PressureHTTP     *http.Client
	ProgressInterval int
}

// Run executes a full scan flow: load inputs, dispatch scan tasks, write batch
// outputs, and persist resume state on interruption/failure.
func Run(ctx context.Context, cfg config.Config, stdout, stderr io.Writer, opts RunOptions) error {
	deps := defaultRunDependencies()
	logger := newLogger(cfg.LogLevel, cfg.Format == "json", stderr)
	if strings.TrimSpace(cfg.CIDRIPCol) == "" {
		cfg.CIDRIPCol = "ip"
	}
	if strings.TrimSpace(cfg.CIDRIPCidrCol) == "" {
		cfg.CIDRIPCidrCol = "ip_cidr"
	}

	if err := ensureFDLimit(cfg.Workers); err != nil {
		return err
	}

	inputs, err := loadRunInputs(cfg, deps)
	if err != nil {
		return err
	}
	plan, err := prepareRunPlan(cfg, inputs, deps, time.Now())
	if err != nil {
		return err
	}

	outputs, err := openBatchOutputs(plan.scanOutputPath, plan.openOnlyPath)
	if err != nil {
		return err
	}
	var scanSuccess bool
	defer func() {
		_ = outputs.Finalize(scanSuccess)
	}()

	workers := cfg.Workers
	if workers <= 0 {
		workers = 1
	}
	queueSize := workers * 2
	if queueSize < 1 {
		queueSize = 1
	}
	progressStep := opts.ProgressInterval
	if progressStep <= 0 {
		progressStep = 100
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctrl := speedctrl.NewController(speedctrl.WithAPIEnabled(!cfg.DisableAPI))
	if !opts.DisableKeyboard {
		if err := speedctrl.StartKeyboardLoop(runCtx, ctrl); err != nil {
			logger.errorf("failed to start keyboard loop: %v", err)
		}
	}
	startManualPauseMonitor(runCtx, ctrl, logger)

	apiErrCh := make(chan error, 1)
	if !cfg.DisableAPI {
		go pollPressureAPI(runCtx, cfg, opts, ctrl, logger, apiErrCh)
	}

	taskCh := make(chan scanTask, queueSize)

	dial := opts.Dial
	if dial == nil {
		dialer := &net.Dialer{LocalAddr: &net.TCPAddr{Port: 0}}
		dial = dialer.DialContext
	}
	resultCh := startScanExecutor(workers, cfg.Timeout, dial, logger, taskCh)

	dispatchErrCh := make(chan error, 1)
	go func() {
		dispatchErrCh <- dispatchTasks(runCtx, dispatchPolicyFromConfig(cfg), ctrl, logger, plan.runtimes, taskCh)
		close(taskCh)
	}()

	var (
		dispatchDone bool
		dispatchErr  error
		runErr       error
		summary      resultSummary
	)
	startedAt := time.Now()
	for !dispatchDone || resultCh != nil {
		select {
		case apiErr := <-apiErrCh:
			if apiErr != nil && runErr == nil {
				runErr = apiErr
				cancel()
			}
		case err := <-dispatchErrCh:
			dispatchDone = true
			dispatchErr = err
			dispatchErrCh = nil
		case res, ok := <-resultCh:
			if !ok {
				resultCh = nil
				continue
			}
			if runErr == nil {
				if err := writeScanRecord(outputs.scanWriter, outputs.openOnlyWriter, res.record); err != nil {
					runErr = err
					cancel()
				}
			}
			applyScanResult(plan.runtimes, res, &summary)
			if runErr == nil {
				emitScanResultEvents(stdout, logger, ctrl, progressStep, plan.runtimes, res, &summary)
			}
		}
	}

	for _, rt := range plan.runtimes {
		if rt.bkt != nil {
			rt.bkt.Close()
		}
	}

	if err := persistResumeState(cfg, opts, logger, plan.runtimes, dispatchErr, runErr); err != nil {
		return err
	}

	if runErr != nil {
		emitCompletionSummary(logger, summary, startedAt, runErr)
		return runErr
	}
	if dispatchErr != nil {
		emitCompletionSummary(logger, summary, startedAt, dispatchErr)
		return dispatchErr
	}
	emitCompletionSummary(logger, summary, startedAt, nil)
	scanSuccess = true
	return nil
}

type cidrGroup struct {
	targets []scanTarget
	port    int
}

func buildCIDRGroups(cidrRecords []input.CIDRRecord) (map[string]cidrGroup, error) {
	out := make(map[string]cidrGroup)
	for _, rec := range cidrRecords {
		cidr := rec.CIDR
		if cidr == "" && rec.Net != nil {
			cidr = rec.Net.String()
		}
		if cidr == "" {
			return nil, fmt.Errorf("record missing ip_cidr")
		}

		selector := ""
		switch {
		case rec.Selector != nil:
			selector = rec.Selector.String()
		case strings.TrimSpace(rec.IPRaw) != "":
			selector = strings.TrimSpace(rec.IPRaw)
		case rec.Net != nil:
			selector = rec.Net.String()
		default:
			return nil, fmt.Errorf("record for cidr %s missing selector", cidr)
		}

		ips, err := task.ExpandIPSelectors([]string{selector})
		if err != nil {
			return nil, fmt.Errorf("expand selector failed for cidr %s: %w", cidr, err)
		}

		group := out[cidr]
		for _, ip := range ips {
			group.targets = append(group.targets, scanTarget{
				ip: ip,
				meta: targetMeta{
					fabName:  rec.FabName,
					cidrName: rec.CIDRName,
				},
			})
		}
		out[cidr] = group
	}

	for cidr, group := range out {
		sort.Slice(group.targets, func(i, j int) bool {
			return ipv4ToUint32(group.targets[i].ip) < ipv4ToUint32(group.targets[j].ip)
		})
		out[cidr] = group
	}
	return out, nil
}

func buildRichGroups(cidrRecords []input.CIDRRecord) (map[string]cidrGroup, error) {
	out := make(map[string]cidrGroup)
	for _, rec := range cidrRecords {
		if !rec.IsRich || !rec.IsValid {
			continue
		}
		key := strings.TrimSpace(rec.ExecutionKey)
		if key == "" {
			return nil, fmt.Errorf("rich record missing execution_key at row %d", rec.RowNumber)
		}
		group := out[key]
		if len(group.targets) == 0 {
			group.port = rec.Port
			group.targets = append(group.targets, scanTarget{
				ip:     rec.DstIP,
				ipCidr: rec.DstNetworkSegment,
				meta: targetMeta{
					fabName:           rec.FabName,
					cidrName:          rec.CIDRName,
					serviceLabel:      rec.ServiceLabel,
					decision:          rec.Decision,
					policyID:          rec.PolicyID,
					reason:            rec.Reason,
					executionKey:      key,
					srcIP:             rec.SrcIP,
					srcNetworkSegment: rec.SrcNetworkSegment,
				},
			})
			out[key] = group
			continue
		}
		if group.port != rec.Port {
			return nil, fmt.Errorf("execution key %s has inconsistent port", key)
		}
		t := &group.targets[0]
		t.meta.fabName = mergeFieldValue(t.meta.fabName, rec.FabName)
		t.meta.cidrName = mergeFieldValue(t.meta.cidrName, rec.CIDRName)
		t.meta.serviceLabel = mergeFieldValue(t.meta.serviceLabel, rec.ServiceLabel)
		t.meta.decision = mergeFieldValue(t.meta.decision, rec.Decision)
		t.meta.policyID = mergeFieldValue(t.meta.policyID, rec.PolicyID)
		t.meta.reason = mergeFieldValue(t.meta.reason, rec.Reason)
		t.meta.srcIP = mergeFieldValue(t.meta.srcIP, rec.SrcIP)
		t.meta.srcNetworkSegment = mergeFieldValue(t.meta.srcNetworkSegment, rec.SrcNetworkSegment)
		out[key] = group
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no usable input rows")
	}
	return out, nil
}

func mergeFieldValue(existing, incoming string) string {
	existing = strings.TrimSpace(existing)
	incoming = strings.TrimSpace(incoming)
	if incoming == "" || existing == incoming {
		return existing
	}
	if existing == "" {
		return incoming
	}
	parts := strings.Split(existing, "|")
	for _, p := range parts {
		if p == incoming {
			return existing
		}
	}
	return existing + "|" + incoming
}

