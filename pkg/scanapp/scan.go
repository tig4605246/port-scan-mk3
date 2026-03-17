package scanapp

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
	"github.com/xuxiping/port-scan-mk3/pkg/input"
	"github.com/xuxiping/port-scan-mk3/pkg/logx"
	"github.com/xuxiping/port-scan-mk3/pkg/ratelimit"
	"github.com/xuxiping/port-scan-mk3/pkg/speedctrl"
	"github.com/xuxiping/port-scan-mk3/pkg/state"
	"github.com/xuxiping/port-scan-mk3/pkg/task"
	"github.com/xuxiping/port-scan-mk3/pkg/writer"
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

type chunkRuntime struct {
	ipCidr  string
	ports   []int
	targets []scanTarget
	state   *task.Chunk
	bkt     *ratelimit.LeakyBucket
}

type scanTarget struct {
	ip                string
	ipCidr            string
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

type scanTask struct {
	chunkIdx          int
	fabName           string
	ipCidr            string
	cidrName          string
	ip                string
	port              int
	serviceLabel      string
	decision          string
	policyID          string
	reason            string
	executionKey      string
	srcIP             string
	srcNetworkSegment string
}

type scanResult struct {
	chunkIdx int
	record   writer.Record
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
	defer outputs.Close()

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
		dispatchErrCh <- dispatchTasks(runCtx, cfg, ctrl, logger, plan.runtimes, taskCh)
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
			if err := writeScanRecord(outputs.scanWriter, outputs.openOnlyWriter, res.record); err != nil && runErr == nil {
				runErr = err
				cancel()
			}
			applyScanResult(plan.runtimes, res, &summary)
			emitScanResultEvents(stdout, logger, ctrl, progressStep, plan.runtimes, res, &summary)
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
	return nil
}

func shouldSaveOnDispatchErr(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func hasIncomplete(runtimes []*chunkRuntime) bool {
	for _, rt := range runtimes {
		if rt.state.ScannedCount < rt.state.TotalCount {
			return true
		}
	}
	return false
}

func collectChunkStates(runtimes []*chunkRuntime) []task.Chunk {
	out := make([]task.Chunk, 0, len(runtimes))
	for _, rt := range runtimes {
		out = append(out, *rt.state)
	}
	return out
}

func loadOrBuildChunks(cfg config.Config, cidrRecords []input.CIDRRecord, portSpecs []input.PortSpec) ([]task.Chunk, error) {
	if cfg.Resume != "" {
		return state.Load(cfg.Resume)
	}
	if hasRichRecords(cidrRecords) {
		return buildRichChunks(cidrRecords)
	}
	groups, err := buildCIDRGroups(cidrRecords)
	if err != nil {
		return nil, err
	}
	rawPorts := make([]string, 0, len(portSpecs))
	for _, p := range portSpecs {
		rawPorts = append(rawPorts, p.Raw)
	}
	cidrs := make([]string, 0, len(groups))
	for cidr := range groups {
		cidrs = append(cidrs, cidr)
	}
	sort.Strings(cidrs)

	out := make([]task.Chunk, 0, len(cidrs))
	for _, cidr := range cidrs {
		g := groups[cidr]
		total := len(g.targets) * len(portSpecs)
		cidrName := ""
		if len(g.targets) > 0 {
			cidrName = g.targets[0].cidrName
		}
		out = append(out, task.Chunk{
			CIDR:         cidr,
			CIDRName:     cidrName,
			Ports:        rawPorts,
			NextIndex:    0,
			ScannedCount: 0,
			TotalCount:   total,
			Status:       "pending",
		})
	}
	return out, nil
}

func buildRuntime(chunks []task.Chunk, cidrRecords []input.CIDRRecord, defaultPorts []input.PortSpec, cfg config.Config) ([]*chunkRuntime, error) {
	var (
		groups map[string]cidrGroup
		err    error
	)
	if hasRichRecords(cidrRecords) {
		groups, err = buildRichGroups(cidrRecords)
	} else {
		groups, err = buildCIDRGroups(cidrRecords)
	}
	if err != nil {
		return nil, err
	}

	runtimes := make([]*chunkRuntime, 0, len(chunks))
	for i := range chunks {
		ch := &chunks[i]
		group, ok := groups[ch.CIDR]
		if !ok {
			return nil, fmt.Errorf("cidr %s from chunk not found in cidr file", ch.CIDR)
		}

		portRows := ch.Ports
		if len(portRows) == 0 {
			if group.port > 0 {
				portRows = []string{fmt.Sprintf("%d/tcp", group.port)}
			} else {
				portRows = make([]string, 0, len(defaultPorts))
				for _, p := range defaultPorts {
					portRows = append(portRows, p.Raw)
				}
			}
			ch.Ports = append(ch.Ports, portRows...)
		}
		ports, err := parsePortRows(portRows)
		if err != nil {
			return nil, err
		}

		expectedTotal := len(group.targets) * len(ports)
		if ch.TotalCount == 0 {
			ch.TotalCount = expectedTotal
		}
		if ch.TotalCount != expectedTotal {
			return nil, fmt.Errorf("chunk total_count mismatch for %s: state=%d expected=%d", ch.CIDR, ch.TotalCount, expectedTotal)
		}
		if ch.NextIndex >= ch.TotalCount {
			ch.Status = "completed"
		} else if ch.Status == "" {
			ch.Status = "pending"
		}
		rt := &chunkRuntime{
			ipCidr:  ch.CIDR,
			ports:   ports,
			targets: group.targets,
			state:   ch,
			bkt:     ratelimit.NewLeakyBucket(cfg.BucketRate, cfg.BucketCapacity),
		}
		runtimes = append(runtimes, rt)
	}
	return runtimes, nil
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
				ip:       ip,
				fabName:  rec.FabName,
				cidrName: rec.CIDRName,
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

func hasRichRecords(cidrRecords []input.CIDRRecord) bool {
	for _, rec := range cidrRecords {
		if rec.IsRich {
			return true
		}
	}
	return false
}

func buildRichChunks(cidrRecords []input.CIDRRecord) ([]task.Chunk, error) {
	groups, err := buildRichGroups(cidrRecords)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]task.Chunk, 0, len(keys))
	for _, key := range keys {
		g := groups[key]
		cidrName := ""
		if len(g.targets) > 0 {
			cidrName = g.targets[0].cidrName
		}
		out = append(out, task.Chunk{
			CIDR:         key,
			CIDRName:     cidrName,
			Ports:        []string{fmt.Sprintf("%d/tcp", g.port)},
			NextIndex:    0,
			ScannedCount: 0,
			TotalCount:   1,
			Status:       "pending",
		})
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
				ip:                rec.DstIP,
				ipCidr:            rec.DstNetworkSegment,
				fabName:           rec.FabName,
				cidrName:          rec.CIDRName,
				serviceLabel:      rec.ServiceLabel,
				decision:          rec.Decision,
				policyID:          rec.PolicyID,
				reason:            rec.Reason,
				executionKey:      key,
				srcIP:             rec.SrcIP,
				srcNetworkSegment: rec.SrcNetworkSegment,
			})
			out[key] = group
			continue
		}
		if group.port != rec.Port {
			return nil, fmt.Errorf("execution key %s has inconsistent port", key)
		}
		t := &group.targets[0]
		t.fabName = mergeFieldValue(t.fabName, rec.FabName)
		t.cidrName = mergeFieldValue(t.cidrName, rec.CIDRName)
		t.serviceLabel = mergeFieldValue(t.serviceLabel, rec.ServiceLabel)
		t.decision = mergeFieldValue(t.decision, rec.Decision)
		t.policyID = mergeFieldValue(t.policyID, rec.PolicyID)
		t.reason = mergeFieldValue(t.reason, rec.Reason)
		t.srcIP = mergeFieldValue(t.srcIP, rec.SrcIP)
		t.srcNetworkSegment = mergeFieldValue(t.srcNetworkSegment, rec.SrcNetworkSegment)
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

