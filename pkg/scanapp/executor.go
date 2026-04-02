package scanapp

import (
	"sync"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/scanner"
)

// startScanExecutor launches worker goroutines that consume scanTask items from taskCh,
// execute TCP scans, and emit structured log events for each result.
//
// Parameters:
//   - workers: number of concurrent scan workers; must be > 0
//   - timeout: per-scan connection timeout
//   - dial: TCP dial function (allows injection for testing)
//   - logger: structured logger for scan events
//   - taskCh: channel delivering scan tasks
//
// Returns a channel that will be closed when all workers finish scanning.
func startScanExecutor(workers int, timeout time.Duration, dial DialFunc, logger *scanLogger, taskCh <-chan scanTask) <-chan scanResult {
	if workers <= 0 {
		workers = 1
	}

	resultCh := make(chan scanResult, workers*2)

	var workerWG sync.WaitGroup
	for i := 0; i < workers; i++ {
		workerWG.Add(1)
		go func() {
			defer workerWG.Done()
			defer func() {
				if r := recover(); r != nil {
					logger.errorf("executor worker panic: %v", r)
				}
			}()
			for t := range taskCh {
				res := scanner.ScanTCP(dial, t.ip, t.port, timeout)
				state := LogEventScanned
				errCause := LogEventNone
				if res.Error != "" {
					errCause = LogEventRuntimeErr
					state = LogEventError
				}
				logger.eventf(LogEventScanResult, t.ip, t.port, state, errCause, map[string]any{
					"status": res.Status,
					"error":  res.Error,
				})
				resultCh <- scanResult{
					chunkIdx: t.chunkIdx,
					record:   recordFromScanTask(t, res),
				}
			}
		}()
	}

	go func() {
		workerWG.Wait()
		close(resultCh)
	}()

	return resultCh
}
