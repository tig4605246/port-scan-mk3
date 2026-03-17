package scanapp

import (
	"sync"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/scanner"
)

func startScanExecutor(workers int, timeout time.Duration, dial DialFunc, logger *scanLogger, taskCh <-chan scanTask) <-chan scanResult {
	if workers <= 0 {
		workers = 1
	}

	resultCh := make(chan scanResult, workers*2)
	if cap(resultCh) < 1 {
		resultCh = make(chan scanResult, 1)
	}

	var workerWG sync.WaitGroup
	for i := 0; i < workers; i++ {
		workerWG.Add(1)
		go func() {
			defer workerWG.Done()
			for t := range taskCh {
				res := scanner.ScanTCP(dial, t.ip, t.port, timeout)
				if res.Error != "" {
					logger.debugf("scan %s:%d status=%s err=%s", t.ip, t.port, res.Status, res.Error)
				}
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