func indexToRuntimeTarget(targets []scanTarget, ports []int, idx int) (scanTarget, int, error) {
	if len(targets) == 0 {
		return scanTarget{}, 0, fmt.Errorf("empty targets")
	}
	if len(ports) == 0 {
		return scanTarget{}, 0, fmt.Errorf("empty ports")
	}
	if idx < 0 {
		return scanTarget{}, 0, fmt.Errorf("negative index")
	}
	targetIdx := idx / len(ports)
	portIdx := idx % len(ports)
	if targetIdx >= len(targets) {
		return scanTarget{}, 0, fmt.Errorf("index out of range")
	}
	return targets[targetIdx], ports[portIdx], nil
}

func ipv4ToUint32(ip string) uint32 {
	parsed := net.ParseIP(ip).To4()
	if parsed == nil {
		return 0
	}
	return binary.BigEndian.Uint32(parsed)
}

func parsePortRows(rows []string) ([]int, error) {
	ports := make([]int, 0, len(rows))
	for _, row := range rows {
		parts := strings.Split(strings.TrimSpace(row), "/")
		if len(parts) != 2 || strings.ToLower(parts[1]) != "tcp" {
			return nil, fmt.Errorf("invalid chunk port row: %s", row)
		}
		n, err := strconv.Atoi(parts[0])
		if err != nil || n < 1 || n > 65535 {
			return nil, fmt.Errorf("invalid chunk port number: %s", row)
		}
		ports = append(ports, n)
	}
	return ports, nil
}

