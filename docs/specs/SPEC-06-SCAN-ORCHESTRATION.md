# SPEC-06: Scan Orchestration Specification

## Overview

```
pkg/scanapp/
├── scan.go                    # Main orchestration (Run function)
├── executor.go                 # Worker pool
├── task_dispatcher.go         # Task dispatch with rate limiting
├── resume_manager.go          # Resume state persistence
├── input_loader.go            # Input loading
├── runtime_types.go           # Core data structures
├── runtime_builder.go         # Run plan composition
├── group_builder.go           # CIDR grouping (basic/rich strategies)
├── chunk_lifecycle.go         # Chunk management
├── result_aggregator.go       # Result processing
├── batch_output.go            # Batch output paths
├── output_files.go            # Output file management
├── scan_helpers.go            # Helper functions
├── scan_logger.go            # Logging
├── dispatch_observer.go       # Dispatch events
├── pressure_monitor.go        # Pressure API polling
├── pressure.go                # PressureFetcher interface
└── runtime_record_mapper.go   # Record mapping
```

## 1. Main Entry Point

### Run Function

```go
func Run(
    ctx context.Context,
    cfg config.Config,
    stdout io.Writer,
    stderr io.Writer,
    opts RunOptions,
) error
```

### RunOptions

```go
type RunOptions struct {
    Dial            scanner.DialFunc      // Custom dial function (optional)
    PressureFetcher PressureFetcher       // Custom pressure fetcher (optional)
    Logger          *logx.Logger          // Custom logger (optional)
}
```

## 2. Pipeline Stages

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           RUN() ENTRY                                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ 1. LOAD (input_loader.go)                                               │
│    loadRunInputs() ──────────────────────────────────────────────────►  │
│    - Read CIDR CSV file                                                 │
│    - Read port file (if required)                                       │
│    Returns: runInputs{cidrRecords, portSpecs}                           │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ 2. PLAN (runtime_builder.go + chunk_lifecycle.go + group_builder.go)   │
│    prepareRunPlan() ─────────────────────────────────────────────────►  │
│    ├─ loadOrBuildChunks() → chunks[]                                   │
│    │   - If resume: load from JSON                                       │
│    │   - Else: build from input records                                 │
│    ├─ buildRuntime() → runtimes[]                                       │
│    │   - Create chunkRuntime per CIDR                                   │
│    │   - Initialize state tracker                                        │
│    │   - Initialize rate limiter bucket                                  │
│    └─ resolveBatchOutputPaths() → output paths                          │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ 3. OUTPUT SETUP (batch_output.go + output_files.go)                    │
│    openBatchOutputs() ──────────────────────────────────────────────►  │
│    - Create .tmp files with headers                                     │
│    - Returns: outputFiles{scanWriter, openWriter}                       │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ 4. EXECUTE (executor.go)                                                │
│    startScanExecutor() ──────────────────────────────────────────────►  │
│    - Create worker pool (config.Workers count)                          │
│    - Each worker: receive task → ScanTCP → send result                  │
│    - Returns: resultCh (channel of scanResult)                          │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ 5. DISPATCH (task_dispatcher.go)                                        │
│    dispatchTasks() ─────────────────────────────────────────────────►  │
│    - Iterate over runtimes (chunk-serial)                              │
│    - For each index:                                                     │
│      ├─ Acquire rate limit token                                         │
│      ├─ Wait on pause gate                                               │
│      ├─ Create scanTask                                                  │
│      ├─ Send to taskCh                                                   │
│      ├─ Update tracker                                                  │
│      └─ Apply delay (if configured)                                     │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ 6. AGGREGATE (result_aggregator.go + main loop)                         │
│    Main event loop ─────────────────────────────────────────────────►  │
│    For each result from resultCh:                                        │
│      ├─ writeScanRecord() → output files                                 │
│      ├─ applyScanResult() → update state                                │
│      └─ emitScanResultEvents() → stdout/log                             │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ 7. RESUME (resume_manager.go)                                           │
│    persistResumeState() ─────────────────────────────────────────────►  │
│    Called on:                                                            │
│      - Scan canceled (SIGINT)                                           │
│      - Runtime error                                                     │
│      - Dispatch error (if canceled/deadline)                            │
│    Saves: chunk states (NextIndex, ScannedCount, Status)                │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│ 8. FINALIZE (output_files.go)                                           │
│    outputs.Finalize(success) ────────────────────────────────────────►   │
│    - If success: rename .tmp → final names                               │
│    - If failure: keep .tmp files                                        │
└─────────────────────────────────────────────────────────────────────────┘
```

## 3. Input Loader (input_loader.go)

### Functions

```go
func loadRunInputs(cfg config.Config) (runInputs, error)

