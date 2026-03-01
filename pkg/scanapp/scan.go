package scanapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
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

type DialFunc func(network, address string, timeout time.Duration) (net.Conn, error)

type RunOptions struct {
	Dial             DialFunc
	ResumeStatePath  string
	PressureLimit    int
	DisableKeyboard  bool
	PressureHTTP     *http.Client
	ProgressInterval int
}

type chunkRuntime struct {
	meta  input.CIDRRecord
	net   *net.IPNet
	ports []int
	state *task.Chunk
	bkt   *ratelimit.LeakyBucket
}

type scanTask struct {
	chunkIdx int
	fabName  string
	cidr     string
	cidrName string
	ip       string
	port     int
}

type scanResult struct {
	chunkIdx int
	record   writer.Record
}

func Run(ctx context.Context, cfg config.Config, stdout, stderr io.Writer, opts RunOptions) error {
	logger := newLogger(cfg.LogLevel, cfg.Format == "json", stderr)

	if err := ensureFDLimit(cfg.Workers); err != nil {
		return err
	}

	cidrRecords, err := readCIDRFile(cfg.CIDRFile)
	if err != nil {
		return err
	}
	portSpecs, err := readPortFile(cfg.PortFile)
	if err != nil {
		return err
	}

	chunks, err := loadOrBuildChunks(cfg, cidrRecords, portSpecs)
	if err != nil {
		return err
	}
	runtimes, err := buildRuntime(chunks, cidrRecords, portSpecs, cfg)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cfg.Output), 0o755); err != nil && filepath.Dir(cfg.Output) != "." {
		return err
	}
	outFile, err := os.Create(cfg.Output)
	if err != nil {
		return err
	}
	defer outFile.Close()
	csvWriter := writer.NewCSVWriter(outFile)

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
		dial = net.DialTimeout
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
						IP:         res.IP,
						Port:       res.Port,
						Status:     res.Status,
						ResponseMS: res.ResponseTimeMS,
						FabName:    t.fabName,
						CIDR:       t.cidr,
						CIDRName:   t.cidrName,
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
		dispatchErrCh <- dispatchTasks(runCtx, cfg, ctrl, logger, runtimes, taskCh)
		close(taskCh)
	}()

	var (
		dispatchDone bool
		dispatchErr  error
		runErr       error
		written      int
	)
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
			ch := runtimes[res.chunkIdx].state
			ch.ScannedCount++
			if ch.ScannedCount >= ch.TotalCount {
				ch.Status = "completed"
			} else {
				ch.Status = "scanning"
			}
			written++
			if written%progressStep == 0 {
				_, _ = fmt.Fprintf(stdout, "progress cidr=%s scanned=%d/%d paused=%t\n", ch.CIDR, ch.ScannedCount, ch.TotalCount, ctrl.IsPaused())
			}
		}
	}

	for _, rt := range runtimes {
		if rt.bkt != nil {
			rt.bkt.Close()
		}
	}

	incomplete := hasIncomplete(runtimes)
	if incomplete || runErr != nil || shouldSaveOnDispatchErr(dispatchErr) {
		savePath := resumePath(cfg, opts)
		if err := state.Save(savePath, collectChunkStates(runtimes)); err != nil {
			return err
		}
		logger.infof("resume state saved to %s", savePath)
	}

	if runErr != nil {
		return runErr
	}
	if dispatchErr != nil {
		return dispatchErr
	}
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

			ip, port, err := task.IndexToIPv4Target(rt.net, rt.ports, i)
			if err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case taskCh <- scanTask{
				chunkIdx: idx,
				fabName:  rt.meta.FabName,
				cidr:     ch.CIDR,
				cidrName: ch.CIDRName,
				ip:       ip,
				port:     port,
			}:
			}
			ch.NextIndex = i + 1
			logger.debugf("dispatch cidr=%s target=%s:%d next_index=%d/%d", ch.CIDR, ip, port, ch.NextIndex, ch.TotalCount)
			if cfg.Delay > 0 {
				time.Sleep(cfg.Delay)
			}
		}
	}
	return nil
}

func readCIDRFile(path string) ([]input.CIDRRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return input.LoadCIDRs(f)
}

func readPortFile(path string) ([]input.PortSpec, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return input.LoadPorts(f)
}

func loadOrBuildChunks(cfg config.Config, cidrRecords []input.CIDRRecord, portSpecs []input.PortSpec) ([]task.Chunk, error) {
	if cfg.Resume != "" {
		return state.Load(cfg.Resume)
	}
	rawPorts := make([]string, 0, len(portSpecs))
	for _, p := range portSpecs {
		rawPorts = append(rawPorts, p.Raw)
	}
	out := make([]task.Chunk, 0, len(cidrRecords))
	for _, rec := range cidrRecords {
		hostCount, err := task.CountIPv4Hosts(rec.Net)
		if err != nil {
			return nil, err
		}
		total := hostCount * len(portSpecs)
		out = append(out, task.Chunk{
			CIDR:         rec.CIDR,
			CIDRName:     rec.CIDRName,
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
	cidrMeta := make(map[string]input.CIDRRecord, len(cidrRecords))
	for _, rec := range cidrRecords {
		cidrMeta[rec.CIDR] = rec
	}

	runtimes := make([]*chunkRuntime, 0, len(chunks))
	for i := range chunks {
		ch := &chunks[i]
		rec, ok := cidrMeta[ch.CIDR]
		if !ok {
			return nil, fmt.Errorf("cidr %s from chunk not found in cidr file", ch.CIDR)
		}

		portRows := ch.Ports
		if len(portRows) == 0 {
			portRows = make([]string, 0, len(defaultPorts))
			for _, p := range defaultPorts {
				portRows = append(portRows, p.Raw)
			}
			ch.Ports = append(ch.Ports, portRows...)
		}
		ports, err := parsePortRows(portRows)
		if err != nil {
			return nil, err
		}

		if ch.TotalCount == 0 {
			hostCount, err := task.CountIPv4Hosts(rec.Net)
			if err != nil {
				return nil, err
			}
			ch.TotalCount = hostCount * len(ports)
		}
		if ch.NextIndex >= ch.TotalCount {
			ch.Status = "completed"
		} else if ch.Status == "" {
			ch.Status = "pending"
		}
		rt := &chunkRuntime{
			meta:  rec,
			net:   rec.Net,
			ports: ports,
			state: ch,
			bkt:   ratelimit.NewLeakyBucket(cfg.BucketRate, cfg.BucketCapacity),
		}
		runtimes = append(runtimes, rt)
	}
	return runtimes, nil
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

func resumePath(cfg config.Config, opts RunOptions) string {
	if opts.ResumeStatePath != "" {
		return opts.ResumeStatePath
	}
	if cfg.Resume != "" {
		return cfg.Resume
	}
	return defaultResumeStateFile
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
	l.logf(0, "debug", format, args...)
}

func (l *scanLogger) infof(format string, args ...any) {
	l.logf(1, "info", format, args...)
}

func (l *scanLogger) errorf(format string, args ...any) {
	l.logf(2, "error", format, args...)
}

func (l *scanLogger) logf(level int, levelName, format string, args ...any) {
	if l == nil || level < l.level {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if l.asJSON {
		logx.LogJSON(l.out, levelName, msg, map[string]any{})
		return
	}
	_, _ = fmt.Fprintf(l.out, "[%s] %s\n", strings.ToUpper(levelName), msg)
}
