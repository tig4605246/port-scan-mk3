# SPEC-04: Task System Specification

## Overview

```
pkg/task/
├── types.go            # Chunk and Task data structures
├── selector_expand.go  # IP selector expansion logic
├── execution_key.go    # Execution key builder
├── index.go           # Index-to-target conversion
└── ipv4.go            # IPv4 utilities
```

## 1. Core Data Structures

### 1.1 Chunk

```go
type Chunk struct {
    CIDR     string     // CIDR boundary (e.g., "10.0.0.0/24")
    Ports    []int      // Ports to scan for this CIDR
    Targets  []string   // Expanded IP selectors
    
    // Progress tracking
    NextIndex   int     // Next task index to dispatch (0-based)
    ScannedCount int   // Number of completed scans
    TotalCount  int     // Total tasks (len(Targets) * len(Ports))
    Status      string  // "pending", "scanning", "completed"
}
```

### 1.2 Task

```go
type Task struct {
    ChunkCIDR string    // Parent CIDR
    IP        string    // Target IP
    Port      int       // Target port
    Index     int       // Flat index within chunk
}
```

## 2. Selector Expansion

### ExpandIPSelectors

```go
func ExpandIPSelectors(selectors []string) ([]string, error)
```

**Input:** Raw selectors (single IPs or CIDR blocks)
- Single IP: `10.0.0.1`
- CIDR block: `10.0.0.8/30`

**Process:**
1. Parse each selector - distinguish IP vs CIDR via `net.ParseIP()` vs `net.ParseCIDR()`
2. For IPs: convert to uint32 via BigEndian binary
3. For CIDRs: compute start/end range via `ipv4Range()`, enumerate all hosts
4. Deduplicate via `map[uint32]struct{}`
5. Sort ascending, convert back to strings

**Output:** Sorted, deduplicated list of IPv4 strings

**Validation:** Rejects non-IPv4 (IPv6) with error

**Example:**
```go
selectors := []string{"10.0.0.1", "10.0.0.8/30"}
expanded, err := ExpandIPSelectors(selectors)
// Returns: ["10.0.0.1", "10.0.0.8", "10.0.0.9", "10.0.0.10", "10.0.0.11"]
```

## 3. Execution Key

### BuildExecutionKey

```go
func BuildExecutionKey(dstIP string, port int, protocol string) (string, error)
```

**Format:** `dst_ip:port/protocol`

**Example:**
```go
key, _ := BuildExecutionKey("10.0.0.1", 443, "tcp")
// Returns: "10.0.0.1:443/tcp"
```

**Validation:**
| Rule | Error |
|------|-------|
| IPv4 address required | "invalid IP address" |
| Port 1-65535 | "invalid port" |
| Protocol must be "tcp" | "protocol must be tcp" |

## 4. Index-to-Target Mapping

### IndexToTarget

```go
func IndexToTarget(idx int, ips []string, ports []int) (string, int)
```

**Formula:**
```go
ipIdx := idx / len(ports)
portIdx := idx % len(ports)
return ips[ipIdx], ports[portIdx]
```

**Example:**
```go
ips := []string{"10.0.0.1", "10.0.0.2"}
ports := []int{80, 443}

IndexToTarget(0, ips, ports) // "10.0.0.1", 80
IndexToTarget(1, ips, ports) // "10.0.0.1", 443
IndexToTarget(2, ips, ports) // "10.0.0.2", 80
IndexToTarget(3, ips, ports) // "10.0.0.2", 443
```

### IndexToIPv4Target

```go
func IndexToIPv4Target(ipNet *net.IPNet, ports []int, idx int) (string, int, error)
```

Direct index-to-target without pre-expanded IP array.

### CountIPv4Hosts

```go
func CountIPv4Hosts(ipNet *net.IPNet) (int, error)
```

Counts hosts in CIDR (excludes network and broadcast for /24 and smaller).

## 5. Task Lifecycle

```
Input Load (pkg/input)
       ↓
Chunk Build (pkg/scanapp/chunk_lifecycle.go)
       ↓
Runtime Build (buildRuntime)
       ↓
Dispatch (pkg/scanapp/task_dispatcher.go)
       ↓
Execute (pkg/scanapp/executor.go)
       ↓
Resume (persistResumeState)
```

### Status Flow

```
pending → scanning → completed
  ↑          ↓
  └──────────┘ (if NextIndex >= TotalCount)
```

## 6. Runtime Types (pkg/scanapp/runtime_types.go)

These types extend the core task types for orchestration:

### scanTarget

```go
type scanTarget struct {
    ip       string
    ipCidr   *net.IPNet
    port     int
    meta     writer.Record  // Rich metadata for output
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

### chunkRuntime

```go
type chunkRuntime struct {
    targets []scanTarget
    ports   []int
    tracker *chunkStateTracker
    bkt     *ratelimit.LeakyBucket
}
```

## 7. Adding New Task Features

### Adding New Task Metadata

1. Add field to `scanTarget` in `runtime_types.go`
2. Update `indexToRuntimeTarget()` in `task_dispatcher.go`
3. Pass through to output in `result_aggregator.go`

### Adding New Selector Syntax

1. Modify `ExpandIPSelectors()` in `selector_expand.go`
2. Add parser for new syntax
3. Update validation rules
4. Add tests

### Adding New Grouping Strategy

1. Create new strategy interface in `group_builder.go`
2. Implement grouping logic
3. Register in `groupBuildStrategy` interface

## 8. Implementation Files Reference

| File | Responsibility |
|------|----------------|
| `pkg/task/types.go` | Chunk, Task data structures |
| `pkg/task/selector_expand.go` | IP selector expansion |
| `pkg/task/execution_key.go` | Execution key building |
| `pkg/task/index.go` | Index-to-target conversion |
| `pkg/task/ipv4.go` | IPv4 utilities |
| `pkg/scanapp/runtime_types.go` | Runtime extensions (scanTarget, scanTask, chunkRuntime) |
| `pkg/scanapp/group_builder.go` | CIDR grouping |
| `pkg/scanapp/chunk_lifecycle.go` | Chunk lifecycle management |

## 9. Integration Points

- **Input**: `input.CIDRRecord` provides IP/CIDR for expansion
- **Rate Limit**: Each chunk gets its own `LeakyBucket`
- **Dispatcher**: Uses `IndexToTarget()` to resolve flat index to (IP, port)
- **Resume**: `Chunk` state (NextIndex, ScannedCount, Status) persisted
- **Writer**: Execution key included in output CSV