func defaultString(primary, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}

func ensureFDLimit(workers int) error {
	var lim syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim); err != nil {
		return nil
	}
	minNeed := uint64(1024)
	if workers > 0 {
		workerNeed := uint64(workers * 8)
		if workerNeed > minNeed {
			minNeed = workerNeed
		}
	}
	if lim.Cur < minNeed {
		return fmt.Errorf("file descriptor limit too low: %d (need >= %d)", lim.Cur, minNeed)
	}
	return nil
}

type scanLogger struct {
	level  int
	asJSON bool
	out    io.Writer
}

func newLogger(level string, asJSON bool, out io.Writer) *scanLogger {
	parsed := 1
	switch strings.ToLower(level) {
	case "debug":
		parsed = 0
	case "info":
		parsed = 1
	case "error":
		parsed = 2
	}
	return &scanLogger{level: parsed, asJSON: asJSON, out: out}
}

func (l *scanLogger) debugf(format string, args ...any) {
	l.logWithFields(0, "debug", fmt.Sprintf(format, args...), nil)
}

func (l *scanLogger) infof(format string, args ...any) {
	l.logWithFields(1, "info", fmt.Sprintf(format, args...), nil)
}

func (l *scanLogger) errorf(format string, args ...any) {
	l.logWithFields(2, "error", fmt.Sprintf(format, args...), nil)
}

func (l *scanLogger) eventf(msg, target string, port int, transition, errCause string, extra map[string]any) {
	fields := map[string]any{
		"target":           target,
		"port":             port,
		"state_transition": transition,
		"error_cause":      errCause,
	}
	for k, v := range extra {
		fields[k] = v
	}
	l.logWithFields(1, "info", msg, fields)
}

func (l *scanLogger) logWithFields(level int, levelName, msg string, fields map[string]any) {
	if l == nil || level < l.level {
		return
	}
	if fields == nil {
		fields = map[string]any{}
	}
	if l.asJSON {
		logx.LogJSON(l.out, levelName, msg, fields)
		return
	}
	if len(fields) > 0 {
		_, _ = fmt.Fprintf(l.out, "[%s] %s fields=%v\n", strings.ToUpper(levelName), msg, fields)
		return
	}
	_, _ = fmt.Fprintf(l.out, "[%s] %s\n", strings.ToUpper(levelName), msg)
}

func statusErrorCause(status string) string {
	s := strings.ToLower(status)
	switch {
	case strings.Contains(s, "timeout"):
		return "timeout"
	case s == "close":
		return "closed"
	default:
		return "none"
	}
}

func errorCause(err error) string {
	if err == nil {
		return "none"
	}
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "deadline_exceeded"
	}
	return "runtime_error"
}