type runInputs struct {
    cidrRecords []input.CIDRRecord
    portSpecs   []input.PortSpec
}
```

### Flow

1. `readCIDRFile(cfg.CIDRFile)` → `input.LoadCIDRsWithColumns()`
2. `readPortFile(cfg.PortFile)` → `input.LoadPorts()` (if port file provided)
3. Return combined `runInputs`

## 4. Chunk Lifecycle (chunk_lifecycle.go)

### loadOrBuildChunks

```go
func loadOrBuildChunks(
    inputs runInputs,
    resumePath string,
) ([]task.Chunk, error)
```

**If resumePath provided:**
- Load chunks from JSON file
- Return loaded chunks

**Otherwise:**
- Call `buildChunks(inputs)` to create new chunks

### buildRuntime

```go
func buildRuntime(
    chunks []task.Chunk,
    policy dispatchPolicy,
) ([]chunkRuntime, error)
```

Creates `chunkRuntime` for each chunk:
- `targets`: scanTarget array
- `ports`: port array
- `tracker`: chunkStateTracker
- `bkt`: LeakyBucket rate limiter

## 5. Group Builder (group_builder.go)

### cidrGroup Structure

```go
type cidrGroup struct {
    targets []scanTarget
}
```

### groupBuildStrategy Interface

```go
type groupBuildStrategy interface {
    ShouldInclude(rec input.CIDRRecord) bool
    Key(rec input.CIDRRecord) (string, error)
    NewGroup(rec input.CIDRRecord) (cidrGroup, error)
    MergeGroup(existing cidrGroup, rec input.CIDRRecord) (cidrGroup, error)
    RequireNonEmpty() bool
}
```

### Build Groups Function

```go
func buildGroups(records []input.CIDRRecord, strategy groupBuildStrategy) (map[string]cidrGroup, error)
```

Core logic:
1. Iterate through records
2. Filter via `ShouldInclude()`
3. Group by `Key()`
4. Create or merge via `NewGroup()`/`MergeGroup()`
5. Sort targets by IP within each group

### Basic Mode (basicGroupStrategy)

```go
func buildCIDRGroups(cidrRecords []input.CIDRRecord) (map[string]cidrGroup, error)
```

- Groups records by CIDR boundary (`ip_cidr`)
- Includes all valid records
- Expands IP selectors within each CIDR

### Rich Mode (richGroupStrategy)

```go
func buildRichGroups(cidrRecords []input.CIDRRecord) (map[string]cidrGroup, error)
```

- Groups by `dst_network_segment` (owner CIDR)
- Only includes `IsRich && IsValid` records
- Implements **CIDR-scoped rate control** with **global execution-key deduplication**
- Reassigns targets to their owner CIDR after initial grouping

### Rich Mode: Execution Key Ownership

```go
// Build owner map during group construction
ownerByExecutionKey := make(map[string]string)

// After initial grouping, reassign non-owner targets
for cidr, group := range groups {
    for _, target := range group.targets {
        ownerCIDR := ownerByExecutionKey[target.meta.executionKey]
        if ownerCIDR != cidr {
            // Move target to owner CIDR's group
            groups[ownerCIDR].targets = append(groups[ownerCIDR].targets, target)
        }
    }
}
```

### Reason-Aware Target Expansion

Rich mode supports intelligent target expansion based on the `reason` field:

```go
const (
    reasonPrecheckAllowAll  = "PRECHECK_ALLOW_ALL"
    reasonMatchPolicyAccept = "MATCH_POLICY_ACCEPT"
)

