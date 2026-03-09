# Port Scan MK3

Developer-first TCP port scanner CLI in Go with fail-fast input validation, pressure-aware pacing, resumable scanning, and e2e verification.

## How It Works

`port-scan-mk3` follows a deterministic pipeline:

1. Parse CLI flags and validate required inputs.
2. Load CIDR CSV and port list, then apply fail-fast validation (format, overlap, containment).
3. Expand `ip` selectors to concrete IPv4 targets grouped by `ip_cidr`.
4. Build scan tasks (`target x port`) and dispatch through pause/rate control.
5. Run TCP probes and emit structured runtime events.
6. Write timestamped batch output files:
   - `scan_results-YYYYMMDDTHHMMSSZ[-n].csv`
   - `opened_results-YYYYMMDDTHHMMSSZ[-n].csv`
7. Save resume state on cancellation/failure so the next run can continue.

## Commands at a Glance

- `validate`: Validate CIDR/port inputs only (no network scan).
- `scan`: Execute full scan pipeline with optional pressure control and resume.

Quick usage:

```bash
go run ./cmd/port-scan validate -cidr-file <cidr.csv> -port-file <ports.csv> -format human
go run ./cmd/port-scan scan -cidr-file <cidr.csv> -port-file <ports.csv>
```

## Flags Quick Reference

This section lists high-impact flags. Full definitions are in [All flags](docs/cli/flags.md).

| Flag | Typical Use |
|------|-------------|
| `-cidr-file` | CIDR input CSV path (required) |
| `-port-file` | Port list path (required) |
| `-cidr-ip-col` / `-cidr-ip-cidr-col` | Map custom CSV column names |
| `-output` | Choose output directory and filename prefix anchor |
| `-resume` | Read/write state from explicit path |
| `-disable-api` | Disable pressure API polling |
| `-pressure-api` / `-pressure-interval` | Configure pressure-based pause control |
| `-workers` / `-timeout` / `-delay` | Tune concurrency and probe pacing |
| `-log-level` / `-format` | Runtime visibility (`human` or `json`) |

## Scenario Cookbook

Use the scenario guide for copy-paste commands and expected behavior:

- [Scenario cookbook](docs/cli/scenarios.md)

## E2E Overview

E2E uses Docker Compose and mock services to verify real scan behavior and artifact outputs.

- Coverage details and topology: [E2E overview](docs/e2e/overview.md)
- Run e2e: `bash e2e/run_e2e.sh`
- Expected artifacts: `e2e/out/report.html`, `e2e/out/report.txt`, batch CSVs, and resume-state snapshots.

## Development Guardrails

- Keep reusable scanning, parsing, orchestration, and writer logic in `pkg/`; keep
  `cmd/port-scan` limited to CLI wiring and user-facing I/O.
- Preserve SOLID boundaries when adding or changing packages: narrow responsibilities,
  consumer-owned interfaces, and composition over god objects.
- Validate changes with `go test ./...`, `bash scripts/coverage_gate.sh`, and
  `bash e2e/run_e2e.sh` when pipeline behavior changes.

## Refactor Workflow

When changing orchestration or CLI behavior:

1. Add or extend the failing test first for the exact contract you are about to protect.
2. Keep `cmd/port-scan/main.go` as the composition root only; route behavior into focused
   handlers or package collaborators.
3. Keep `pkg/scanapp/scan.go` as orchestration glue; extract input loading, runtime building,
   dispatch, monitoring, aggregation, and resume logic into narrowly scoped files instead of
   regrowing a monolith.
4. Preserve `pkg/cli/output.go` payload shapes and scan artifact naming unless the spec and
   release notes explicitly call out a contract change.
5. Record the actual red/green/refactor commands in
   `specs/001-solid-refactor-tdd/quickstart.md` before considering the work complete.

## Architecture Diagram

Static HTML + CSS architecture diagram:

- [Architecture diagram](docs/architecture/diagram.html)

## Where to Go Deeper

- [All flags](docs/cli/flags.md)
- [Scenario cookbook](docs/cli/scenarios.md)
- [E2E overview](docs/e2e/overview.md)
- [Architecture diagram](docs/architecture/diagram.html)

## Verification Commands

- Unit + integration: `go test ./...`
- Coverage gate: `bash scripts/coverage_gate.sh`
- E2E gate: `bash e2e/run_e2e.sh`
