package scanapp

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
	"github.com/xuxiping/port-scan-mk3/pkg/input"
	"github.com/xuxiping/port-scan-mk3/pkg/logx"
	"github.com/xuxiping/port-scan-mk3/pkg/ratelimit"
	"github.com/xuxiping/port-scan-mk3/pkg/scanner"
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

	outFile, err := os.Create(plan.scanOutputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	csvWriter := writer.NewCSVWriter(outFile)
	if err := csvWriter.WriteHeader(); err != nil {
		return err
	}

	openOnlyFile, err := os.Create(plan.openOnlyPath)
	if err != nil {
		return err
	}
	defer openOnlyFile.Close()
	openOnlyWriter := writer.NewOpenOnlyWriter(writer.NewCSVWriter(openOnlyFile))
	if err := openOnlyWriter.WriteHeader(); err != nil {
		return err
	}

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
	resultCh := make(chan scanResult, queueSize)

	dial := opts.Dial
	if dial == nil {
		dialer := &net.Dialer{LocalAddr: &net.TCPAddr{Port: 0}}
		dial = dialer.DialContext
	}

	var workerWG sync.WaitGroup
	for i := 0; i < workers; i++ {
		workerWG.Add(1)
		go func() {
			defer workerWG.Done()
			for t := range taskCh {
				res := scanner.ScanTCP(dial, t.ip, t.port, cfg.Timeout)
				if res.Error != "" {
					logger.debugf("scan %s:%d status=%s err=%s", t.ip, t.port, res.Status, res.Error)
				}
				resultCh <- scanResult{
					chunkIdx: t.chunkIdx,
					record: writer.Record{
						IP:                res.IP,
						IPCidr:            t.ipCidr,
						Port:              res.Port,
						Status:            res.Status,
						ResponseMS:        res.ResponseTimeMS,
						FabName:           t.fabName,
						CIDRName:          t.cidrName,
						ServiceLabel:      t.serviceLabel,
						Decision:          t.decision,
						PolicyID:          t.policyID,
						Reason:            t.reason,
						ExecutionKey:      t.executionKey,
						SrcIP:             t.srcIP,
						SrcNetworkSegment: t.srcNetworkSegment,
					},
				}
			}
		}()
	}

	go func() {
		workerWG.Wait()
		close(resultCh)
	}()

	dispatchErrCh := make(chan error, 1)
	go func() {
		dispatchErrCh <- dispatchTasks(runCtx, cfg, ctrl, logger, plan.runtimes, taskCh)
		close(taskCh)
	}()

	var (
		dispatchDone bool
		dispatchErr  error
		runErr       error
		written      int
		openCount    int
		closeCount   int
		timeoutCount int
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
			if err := csvWriter.Write(res.record); err != nil && runErr == nil {
				runErr = err
				cancel()
			}
			if err := openOnlyWriter.Write(res.record); err != nil && runErr == nil {
				runErr = err
				cancel()
			}
			ch := plan.runtimes[res.chunkIdx].state
			ch.ScannedCount++
			if ch.ScannedCount >= ch.TotalCount {
				ch.Status = "completed"
			} else {
				ch.Status = "scanning"
			}
			written++
			switch {
			case strings.EqualFold(res.record.Status, "open"):
				openCount++
			case strings.Contains(strings.ToLower(res.record.Status), "timeout"):
				timeoutCount++
			default:
				closeCount++
			}
			logger.eventf("scan_result", res.record.IP, res.record.Port, "scanned", statusErrorCause(res.record.Status), map[string]any{
				"status":           res.record.Status,
				"response_time_ms": res.record.ResponseMS,
				"cidr":             res.record.IPCidr,
			})
			if written%progressStep == 0 {
				_, _ = fmt.Fprintf(stdout, "progress cidr=%s scanned=%d/%d paused=%t\n", ch.CIDR, ch.ScannedCount, ch.TotalCount, ctrl.IsPaused())
				completionRate := 0.0
				if ch.TotalCount > 0 {
					completionRate = float64(ch.ScannedCount) / float64(ch.TotalCount)
				}
				logger.eventf("scan_progress", "", 0, "progress", "none", map[string]any{
					"cidr":            ch.CIDR,
					"scanned_count":   ch.ScannedCount,
					"total_count":     ch.TotalCount,
					"completion_rate": completionRate,
					"paused":          ctrl.IsPaused(),
				})
			}
		}
	}

	for _, rt := range plan.runtimes {
		if rt.bkt != nil {
			rt.bkt.Close()
		}
	}

	incomplete := hasIncomplete(plan.runtimes)
	if incomplete || runErr != nil || shouldSaveOnDispatchErr(dispatchErr) {
		savePath := resumePath(cfg, opts)
		if err := state.Save(savePath, collectChunkStates(plan.runtimes)); err != nil {
			return err
		}
		logger.infof("resume state saved to %s", savePath)
	}

	if runErr != nil {
		logger.eventf("scan_completion", "", 0, "completion_summary", errorCause(runErr), map[string]any{
			"total_tasks":   written,
			"open_count":    openCount,
			"close_count":   closeCount,
			"timeout_count": timeoutCount,
			"duration_ms":   time.Since(startedAt).Milliseconds(),
			"success":       false,
		})
		return runErr
	}
	if dispatchErr != nil {
		logger.eventf("scan_completion", "", 0, "completion_summary", errorCause(dispatchErr), map[string]any{
			"total_tasks":   written,
			"open_count":    openCount,
			"close_count":   closeCount,
			"timeout_count": timeoutCount,
			"duration_ms":   time.Since(startedAt).Milliseconds(),
			"success":       false,
		})
		return dispatchErr
	}
	logger.eventf("scan_completion", "", 0, "completion_summary", "none", map[string]any{
		"total_tasks":   written,
		"open_count":    openCount,
		"close_count":   closeCount,
		"timeout_count": timeoutCount,
		"duration_ms":   time.Since(startedAt).Milliseconds(),
		"success":       true,
	})
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