func richTargetIPs(rec input.CIDRRecord) ([]string, error) {
    switch strings.EqualFold(rec.Reason, reasonPrecheckAllowAll) {
    case true:
        // Expand entire dst_network_segment CIDR
        return task.ExpandIPSelectors([]string{rec.DstNetworkSegment})
    case strings.EqualFold(rec.Reason, reasonMatchPolicyAccept):
        // Single dst_ip only
        return []string{rec.DstIP}
    default:
        // Default: single dst_ip
        return []string{rec.DstIP}
    }
}
```

### Metadata Merging

When duplicate execution keys are encountered:

```go
func mergeFieldValue(existing, incoming string) string {
    // Merge with pipe separator: "value1|value2"
    // Deduplicates: "a|b" + "a" = "a|b"
}
```

Merges: `fabName`, `cidrName`, `serviceLabel`, `decision`, `policyID`, `reason`, `srcIP`, `srcNetworkSegment`

## 6. Task Dispatcher (task_dispatcher.go)

### dispatchTasks

```go
func dispatchTasks(
    ctx context.Context,
    runtimes []chunkRuntime,
    taskCh chan<- scanTask,
    policy dispatchPolicy,
    ctrl *speedctrl.Controller,
) error
```

**dispatchPolicy struct:**
```go
type dispatchPolicy struct {
    bucketRate     int
    bucketCapacity int
    delay          time.Duration
}
```

### Dispatch Loop

```go
for _, rt := range runtimes {
    // Chunk-serial: process each CIDR sequentially
    for i := rt.tracker.NextIndex; i < rt.tracker.TotalCount; i++ {
        // Acquire rate limit token
        if err := rt.bkt.Acquire(ctx); err != nil {
            return err
        }

        // Wait on pause gate
        <-ctrl.Gate()

        // Resolve target from index
        target := indexToRuntimeTarget(rt, i)

        // Create and send task
        task := scanTask{
            chunkIdx: idx,
            ipCidr:   rt.targets[i].ipCidr.String(),
            ip:       target.ip,
            port:     target.port,
            meta:     target.meta,
        }
        taskCh <- task

        // Update tracker
        rt.tracker.NextIndex = i + 1
    }
}
```

## 7. Executor (executor.go)

### startScanExecutor

```go
func startScanExecutor(
    workers int,
    timeout time.Duration,
    dial scanner.DialFunc,
    taskCh <-chan scanTask,
    resultCh chan<- scanResult,
) func() error
```

**Worker Pool:**
```go
for w := 0; w < workers; w++ {
    go func() {
        for task := range taskCh {
            // Perform TCP probe
            r := scanner.ScanTCP(dial, task.ip, task.port, timeout)
            
            // Convert to result
            result := scanResult{
                chunkIdx: task.chunkIdx,
                record: writer.Record{
                    IP:         task.ip,
                    IPCidr:     task.ipCidr,
                    Port:       task.port,
                    Status:     r.Status,
                    ResponseMS: r.ResponseTimeMS,
                    // ... rich metadata from task.meta
                },
            }
            resultCh <- result
        }
    }()
}
```

## 8. Result Aggregator (result_aggregator.go)

### applyScanResult

```go
func applyScanResult(
    runtimes []chunkRuntime,
    summary *scanSummary,
    result scanResult,
)
```

Updates:
- `runtimes[result.chunkIdx].tracker.ScannedCount++`
- `summary.{written, openCount, closeCount, timeoutCount}++`

### writeScanRecord

```go
func writeScanRecord(
    outputs *outputFiles,
    record writer.Record,
) error
```

Writes to both:
- All results writer (CSV)
- Open-only writer (filtered)

## 9. Resume Manager (resume_manager.go)

### persistResumeState

```go
func persistResumeState(
    runtimes []chunkRuntime,
    resumePath string,
    shouldSave bool,
) error
```

**Save conditions:**
- Incomplete (some tasks not done) AND
- (Error occurred OR shouldSaveOnDispatchErr)

**Saved state:**
```go
type ChunkState struct {
    CIDR         string `json:"cidr"`
    NextIndex    int    `json:"nextIndex"`
    ScannedCount int    `json:"scannedCount"`
    Status       string `json:"status"`
}
```

## 10. Output Files (output_files.go)

### resolveBatchOutputPaths

```go
func resolveBatchOutputPaths(outputDir string) (scanPath, openedPath string, err error)
```

**Naming format:**
- `scan_results-YYYYMMDDTHHMMSSZ.csv`
- `opened_results-YYYYMMDDTHHMMSSZ.csv`

**Collision handling:** Adds `-n` suffix for same-second runs

### openBatchOutputs

```go
func openBatchOutputs(scanPath, openedPath string) (*outputFiles, error)
```

Creates `.tmp` files:
- `scan_results-...tmp`
- `opened_results-...tmp`

### Finalize

```go
func (o *outputFiles) Finalize(success bool) error
```

- Success: rename `.tmp` → final names
- Failure: keep `.tmp` files (for debugging)

## 11. Runtime Types (runtime_types.go)

### scanTarget

```go
type scanTarget struct {
    ip       string
    ipCidr   *net.IPNet
    port     int
    meta     writer.Record  // Rich metadata
}
```

### scanTask

```go
type scanTask struct {
    chunkIdx int
    ipCidr   string
    ip       string
    port     int
    meta     writer.Record
}
```

### scanResult

```go
type scanResult struct {
    chunkIdx int
    record   writer.Record
}
```

### chunkRuntime

```go
type chunkRuntime struct {
    targets []scanTarget
    ports   []int
    tracker *chunkStateTracker
    bkt     *ratelimit.LeakyBucket
}
```

### chunkStateTracker

```go
type chunkStateTracker struct {
    NextIndex    int
    ScannedCount int
    TotalCount   int
    Status       string
}
```

## 12. Adding New Orchestration Features

### Adding New Dispatch Policy

1. Add field to `dispatchPolicy` struct
2. Update `buildRuntime()` to use new field
3. Update `dispatchTasks()` to apply policy

### Adding New Output Format

1. Add writer to `outputFiles` struct
2. Update `openBatchOutputs()` to create new writer
3. Update `writeScanRecord()` to write to new output
4. Update `Finalize()` for new file type

### Adding New Progress Tracking

1. Add field to `scanSummary` struct
2. Update `applyScanResult()` to track new metric
3. Update `emitScanResultEvents()` to output new metric

## 13. Implementation Files Reference

| File | Responsibility |
|------|----------------|
| `pkg/scanapp/scan.go` | Main orchestration entry |
| `pkg/scanapp/executor.go` | Worker pool execution |
| `pkg/scanapp/task_dispatcher.go` | Task dispatch |
| `pkg/scanapp/resume_manager.go` | Resume persistence |
| `pkg/scanapp/input_loader.go` | Input loading |
| `pkg/scanapp/runtime_types.go` | Data structures |
| `pkg/scanapp/runtime_builder.go` | Run plan |
| `pkg/scanapp/group_builder.go` | CIDR grouping |
| `pkg/scanapp/chunk_lifecycle.go` | Chunk lifecycle |
| `pkg/scanapp/result_aggregator.go` | Result processing |
| `pkg/scanapp/batch_output.go` | Batch paths |
| `pkg/scanapp/output_files.go` | File management |

## 14. Integration Points

- **CLI**: `scanapp.Run(ctx, cfg, stdout, stderr, opts)` called from `cmd/port-scan`
- **Config**: All settings from `config.Config`
- **Scanner**: Uses `scanner.ScanTCP()` via worker pool
- **Rate Limit**: Uses `ratelimit.LeakyBucket` per chunk
- **Speed Control**: Uses `speedctrl.Controller` pause gate
- **Writer**: Output via `writer.CSVWriter` and `writer.OpenOnlyWriter`
- **State**: Resume via `pkg/state` package
