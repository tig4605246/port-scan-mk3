# STRUCTURE.md - Directory Layout

## Directory Structure

```
port-scan-mk3/
├── cmd/
│   └── port-scan/           # CLI entry point
│       ├── main.go
│       └── command_handlers.go
│
├── pkg/                     # Reusable packages
│   ├── cli/                 # CLI utilities
│   ├── config/              # Flag parsing
│   ├── input/               # CSV parsing
│   ├── task/                # Task modeling
│   ├── scanner/             # TCP probes
│   ├── scanapp/             # Orchestration
│   ├── writer/              # CSV output
│   ├── ratelimit/           # Rate limiting
│   ├── speedctrl/           # Pause control
│   ├── validate/            # Validation
│   ├── state/               # Resume/signal
│   └── logx/                # Logging
│
├── tests/
│   └── integration/          # Integration tests
│       └── testdata/         # Test fixtures
│
├── e2e/                     # End-to-end tests
│   ├── mock-pressure-api/   # Mock service
│   └── report/              # Test reports
│
├── docs/
│   ├── specs/               # Implementation specs
│   └── plans/               # Feature plans
│
└── scripts/                 # Build scripts
```

## Key Locations

| Purpose | Location |
|---------|----------|
| CLI main | `cmd/port-scan/main.go` |
| Scan orchestration | `pkg/scanapp/scan.go` |
| Input parsing | `pkg/input/` |
| Task expansion | `pkg/task/` |
| TCP scanning | `pkg/scanner/scanner.go` |
| CSV writer | `pkg/writer/csv_writer.go` |

## Naming Conventions

- **Packages**: lowercase, single word preferred
- **Files**: snake_case for Go files
- **Types**: PascalCase
- **Functions**: PascalCase
- **Constants**: PascalCase or SCREAMING_SNAKE_CASE
