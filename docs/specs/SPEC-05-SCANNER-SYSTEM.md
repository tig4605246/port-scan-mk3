# SPEC-05: Scanner System Specification

## Overview

```
pkg/scanner/
├── scanner.go         # Core TCP probe implementation
├── scanner_test.go   # Unit tests
└── scanner_extra_test.go  # Edge case tests
```

## 1. Core Function

### ScanTCP

```go
func ScanTCP(
    dial func(context.Context, string, string) (net.Conn, error),
    ip string,
    port int,
    timeout time.Duration,
) Result
```

**Parameters:**
| Parameter | Type | Description |
|------------|------|-------------|
| `dial` | `func(context.Context, string, string) (net.Conn, error)` | Dial function (typically `(&net.Dialer{}).DialContext`) |
| `ip` | `string` | Target IP address |
| `port` | `int` | Target port |
| `timeout` | `time.Duration` | Connection timeout |

**Returns:** `Result` struct

## 2. Result Structure

```go
type Result struct {
    IP             string  // Input IP
    Port           int     // Input port
    Status         string  // "open" | "close" | "close(timeout)"
    ResponseTimeMS int64   // Elapsed time in milliseconds (only for "open")
    Error          string  // Error message (only for non-"open")
}
```

### Status Values

| Status | Description |
|--------|-------------|
| `"open"` | TCP connection successful |
| `"close"` | TCP connection refused/reset |
| `"close(timeout)"` | Connection timed out |

## 3. DialFunc Interface

### Definition (from pkg/scanapp/scan.go)

```go
type DialFunc func(context.Context, string, string) (net.Conn, error)
```

### Signature Breakdown

```
(context, network, address)
   ↓        ↓         ↓
 ctx   "tcp"    "10.0.0.1:443"
```

### Common Implementations

**Standard dialer:**
```go
dial := (&net.Dialer{
    Timeout:   timeout,
    LocalAddr: localAddr,  // Optional: bind to specific local port
}).DialContext
```

**Custom dialer (for testing):**
```go
mockDial := func(ctx context.Context, network, address string) (net.Conn, error) {
    // Return mock connection or error
}
```

## 4. Implementation Details

### Connection Flow

```
1. Construct address: net.JoinHostPort(ip, port)
2. Create timeout context: context.WithTimeout(ctx, timeout)
3. Call dial function: dial(ctx, "tcp", address)
4. On success:
   - Close connection immediately
   - Record response time
   - Return Status="open"
5. On error:
   - Classify error type
   - Return appropriate status
```

### Timeout Classification

```go
// Check if error is a timeout
var netErr net.Error
if errors.As(err, &netErr) && netErr.Timeout() {
    return Result{Status: "close(timeout)", Error: err.Error()}
}

// Check for context deadline exceeded
if errors.Is(err, context.DeadlineExceeded) {
    return Result{Status: "close(timeout)", Error: err.Error()}
}

// All other errors = connection refused/reset
return Result{Status: "close", Error: err.Error()}
```

### Response Time Measurement

```go
start := time.Now()
conn, err := dial(ctx, "tcp", address)
elapsed := time.Since(start)

if err != nil {
    // Handle error
}

conn.Close()
return Result{
    Status:         "open",
    ResponseTimeMS: elapsed.Milliseconds(),
}
```

## 5. Usage Patterns

### Basic Usage

```go
result := scanner.ScanTCP(
    (&net.Dialer{}).DialContext,
    "10.0.0.1",
    443,
    100*time.Millisecond,
)

if result.Status == "open" {
    fmt.Printf("Port %d is open (response time: %dms)\n", 
        result.Port, result.ResponseTimeMS)
}
```

### With Custom Timeout

```go
cfg.Timeout = 500 * time.Millisecond
result := scanner.ScanTCP(dial, ip, port, cfg.Timeout)
```

### With Local Address Binding

```go
dialer := &net.Dialer{
    Timeout: timeout,
    LocalAddr: &net.TCPAddr{
        IP:   net.ParseIP("0.0.0.0"),
        Port: 0, // Let OS choose port
    },
}
result := scanner.ScanTCP(dialer.DialContext, ip, port, timeout)
```

### Testing with Mock

```go
mockConn := &mockConn{...}
mockDial := func(ctx context.Context, network, address string) (net.Conn, error) {
    return mockConn, nil
}

result := scanner.ScanTCP(mockDial, "10.0.0.1", 443, timeout)
// result.Status will be "open"
```

## 6. Design Decisions

| Decision | Rationale |
|----------|-----------|
| Injectable dial function | Enables unit testing with mock dialers |
| Context-based timeout | Standard Go pattern for cancellation/timeout |
| Immediate connection close | Stateless port scanning - no connection reuse |
| No connection pooling | Designed for high-volume scanning |
| Minimal footprint | Single probe = single function call |

## 7. Configuration

### Timeout Recommendations

| Network Type | Recommended Timeout |
|--------------|-------------------|
| Local LAN | 100ms - 500ms |
| Data center | 100ms - 200ms |
| Internet | 1s - 5s |

### Worker Pool Configuration

Workers consume `ScanTCP` results. Configuration in `config.Config`:
- `Workers`: Number of concurrent workers
- `Timeout`: Per-probe timeout
- `BucketRate`: Rate limit tokens/second
- `BucketCapacity`: Burst allowance

## 8. Extending the Scanner

### Adding New Protocols

1. Create new function (e.g., `ScanUDP`)
2. Follow same signature pattern
3. Add to `DialFunc` interface if needed

### Adding TLS Support

```go
tlsDial := func(ctx context.Context, network, address string) (net.Conn, error) {
    return tls.Dial("tcp", address, &tls.Config{...})
}
```

### Adding Custom Error Classification

Modify timeout classification logic in `scanner.go`:

```go
// Add custom error handling
if errors.Is(err, someCustomError) {
    return Result{Status: "close(special)", Error: err.Error()}
}
```

## 9. Implementation Files Reference

| File | Responsibility |
|------|----------------|
| `pkg/scanner/scanner.go` | Core TCP probe implementation |
| `pkg/scanner/scanner_test.go` | Basic unit tests |
| `pkg/scanner/scanner_extra_test.go` | Edge case tests (timeout, refused) |
| `pkg/scanapp/scan.go` | DialFunc interface definition |
| `pkg/scanapp/executor.go` | Worker pool consuming ScanTCP |

## 10. Integration Points

- **Executor**: `startScanExecutor()` in `executor.go` calls `ScanTCP`
- **Config**: Timeout from `config.Config.Timeout`
- **Runtime**: Dial function passed via `RunOptions.Dial`
