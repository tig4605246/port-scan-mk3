# ARCHITECTURE.md - System Design

## Architectural Pattern

- **Pattern**: Pipeline orchestration with worker pool
- **Style**: Library-first design following SOLID principles

## Layers

```
┌─────────────────────────────────────────┐
│  CLI Layer (cmd/port-scan)              │  User interface
├─────────────────────────────────────────┤
│  Config Layer (pkg/config)              │  Flag parsing
├─────────────────────────────────────────┤
│  Core Domain Layer                      │
│  - pkg/input: Parsing                  │
│  - pkg/task: Task modeling              │
│  - pkg/scanner: TCP probes              │
├─────────────────────────────────────────┤
│  Orchestration Layer (pkg/scanapp)      │  Pipeline control
├─────────────────────────────────────────┤
│  Control Layer                          │
│  - pkg/ratelimit: Rate limiting         │
│  - pkg/speedctrl: Pause control         │
├─────────────────────────────────────────┤
│  I/O Layer                              │
│  - pkg/writer: CSV output               │
│  - pkg/state: Resume/signal             │
└─────────────────────────────────────────┘
```

## Data Flow

1. **Load**: CLI → Config → Input files
2. **Plan**: Build chunks, runtimes
3. **Dispatch**: Rate-limited task dispatch
4. **Execute**: Worker pool performs TCP probes
5. **Aggregate**: Results → CSV output
6. **Resume**: Save state on interrupt

## Key Abstractions

| Abstraction | Package | Purpose |
|-------------|---------|---------|
| `DialFunc` | scanner | TCP dial interface |
| `PressureFetcher` | scanapp | Pressure API interface |
| `groupBuildStrategy` | scanapp | Grouping interface |
| `LeakyBucket` | ratelimit | Rate limiter interface |

## Entry Points

- `cmd/port-scan/main.go`: CLI entry
- `scanapp.Run()`: Scan orchestration entry