func dispatchTasks(ctx context.Context, cfg config.Config, ctrl *speedctrl.Controller, logger *scanLogger, runtimes []*chunkRuntime, taskCh chan<- scanTask) error {
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
				chunkIdx:          idx,
				fabName:           target.fabName,
				ipCidr:            defaultString(target.ipCidr, ch.CIDR),
				cidrName:          target.cidrName,
				ip:                target.ip,
				port:              port,
				serviceLabel:      target.serviceLabel,
				decision:          target.decision,
				policyID:          target.policyID,
				reason:            target.reason,
				executionKey:      target.executionKey,
				srcIP:             target.srcIP,
				srcNetworkSegment: target.srcNetworkSegment,
			}:
			}
			ch.NextIndex = i + 1
			logger.debugf("dispatch cidr=%s target=%s:%d next_index=%d/%d", ch.CIDR, target.ip, port, ch.NextIndex, ch.TotalCount)
			if cfg.Delay > 0 {
				time.Sleep(cfg.Delay)
			}
		}
	}
	return nil
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

func startManualPauseMonitor(ctx context.Context, ctrl *speedctrl.Controller, logger *scanLogger) {
	go func() {
		prev := ctrl.ManualPaused()
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				curr := ctrl.ManualPaused()
				if curr != prev {
					if curr {
						logger.infof("[Manual] 接收到按鍵指令，掃描已手動暫停")
					} else {
						logger.infof("[Manual] 掃描已手動恢復")
					}
					prev = curr
				}
			}
		}
	}()
}

func pollPressureAPI(ctx context.Context, cfg config.Config, opts RunOptions, ctrl *speedctrl.Controller, logger *scanLogger, errCh chan<- error) {
	interval := cfg.PressureInterval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	threshold := opts.PressureLimit
	if threshold <= 0 {
		threshold = defaultPressureLimit
	}
	client := opts.PressureHTTP
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Second}
	}

	var consecutiveFailures int
	var prevPaused bool
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pressure, err := fetchPressure(client, cfg.PressureAPI)
			if err != nil {
				consecutiveFailures++
				if consecutiveFailures <= 2 {
					logger.errorf("pressure api request failed (%d/3): %v", consecutiveFailures, err)
					continue
				}
				select {
				case errCh <- fmt.Errorf("pressure api failed 3 times: %w", err):
				default:
				}
				return
			}
			consecutiveFailures = 0

			paused := pressure >= threshold
			ctrl.SetAPIPaused(paused)
			if paused != prevPaused {
				if paused {
					logger.infof("[API] 路由器壓力過載，掃描已自動暫停 pressure=%d threshold=%d", pressure, threshold)
				} else {
					logger.infof("[API] 路由器壓力恢復，掃描已自動恢復 pressure=%d threshold=%d", pressure, threshold)
				}
				prevPaused = paused
			}
		}
	}
}

func fetchPressure(client *http.Client, url string) (int, error) {
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("pressure api status=%d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, err
	}
	raw, ok := body["pressure"]
	if !ok {
		return 0, fmt.Errorf("pressure field missing")
	}
	switch v := raw.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case string:
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		return n, nil
	default:
		return 0, fmt.Errorf("unsupported pressure field type: %T", raw)
	}
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
