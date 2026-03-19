# STACK.md - Technology Stack

## Languages & Runtime

- **Language**: Go 1.24.x
- **Toolchain**: go1.24.4
- **Platform**: Linux

## Core Dependencies

| Package | Purpose | Source |
|---------|---------|--------|
| `golang.org/x/sys` | System calls | Standard library extended |
| `golang.org/x/term` | Terminal I/O | Standard library extended |

## Internal Packages

| Package | Purpose |
|---------|---------|
| `pkg/input` | CIDR/Rich CSV parsing |
| `pkg/task` | Task modeling, IP selector expansion |
| `pkg/scanner` | TCP probe primitive |
| `pkg/scanapp` | Scan orchestration |
| `pkg/writer` | CSV output writer |
| `pkg/ratelimit` | Leaky bucket rate limiting |
| `pkg/speedctrl` | Pressure-based pause control |
| `pkg/validate` | Input validation |
| `pkg/state` | Resume/signal handling |
| `pkg/logx` | Logging utilities |
| `pkg/config` | CLI flag parsing |
| `pkg/cli` | CLI utilities |

## Configuration

- **Go Modules**: `go.mod`
- **Testing**: `go test` with coverage
- **Code Quality**: golint, go vet

## Build Commands

```bash
go build ./...
go test ./...
go test -coverprofile=coverage.out ./...
```

## Project Type

- **CLI Tool**: TCP port scanner
- **Architecture**: Library-first design (packages in `pkg/`, CLI in `cmd/`)
