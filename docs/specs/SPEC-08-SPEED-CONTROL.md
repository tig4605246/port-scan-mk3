# SPEC-08: Speed Control System Specification

## Overview

```
pkg/speedctrl/
├── controller.go       # Core pause controller
├── controller_test.go  # Unit tests
├── keyboard.go         # Manual pause (spacebar)
└── extra_test.go       # Additional tests

pkg/scanapp/
├── pressure.go              # PressureFetcher interface
├── pressure_monitor.go      # API polling
└── pressure_test.go         # Tests
```

## 1. Controller (pkg/speedctrl/controller.go)

### Controller Struct

```go
type Controller struct {
    mu           sync.RWMutex
    apiPaused    bool        // Pressure-based auto-pause
    manualPaused bool       // Keyboard/user-initiated pause
    gate         chan struct{}  // Blocking channel for pause
}
```

### Constructor

```go
func NewController(opts ...Option) *Controller
```

Options:
```go
type Option func(*Controller)

func WithInitialGate(closed bool) Option  // Start with closed gate
```

### Key Methods

| Method | Description |
|--------|-------------|
| `SetAPIPaused(v bool)` | Set pressure-based pause state |
| `SetManualPaused(v bool)` | Set keyboard/manual pause state |
| `ToggleManualPaused()` | Toggle manual pause, return new state |
| `ManualPaused() bool` | Query manual pause state |
| `IsPaused() bool` | Returns true if ANY pause active |
| `Gate() <-chan struct{}` | Returns pause gate channel |

### Gate Logic

```
┌─────────────────────────────────────────┐
│           recomputeLocked()              │
├─────────────────────────────────────────┤
│ if apiPaused OR manualPaused:           │
│    gate = make(chan struct{})  [BLOCK]  │
│ else:                                   │
│    close(gate)                [PASS]   │
└─────────────────────────────────────────┘
```

- **Paused (gate empty)**: `<-gate` blocks → dispatch waits
- **Not paused (gate closed)**: `<-gate` returns immediately → dispatch proceeds

## 2. Pressure-Based Pause (pkg/scanapp/pressure_monitor.go)

### pollPressureAPI

```go
func pollPressureAPI(
    ctx context.Context,
    cfg config.Config,
    opts RunOptions,
    ctrl *speedctrl.Controller,
    logger *logx.Logger,
    errCh chan<- error,
)
```

**Flow:**
1. Get interval from config (default 5s)
2. Get threshold from options (default 90)
3. Loop until context cancelled:
   - Call `fetcher.Fetch(ctx)` to get pressure
   - Compare against threshold:
     - `pressure >= threshold` → `ctrl.SetAPIPaused(true)`
     - `pressure < threshold` → `ctrl.SetAPIPaused(false)`
   - Log state changes
   - Sleep for interval

**Failure Handling:**
- Allows 2 consecutive failures (logs warnings)
- On 3rd failure: sends error to `errCh` and terminates

### Threshold

| Threshold | Behavior |
|-----------|----------|
| >= 90 | Pause scanning |
| < 90 | Resume scanning |

Default threshold is hardcoded to 90 (not exposed as CLI flag).

## 3. Pressure Fetcher Interface (pkg/scanapp/pressure.go)

### PressureFetcher Interface

```go
type PressureFetcher interface {
    Fetch(ctx context.Context) (int, error)
}
```

### SimplePressureFetcher

```go
type SimplePressureFetcher struct {
    URL string
}

func (f *SimplePressureFetcher) Fetch(ctx context.Context) (int, error)
```

**Expected Response Format:**
```json
{"pressure": 85}
```

**Flexible parsing:** Handles int, float64, and string JSON values.

### AuthenticatedPressureFetcher

```go
type AuthenticatedPressureFetcher struct {
    AuthURL string  // Token endpoint
    DataURL string  // Data endpoint
}

func (f *AuthenticatedPressureFetcher) Fetch(ctx context.Context) (int, error)
```

**Flow:**
1. GET to `AuthURL` → receive token
2. GET to `DataURL` with `Authorization: Bearer <token>`
3. Parse pressure from response

## 4. Manual Pause (pkg/speedctrl/keyboard.go)

### StartKeyboardLoop

```go
func StartKeyboardLoop(ctx context.Context, ctrl *Controller) error
```

**Behavior:**
- Only activates if `term.IsTerminal` returns true
- Puts terminal in raw mode (immediate key capture)
- Listens for keypresses
- On space key (ASCII 32): calls `ctrl.ToggleManualPaused()`
- Restores terminal on exit

### Manual Pause Monitor

In `pressure_monitor.go`:
```go
func startManualPauseMonitor(ctx context.Context, ctrl *Controller, logger *logx.Logger)
```

- Polls `ctrl.ManualPaused()` every 200ms
- Logs state changes

## 5. Integration in Scan Pipeline

### Creation

In `scanapp/scan.go`:
```go
ctrl := speedctrl.NewController()
```

### Goroutine Startup

```go
// Keyboard pause
go func() {
    if err := speedctrl.StartKeyboardLoop(ctx, ctrl); err != nil {
        // Ignore (not a terminal)
    }
}()

// Manual pause logging
go startManualPauseMonitor(ctx, ctrl, logger)

// Pressure API polling
go pollPressureAPI(ctx, cfg, opts, ctrl, logger, errCh)
```

### Gate Usage

In `task_dispatcher.go`:
```go
func dispatchTasks(..., ctrl *speedctrl.Controller, ...) {
    for ... {
        // Wait on pause gate
        <-ctrl.Gate()
        // ... dispatch task
    }
}
```

## 6. CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-pressure-api` | "http://localhost:8080/api/pressure" | Pressure API URL |
| `-pressure-interval` | 5s | Polling interval |
| `-disable-api` | false | Disable pressure API |

### Example

```bash
# With pressure API
go run ./cmd/port-scan scan -cidr-file ips.csv -pressure-api http://api.example.com/pressure

# Without pressure API
go run ./cmd/port-scan scan -cidr-file ips.csv -disable-api

# Custom interval
go run ./cmd/port-scan scan -cidr-file ips.csv -pressure-interval 2s
```

## 7. Design Decisions

| Decision | Rationale |
|----------|-----------|
| Two pause sources | API-driven (automatic) AND manual (user control) |
| Channel-based gate | Simple, idiomatic blocking mechanism |
| Hardcoded threshold | 90 is a common default, not meant to be changed frequently |
| 2 failure tolerance | Allows for transient API failures |
| Terminal detection | Only enable keyboard when actually in terminal |

## 8. Adding New Pause Sources

### Adding New Pause Mechanism

1. Add field to `Controller`:
```go
type Controller struct {
    mu           sync.RWMutex
    apiPaused    bool
    manualPaused bool
    newPaused    bool  // Add new source
    gate         chan struct{}
}
```

2. Update `recomputeLocked()`:
```go
func (c *Controller) recomputeLocked() {
    paused := c.apiPaused || c.manualPaused || c.newPaused
    // ... gate logic
}
```

3. Add setter method:
```go
func (c *Controller) SetNewPaused(v bool) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.newPaused = v
    c.recomputeLocked()
}
```

## 9. Implementation Files Reference

| File | Responsibility |
|------|----------------|
| `pkg/speedctrl/controller.go` | Core pause controller |
| `pkg/speedctrl/keyboard.go` | Manual pause (spacebar) |
| `pkg/speedctrl/controller_test.go` | Unit tests |
| `pkg/scanapp/pressure.go` | PressureFetcher interface |
| `pkg/scanapp/pressure_monitor.go` | API polling |
| `pkg/scanapp/scan.go` | Integration |

## 10. Integration Points

- **Config**: API URL, interval from `config.Config`
- **Dispatcher**: Gate controls task dispatch
- **Logger**: State changes logged via `logx.Logger`
- **Options**: Custom `PressureFetcher` via `RunOptions`
